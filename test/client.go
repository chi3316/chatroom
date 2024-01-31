package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	// 连接到服务器
	conn, err := net.Dial("tcp", "localhost:8888")
	if err != nil {
		fmt.Println("Error connecting:", err.Error())
		return
	}
	defer conn.Close()

	fmt.Println("Connected to server. Type 'exit' to quit.")

	// 启动一个 goroutine 用于接收服务器消息
	go receiveMessages(conn)

	// 循环发送消息到服务器
	for {
		// 从用户输入读取消息
		reader := bufio.NewReader(os.Stdin)
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading message:", err.Error())
			return
		}

		// 移除末尾的换行符
		message = strings.TrimSpace(message)

		// 发送消息到服务器
		_, err = conn.Write([]byte(message))
		if err != nil {
			fmt.Println("Error sending message:", err.Error())
			return
		}

		// 如果用户输入 'exit'，退出循环
		if message == "exit" {
			break
		}
	}
}

// 接收服务器消息并打印到控制台
func receiveMessages(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading message:", err.Error())
			return
		}

		// 打印接收到的消息
		fmt.Println("Received message:", message)
	}
}
