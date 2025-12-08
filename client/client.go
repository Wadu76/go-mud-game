package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	fmt.Println("正在连接服务器...")
	conn, err := net.Dial("tcp", "127.0.0.1:8888")
	if err != nil {
		fmt.Println("连接失败", err)
		return 
	}

	defer conn.Close()
	fmt.Println("连接成功！请输入消息（输入exit退出）：")

	// 读取输入
	inputReader := bufio.NewReader(os.Stdin) //从控制台读取输入

	for {
		//读取用户输入
		fmt.Print(">")
		input, _ := inputReader.ReadString('\n')
		
		//发送到服务器
		_, err := conn.Write([]byte(input))
		if err != nil {
		    fmt.Println("发送失败", err)
			break
		}

		//接收服务器返回的消息
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("服务器断开连接")
			break
		}

		fmt.Println("服务器返回的消息：", string(buf[:n]))
		
	}
}