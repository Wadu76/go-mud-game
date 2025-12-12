package game

import "fmt"

type Room struct {
	Name        string
	Description string
	Exits       map[string]*Room
	//键为方向，值为下一个房间
}

//工厂函数
func NewRoom(name, desc string) *Room {
	return &Room{
		Name:        name,
		Description: desc,
		Exits:       make(map[string]*Room),
	}
}

//连接两个房间
func (r *Room) Link(direction string, next *Room) {
	r.Exits[direction] = next
	test_log := fmt.Sprintf("房间[%s]与房间[%s]Link成功！\n", r.Name, next.Name)
	fmt.Println(test_log)
	//让room指向下一个room
}

//获取房间描述,来到房间后可以看到的信息
func (r *Room) GetInfo() string {

	if r == nil {
		return "你来到了一个未知的地方..."
	}

	info := fmt.Sprintf("[%s]\n %s\n 可以看到的出口有:", r.Name, r.Description)
	//出口即与该room link的其他room，可能有多个所以我们后面拼接起来即可
	for dir := range r.Exits {
		info += dir + " "
	}
	return info
}
