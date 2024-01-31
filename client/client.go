package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int //当前客户端的模式
}

func NewClient(serverIp string, serverPort int) *Client {
	//创建客户端对象
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       -1,
	}

	//连接对应的服务器
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial error:", err)
		return nil
	}

	client.conn = conn
	return client
}

// 处理server回应的消息，直接显示到标准输出
func (this *Client) ReceiveMsg() {
	//一旦有this.conn有数据，就直接copy到stdout标准输出上，永久阻塞监听
	io.Copy(os.Stdout, this.conn)
}

// 更新用户名
func (this *Client) UpdateName() bool {
	fmt.Println(">>>>>>>>请输入用户名：")
	fmt.Scanln(&this.Name)

	sendmsg := "rename|" + this.Name
	_, err := this.conn.Write([]byte(sendmsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return false
	}

	return true
}

// 查询在线用户
func (this *Client) selectUsers() {
	sendMsg := "who"
	_, err := this.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn Write err:", err)
		return
	}
}

// 私聊模式
func (this *Client) PrivateChat() {
	var remoteName string
	var chatMsg string
	this.selectUsers()
	fmt.Println(">>>>>>>>>>请输入聊天对象[用户名],exit退出")
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		fmt.Println(">>>>>>请输入聊天内容，exit退出.")
		fmt.Scanln(&chatMsg)
		for chatMsg != "exit" {
			//发送给服务器
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n"
				_, err := this.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn Write err:", err)
					break
				}
			}
			chatMsg = ""
			fmt.Println(">>>>>>请输入聊天内容，exit退出.")
			fmt.Scanln(&chatMsg)
		}

		this.selectUsers()
		fmt.Println(">>>>>>>>>>请输入聊天对象[用户名],exit退出")
		fmt.Scanln(&remoteName)
	}
}

// 公聊模式
func (this *Client) PublicChat() {
	var chatMsg string
	fmt.Println(">>>>>>请输入聊天内容，exit退出.")
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		//发送给服务器
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := this.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn Write err:", err)
				break
			}
		}
		chatMsg = ""
		fmt.Println(">>>>>>请输入聊天内容，exit退出.")
		fmt.Scanln(&chatMsg)
	}
}

// 客户端菜单
func (this *Client) menu() bool {
	var flag int
	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0. 退出")

	fmt.Scanln(&flag)
	if flag >= 0 && flag <= 3 {
		this.flag = flag
		return true
	} else {
		fmt.Println(">>>>>>请输入合法的数字")
		return false
	}
}

func (this *Client) Run() {
	for this.flag != 0 {
		for this.menu() != true {
		}
		switch this.flag {
		case 1:
			this.PublicChat()
			break
		case 2:
			this.PrivateChat()
			break
		case 3:
			this.UpdateName()
			break
		}
	}
}

var serverIp string
var serverPort int

// 命令行解析
// ./clien -ip 127.0.0.1 -port 8888
func init() {
	flag.StringVar(&serverIp, "ip", "localhost", "设置服务器ip地址（默认127.0.0.1）")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器端口（默认8888）")
}

func main() {
	//命令行解析
	flag.Parse()
	client := NewClient("localhost", 8888)
	if client == nil {
		fmt.Println(">>>>>>连接服务器失败...")
		return
	}
	//接收服务器发送的消息
	go client.ReceiveMsg()
	fmt.Println(">>>>>>连接服务器成功...")
	client.Run()
}
