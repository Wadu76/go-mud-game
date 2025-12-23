package network

//管理所有在线玩家的连接，时刻听着广播通道
import (
	"fmt"
	"mud-server/database"
	"mud-server/game"
	"net"
	"sync"
)

// 全局world
type World struct {
	//读写锁
	mu sync.RWMutex

	//在线玩家列表，key为玩家地址（string） value是连接net.conn
	//key接下来改为玩家名字 依旧是string
	OnlinePlayers map[string]net.Conn

	//广播通道
	MessageChannel chan string

	//世界内共用的怪物
	Boss *game.Monster

	//出生点房间，玩家上线会自动进入
	StartRoom *game.Room

	AllRooms map[string]*game.Room
}

// 全局变量，整个游戏就只有一个世界
var GlobalWorld *World

// 把所有房间放一起
// var AllRooms map[string]*game.Room
func InitWorld() {

	//创建房间
	//make(map[string]*game.Room)
	//InitAllRoomsTogether()
	town := game.NewRoom("新手村广场", "这里是梦开始的地方，十分安全，可以在这里接冒险者工会的任务。")
	forest := game.NewRoom("黑暗森林", "在广场旁边的森林，这里树木丛生，传来着各种奇奇怪怪的声音...")
	cave := game.NewRoom("恶龙巢穴", "深不见底的洞穴，这里甚至能闻到硫磺味。")

	tempRooms := make(map[string]*game.Room)
	town.AddToMap(tempRooms)
	forest.AddToMap(tempRooms)
	cave.AddToMap(tempRooms)
	//连接房间
	fmt.Println("房间创建完成，开始连接房间...") // 添加调试信息

	//连接各个房间
	//广场北边是森林
	town.Link("north", forest)
	//森林南边是广场
	forest.Link("south", town)

	//森林东边是洞穴
	forest.Link("east", cave)
	//洞穴西边是森林
	cave.Link("west", forest)

	GlobalWorld = &World{
		OnlinePlayers:  make(map[string]net.Conn),
		MessageChannel: make(chan string, 10), //缓冲区大小10

		//boss赋值为Newmonster的返回值，即Monster这个结构体
		Boss: game.NewMonster("史莱姆王", 100, 100, 50),

		//在此处初始化出生点房间
		StartRoom: town,

		AllRooms: tempRooms,
	}
	//cesh
	//town.Items["tword"] = game.NewItem("tword", "test") 剑只在内存里，没有数据库ID，所以不能这样写
	loadWorldItems() //加载世界物品
	//InitAllRoomsTogether()   //把所有房间放一起
	//启动独立的Goroutine，负责分发广播
	go GlobalWorld.BroadcastLoop()
}

//持续从通道1拿信息，拿到了就发给所有人

func (w *World) BroadcastLoop() {
	for {
		//从chan拿消息，若没有就等
		msg := <-w.MessageChannel

		//遍历所有在线玩家，发送消息
		w.mu.RLock() //加读锁
		//在线玩家列表，key为玩家地址（string） value是连接net.conn
		for addr, conn := range w.OnlinePlayers {
			//把msg发给每个conn

			conn.Write([]byte(msg))
			fmt.Printf("已广播给 %s: %s", addr, msg)
		}
		w.mu.RUnlock() //解锁
	}
}

// 玩家加入游戏
func (w *World) AddPlayer(name string, conn net.Conn) {
	w.mu.Lock()
	//RemoteAddr()返回远程地址，类型为net.Addr,其对应.String()方法返回字符串格式的地址,即玩家名
	w.OnlinePlayers[name] = conn //名字作为key
	w.mu.Unlock()

	//w.MessageChannel <- fmt.Sprintf("🔈 系统广播: 玩家 [%s] 加入了游戏! \n>", conn.RemoteAddr())
}

// 玩家离开游戏
func (w *World) RemovePlayer(name string, conn net.Conn) {
	w.mu.Lock()

	delete(w.OnlinePlayers, name)
	w.mu.Unlock()

	//w.MessageChannel <- fmt.Sprintf("🔈 系统广播: 玩家 [%s] 离开了游戏! \n>", conn.RemoteAddr())
}

// 房间内部广播
func (w *World) BroadcastToRoom(room *game.Room, msg string) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	//遍历该房间有的玩家
	for playerName := range room.Players {
		//
		if conn, ok := w.OnlinePlayers[playerName]; ok {
			conn.Write([]byte(msg))
		}
	}
}

// 用于读取无主的物品放到对应房间里
func loadWorldItems() {
	var items []game.Item
	//找出所有RoomName不为空的物品
	database.DB.Where("room_name != ''").Find(&items)

	for _, item := range items {
		if room, ok := GlobalWorld.AllRooms[item.RoomName]; ok {
			newItem := item
			room.Items[item.Name] = &newItem
			fmt.Printf("加载物品: %s 到 %s\n", item.Name, item.RoomName)
		}
	}
}
