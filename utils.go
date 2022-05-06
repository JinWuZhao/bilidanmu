package main

import (
	"bytes"
	"compress/zlib"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
)

func GetRealRoomID(short int) (uint32, error) {
	url := fmt.Sprintf("%s?id=%d", RealID, short)
	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("http.Get() error: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	rawData, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("io.ReadAll() error: %w", err)
	}

	return json.Get(rawData, "data", "room_id").ToUint32(), nil
}

// GetToken return the necessary token for connecting to the server
func GetToken(roomId uint32) (string, error) {
	url := fmt.Sprintf("%s?room_id=%d&platform=pc&player=web", keyUrl, roomId)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("http.Get() error: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	rawData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("io.ReadAll error: %w", err)
	}
	return json.Get(rawData, "data").Get("token").ToString(), nil
}

func GetRoomInfo(roomId uint32) (*RoomInfo, error) {
	url := fmt.Sprintf("%s?room_id=%d", roomInfoUrl, roomId)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http.Get() error: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	rawData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll() error: %w", err)
	}

	return &RoomInfo{
		RoomId:     roomId,
		UpUid:      json.Get(rawData, "data").Get("room_info").Get("uid").ToUint32(),
		Title:      json.Get(rawData, "data").Get("room_info").Get("title").ToString(),
		Tags:       json.Get(rawData, "data").Get("room_info").Get("tags").ToString(),
		LiveStatus: json.Get(rawData, "data").Get("room_info").Get("live_status").ToBool(),
		LockStatus: json.Get(rawData, "data").Get("room_info").Get("lock_status").ToBool(),
	}, nil
}

func ZlibInflate(compress []byte) ([]byte, error) {
	var out bytes.Buffer
	c := bytes.NewReader(compress)
	r, err := zlib.NewReader(c)
	if err != zlib.ErrChecksum && err != zlib.ErrDictionary && err != zlib.ErrHeader && r != nil {
		_, _ = io.Copy(&out, r)
		if err := r.Close(); err != nil {
			log.Println("r.close err:", err)
			return nil, err
		}
		return out.Bytes(), nil
	}
	return nil, err
}

func ByteArrToDecimal(src []byte) (sum int) {
	if src == nil {
		return 0
	}
	b := []byte(hex.EncodeToString(src))
	l := len(b)
	for i := l - 1; i >= 0; i-- {
		base := int(math.Pow(16, float64(l-i-1)))
		var mul int
		if int(b[i]) >= 97 {
			mul = int(b[i]) - 87
		} else {
			mul = int(b[i]) - 48
		}

		sum += base * mul
	}
	return
}
