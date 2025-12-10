package network

//管理所有在线玩家的连接，时刻听着广播通道
import (
	"fmt"
	"net"
	"sync"
)

// 全局world
type World struct {
	//读写锁
	mu sync.RWMutex

	//在线玩家列表，key为玩家地址（string） value是连接net.conn
	OnlinePlayers map[string]net.Conn

	//广播通道
	MessageChannel chan string
}

// 全局变量，整个游戏就只有一个世界
var GlobalWorld *World

func InitWorld() {
	GlobalWorld = &World{
		OnlinePlayers:  make(map[string]net.Conn),
		MessageChannel: make(chan string, 10), //缓冲区大小10

	}

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
func (w *World) AddPlayer(conn net.Conn) {
	w.mu.Lock()
	//RemoteAddr()返回远程地址，类型为net.Addr,其对应.String()方法返回字符串格式的地址,即玩家名
	w.OnlinePlayers[conn.RemoteAddr().String()] = conn
	w.mu.Unlock()

	w.MessageChannel <- fmt.Sprintf("🔈 系统广播: 玩家 [%s] 加入了游戏! \n>", conn.RemoteAddr())
}

// 玩家离开游戏
func (w *World) RemovePlayer(conn net.Conn) {
	w.mu.Lock()
	delete(w.OnlinePlayers, conn.RemoteAddr().String())
	w.mu.Unlock()

	w.MessageChannel <- fmt.Sprintf("🔈 系统广播: 玩家 [%s] 离开了游戏! \n>", conn.RemoteAddr())
}
