package bilidanmu

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"net/url"
	"time"

	jsoniter "github.com/json-iterator/go"
	"nhooyr.io/websocket"
)

const (
	RealID      = "http://api.live.bilibili.com/room/v1/Room/room_init" // params: id=xxx
	DanMuServer = "hw-bj-live-comet-06.chat.bilibili.com:443"
	keyUrl      = "https://api.live.bilibili.com/room/v1/Danmu/getConf"                 // params: room_id=xxx&platform=pc&player=web
	roomInfoUrl = "https://api.live.bilibili.com/xlive/web-room/v1/index/getInfoByRoom" // params: room_id=xxx
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type Client struct {
	Room      *RoomInfo    `json:"room"`
	Request   *RequestInfo `json:"request"`
	conn      *websocket.Conn
	Connected bool `json:"connected"`
	stop      chan error
}

type RoomInfo struct {
	RoomId     uint32 `json:"room_id"`
	UpUid      uint32 `json:"up_uid"`
	Title      string `json:"title"`
	Online     uint32 `json:"online"`
	Tags       string `json:"tags"`
	LiveStatus bool   `json:"live_status"`
	LockStatus bool   `json:"lock_status"`
}

type RequestInfo struct {
	Uid       uint8  `json:"uid"`
	RoomId    uint32 `json:"roomid"`
	ProtoVer  uint8  `json:"protover"`
	Platform  string `json:"platform"`
	ClientVer string `json:"clientver"`
	Type      uint8  `json:"type"`
	Key       string `json:"key"`
}

func NewRequestInfo(roomId uint32) (*RequestInfo, error) {
	t, err := GetToken(roomId)
	if err != nil {
		return nil, fmt.Errorf("GetToken() error: %w", err)
	}
	return &RequestInfo{
		Uid:       0,
		RoomId:    roomId,
		ProtoVer:  2,
		Platform:  "web",
		ClientVer: "1.10.2",
		Type:      2,
		Key:       t,
	}, nil
}

func NewClient(roomId uint32) (*Client, error) {
	request, err := NewRequestInfo(roomId)
	if err != nil {
		return nil, fmt.Errorf("NewRequestInfo() error: %w", err)
	}
	room, err := GetRoomInfo(roomId)
	if err != nil {
		return nil, fmt.Errorf("GetRoomInfo() error: %w", err)
	}
	return &Client{
		Room:      room,
		Request:   request,
		conn:      nil,
		Connected: false,
		stop:      make(chan error, 1),
	}, nil
}

func (c *Client) Start(ctx context.Context, receiver func(Message)) error {
	u := url.URL{Scheme: "wss", Host: DanMuServer, Path: "/sub"}
	conn, _, err := websocket.Dial(ctx, u.String(), nil)
	if err != nil {
		return fmt.Errorf("websocket.Dial() error: %w", err)
	}
	c.conn = conn

	log.Println("当前直播间状态：", c.Room.LiveStatus)
	log.Println("连接弹幕服务器 ", DanMuServer, " 成功，正在发送握手包...")

	r, err := json.Marshal(c.Request)
	if err != nil {
		return fmt.Errorf("json.Marshal() error: %w", err)
	}
	err = c.SendPackage(ctx, 0, 16, 1, 7, 1, r)
	if err != nil {
		return fmt.Errorf("c.SendPackage() error: %w", err)
	}
	go c.ReceiveMsg(ctx, receiver)
	go c.HeartBeat(ctx)
	return nil
}

func (c *Client) SendPackage(ctx context.Context, packetLen uint32, magic uint16, ver uint16, typeID uint32, param uint32, data []byte) error {
	packetHead := new(bytes.Buffer)

	if packetLen == 0 {
		packetLen = uint32(len(data) + 16)
	}
	var pData = []interface{}{
		packetLen,
		magic,
		ver,
		typeID,
		param,
	}

	// 将包的头部信息以大端序方式写入字节数组
	for _, v := range pData {
		err := binary.Write(packetHead, binary.BigEndian, v)
		if err != nil {
			return fmt.Errorf("binary.Write() error: %w", err)
		}
	}

	// 将包内数据部分追加到数据包内
	sendData := append(packetHead.Bytes(), data...)

	err := c.conn.Write(ctx, websocket.MessageBinary, sendData)
	if err != nil {
		return fmt.Errorf("c.conn.Write() error: %w", err)
	}

	return nil
}

func (c *Client) ReceiveMsg(ctx context.Context, receiver func(Message)) {
loop:
	for {
		select {
		case <-ctx.Done():
			c.stop <- nil
			break loop
		default:
		}

		_, msg, err := c.conn.Read(ctx)
		if err != nil {
			log.Println("ReceiveMsg(): c.conn.Read() error:", err)
			time.Sleep(time.Second)
			continue
		}

		switch msg[11] {
		case 8:
			log.Println("握手包收发完毕，连接成功")
			c.Connected = true
		case 3:
			onlineNow := ByteArrToDecimal(msg[16:])
			if uint32(onlineNow) != c.Room.Online {
				c.Room.Online = uint32(onlineNow)
				log.Println("当前房间人气变动：", uint32(onlineNow))
			}
		case 5:
			inflated, err := ZlibInflate(msg[16:])
			if err == nil {
				for len(inflated) > 0 {
					l := ByteArrToDecimal(inflated[:4])
					c := CMD(json.Get(inflated[16:l], "cmd").ToString())
					switch c {
					case CMDDanmuMsg:
						m := NewDanMuMsg()
						m.Decode(inflated[16:l])
						receiver(m)
					case CMDSendGift:
						g := NewGift()
						g.Decode(inflated[16:l])
						receiver(g)
					case CMDWelcomeVip:
						u := NewWelcomeVip()
						u.Decode(inflated[16:l])
						receiver(u)
					case CMDWelcomeGuard:
						u := NewWelcomeGuard()
						u.Decode(inflated[16:l])
						receiver(u)
					case CMDEntry:
						u := NewWelcomeEntry()
						u.Decode(inflated[16:l])
						receiver(u)
					}
					inflated = inflated[l:]
				}
			}
		}
	}
}

func (c *Client) HeartBeat(ctx context.Context) {
loop:
	for {
		select {
		case <-ctx.Done():
			c.stop <- nil
			break loop
		default:
		}

		if c.Connected {
			obj := []byte("5b6f626a656374204f626a6563745d")
			err := c.SendPackage(ctx, 0, 16, 1, 2, 1, obj)
			if err != nil {
				log.Println("HeartBeat(): c.SendPackage() error: ", err)
				time.Sleep(time.Second * 30)
				continue
			}
			time.Sleep(30 * time.Second)
		}
	}
}

func (c *Client) WaitForStop() error {
	err := <-c.stop
	if err != nil {
		return fmt.Errorf("client stopped with error: %w", err)
	}
	return nil
}
