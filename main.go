package main

import (
	"context"
	"fmt"
	"log"
	"os"
)

func main() {
	var roomId uint32
	log.Print("请输入房间号，长短 ID 均可：")
	_, err := fmt.Scanf("%d", &roomId)
	if err != nil {
		log.Println("房间号输入错误，请退出重新输入！")
		os.Exit(0)
	}

	// 兼容房间号短 ID
	if roomId >= 100 && roomId < 1000 {
		roomId, err = GetRealRoomID(int(roomId))
		if err != nil {
			log.Println("房间号输入错误，请退出重新输入！")
			os.Exit(0)
		}
	}

	c, err := NewClient(roomId)
	if err != nil {
		log.Println("models.NewClient() error:", err)
		return
	}
	err = c.Start(context.Background(), func(message Message) {
		switch m := message.(type) {
		case *DanMuMsg:
			log.Printf("%d-%s | %d-%s: %s\n", m.MedalLevel, m.MedalName, m.ULevel, m.Uname, m.Text)
		case *Gift:
			log.Printf("%s %s 价值 %d 的 %s\n", m.UUname, m.Action, m.Price, m.GiftName)
		case *WelcomeVip:
			log.Printf("欢迎VIP %s 进入直播间", m.UserName)
		case *WelcomeGuard:
			log.Printf("欢迎房管 %s 进入直播间", m.UserName)
		case *WelcomeEntry:
			log.Println(m.Message)
		}
	})
	if err != nil {
		log.Println("c.Start() error:", err)
		return
	}

	select {}
}
