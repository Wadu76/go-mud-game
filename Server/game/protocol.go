package game

//服务器发给Unity客户端的标准数据包
type ServerMessage struct {
	Event string      `json:"event"` //事件类型：比如chat，登录，生命值改变等
	Data  interface{} `json:"data"`  //具体数据 interface{}可以包含任何类型的数据
}

//专门用来同步血量的数据
type HPUpdateMsg struct {
	//掉血对象
	TargetName string `json:"target_name"`
	//当前血量
	CurrentHP int `json:"current_hp"`
	//最大血量
	MaxHP int `json:"max_hp"`
}

//那么为了能发送，就要改attack逻辑，使其可以发送json
//发送json给unity后就能在unity中处理了，可以弄个血条了
//attack：monster / player都有 ， player的在server.go里面
