package bilidanmu

import (
	"context"
	"log"
	"testing"
)

func TestClient(t *testing.T) {
	c, err := NewClient(5050)
	if err != nil {
		t.Error("models.NewClient() error:", err)
		return
	}
	ctx := context.Background()
	err = c.Start(ctx, func(message Message) {
		switch m := message.(type) {
		case *DanMuMsg:
			log.Printf("%d-%s | %d-%s: %s\n", m.MedalLevel, m.MedalName, m.ULevel, m.Uname, m.Text)
		case *Gift:
			log.Printf("%s %s 价值 %d 的 %s\n", m.UUname, m.Action, m.Price, m.GiftName)
		case *WelcomeVip:
			log.Printf("欢迎VIP %s 进入直播间\n", m.UserName)
		case *WelcomeGuard:
			log.Printf("欢迎房管 %s 进入直播间\n", m.UserName)
		case *WelcomeEntry:
			log.Println(m.Message)
		}
	})
	if err != nil {
		t.Error("c.Start() error:", err)
		return
	}
	_ = c.WaitForStop()
}
