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
	go readFromServer(conn)

	// 读取输入
	inputReader := bufio.NewReader(os.Stdin) //从控制台读取输入

	for {
		//读取用户输入
		//fmt.Print(">") ui混乱原因
		input, _ := inputReader.ReadString('\n')

		//发送到服务器
		_, err := conn.Write([]byte(input))
		if err != nil {
			fmt.Println("发送失败", err)
			break
		}

	}
}

func readFromServer(conn net.Conn) {
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("\n服务器断开连接")
			os.Exit(0)
		}

		fmt.Print(string(buf[:n]))
	}
}
