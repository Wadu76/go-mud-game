package network

import (
	"fmt"
	"mud-server/game"
	"net"
	"strings"
)

// StartServer 启动TCP服务器
func StartServer() {
	//1监听端口 8888
	listener, err := net.Listen("tcp", ":8888") //err是错误信息， listener是监听对象
	if err != nil {
		fmt.Println("启动服务器失败：", err)
		return
	}
	//defer确保函数退出前关闭listener
	defer listener.Close()

	fmt.Println(" 🚀游戏服务已启动，正在监听8888端口...")

	//2等待客户端连接，无限循环a
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("连接建立失败:", err)
			continue
		}
		fmt.Println(" 有新玩家连接:", conn.RemoteAddr())

		//3开启一个Goroutine处理新玩家
		go handleConnection(conn)
	}
}

// 处理单个玩家的连接
func handleConnection(conn net.Conn) {
	defer conn.Close() //玩家断开时关闭连接

	fmt.Println("新玩家接入，正在初始化游戏数据...")

	//初始化游戏数据,以后会存入全局
	hero := game.NewPlayer("瓦度", 1, 100, 100)
	monster := game.NewMonster("史莱姆王", 50, 50, 20)

	conn.Write([]byte("===== 欢迎来到GO MUD 在线测试版 =====\n 请输入 attack, heal, status\n>"))

	buf := make([]byte, 1024) //缓冲区
	for {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("玩家断开连接:", conn.RemoteAddr())
			return
		}

		//去掉空格和换行
		input := string(buf[:n])
		command := strings.TrimSpace(strings.ToLower(input))

		//处理空指令
		if command == "" {
			conn.Write([]byte("> "))
			continue
		}
		//游戏逻辑路由
		var response string //f发回给客户端的话

		switch command {
		case "attack":
			log1 := hero.Attack(monster)
			response = log1 + "\n"

			if monster.HP > 0 {
				log := monster.Attack(hero)
				response += log + "\n"
			} else {
				response += fmt.Sprintf("成功击败了史莱姆王！获得 %d 经验\n", monster.Exp)
			}

		case "heal":
			log1 := hero.Heal()
			log2 := monster.Attack(hero)
			response = log1 + "\n" + log2 + "\n"

		case "status":
			response = fmt.Sprintf("状态：[%s] HP: %d/%d VS [%s] HP: %d/%d", hero.Name, hero.HP, hero.MaxHP, monster.Name, monster.HP, monster.MaxHP)

		case "exit":
			conn.Write([]byte("加纳！\n"))
			return

		default:
			response = fmt.Sprintf("未知指令 '%s'，请输入 attack, heal, status\n", command)
		}

		if hero.HP <= 0 {
			response += "/(ㄒoㄒ)/~~ 胜败乃兵家常事，重新连接复活再来吧！\n"
			conn.Write([]byte(response))
			return //踢走输掉的玩家
		}
		//最终战斗信息
		response += ">"
		conn.Write([]byte(response))

	}

}
