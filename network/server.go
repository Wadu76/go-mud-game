package network

import (
	"fmt"
	"net"
)

//StartServer 启动TCP服务器
func StartServer() {
	//1监听端口 8888
	listener, err := net.Listen("tcp", ":8888")  //err是错误信息， listener是监听对象
	if err != nil {
		fmt.Println("启动服务器失败：", err)
		return 
	}
	//defer确保函数退出前关闭listener
	defer listener.Close() 

	fmt.Println(" 🚀游戏服务已启动，正在监听8888端口...")

	//2等待客户端连接，无限循环
	for{
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


//处理单个玩家的连接
func handleConnection(conn net.Conn) {
	defer conn.Close()	//玩家断开时关闭连接

	buf := make([]byte, 1024) //缓冲区
	for {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("玩家断开连接:", conn.RemoteAddr())
			return
		}

		//处理消息 把收到的数据转成字符串
		msg := string(buf[:n])
		fmt.Printf("收到信息: %s\n", msg)

		//给玩家回复信息
		conn.Write([]byte("服务器已收到你的消息:" + msg))
	}

}