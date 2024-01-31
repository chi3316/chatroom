package main

import (
	"fmt"
	"io"
	"net"
	"sync"
)

type Server struct {
	Ip   string
	Port int
	//存储在线用户的表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex
	//消息广播的channel
	Message chan string
}

// NewService 创建一个Server的接口
func NewService(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

// ListenMessage 监听Message广播消息channel的goroutine ,一旦有消息就全部发送给在线user
func (this *Server) ListenMessage() {
	for {
		msg := <-this.Message
		//将msg发送给全部的在线用户
		this.mapLock.Lock()
		for _, cil := range this.OnlineMap {
			cil.C <- msg
		}
		this.mapLock.Unlock()
	}
}

// BroadCast 创建广播消息内容的方法
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sendMsg
}

// Handler 处理接收后的数据
func (this *Server) Handler(conn net.Conn) {
	user := NewUser(conn, this)
	//用户上线，将用户加到onlineMap中
	user.Online()
	//接收用户发送的消息并广播
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err", err)
			}

			//提取用户信息，去除"\n"
			msg := string(buf[:n])
			//msg := string(buf)
			//fmt.Println(msg)
			//用户针对msg进行消息处理
			user.DoMessage(msg)
		}
	}()

	//当前handle阻塞
	select {}
}

// 启动服务器的方法
func (this *Server) Start() {
	//Socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen err", err)
		return
	}

	//close listen socket
	defer listener.Close()
	//启动监听Message的goroutine
	go this.ListenMessage()
	//accept
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err", err)
			continue
		}
		//do handler
		go this.Handler(conn)
	}
}
