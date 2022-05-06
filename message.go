package bilidanmu

type Message interface {
	Decode([]byte)
}

type DanMuMsg struct {
	UID        uint32 `json:"uid"`
	Uname      string `json:"uname"`
	ULevel     uint32 `json:"ulevel"`
	Text       string `json:"text"`
	MedalLevel uint32 `json:"medal_level"`
	MedalName  string `json:"medal_name"`
}

func NewDanMuMsg() *DanMuMsg {
	return &DanMuMsg{
		UID:        0,
		Uname:      "",
		ULevel:     0,
		Text:       "",
		MedalLevel: 0,
		MedalName:  "无勋章",
	}
}

func (d *DanMuMsg) Decode(src []byte) {
	d.UID = json.Get(src, "info", 2, 0).ToUint32()
	d.Uname = json.Get(src, "info", 2, 1).ToString()
	d.ULevel = json.Get(src, "info", 4, 0).ToUint32()
	d.Text = json.Get(src, "info", 1).ToString()
	d.MedalName = json.Get(src, "info", 3, 1).ToString()
	if d.MedalName == "" {
		d.MedalName = "无勋章"
	}
	d.MedalLevel = json.Get(src, "info", 3, 0).ToUint32()
	return
}

type Gift struct {
	UUname   string `json:"u_uname"`
	Action   string `json:"action"`
	Price    uint32 `json:"price"`
	GiftName string `json:"gift_name"`
}

func NewGift() *Gift {
	return &Gift{
		UUname:   "",
		Action:   "",
		Price:    0,
		GiftName: "",
	}
}

func (g *Gift) Decode(src []byte) {
	g.UUname = json.Get(src, "data", "uname").ToString()
	g.Action = json.Get(src, "data", "action").ToString()
	nums := json.Get(src, "data", "num").ToUint32()
	g.Price = json.Get(src, "data", "price").ToUint32() * nums
	g.GiftName = json.Get(src, "data", "giftName").ToString()
}

type WelcomeVip struct {
	UserName string
}

func NewWelcomeVip() *WelcomeVip {
	return &WelcomeVip{
		UserName: "",
	}
}

func (u *WelcomeVip) Decode(src []byte) {
	u.UserName = json.Get(src, "data", "uname").ToString()
}

type WelcomeGuard struct {
	UserName string
}

func NewWelcomeGuard() *WelcomeGuard {
	return &WelcomeGuard{
		UserName: "",
	}
}

func (u *WelcomeGuard) Decode(src []byte) {
	u.UserName = json.Get(src, "data", "username").ToString()
}

type WelcomeEntry struct {
	Message string
}

func NewWelcomeEntry() *WelcomeEntry {
	return &WelcomeEntry{
		Message: "",
	}
}

func (u *WelcomeEntry) Decode(src []byte) {
	u.Message = json.Get(src, "data", "copy_writing").ToString()
}

type CMD string

var (
	CMDDanmuMsg     CMD = "DANMU_MSG"     // 普通弹幕信息
	CMDSendGift     CMD = "SEND_GIFT"     // 普通的礼物，不包含礼物连击
	CMDWelcomeVip   CMD = "WELCOME"       // 欢迎VIP
	CMDWelcomeGuard CMD = "WELCOME_GUARD" // 欢迎房管
	CMDEntry        CMD = "ENTRY_EFFECT"  // 欢迎舰长等头衔
)
