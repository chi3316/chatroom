package main

import (
	"fmt"
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

// 创建一个Server的接口
func NewService(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

// 监听Message广播消息channel的goroutine ,一旦有消息就全部发送给全部的在线user
func (this *Server) ListenMessager() {
	msg := <-this.Message
	//将msg发送给全部的在线用户
	this.mapLock.Lock()
	for _, cil := range this.OnlineMap {
		cil.C <- msg
	}
	this.mapLock.Unlock()
}

// 广播消息的方法
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sendMsg
}

// Handler 处理接收后的数据
func (this *Server) Handler(conn net.Conn) {
	fmt.Println("连接建立成功")
	user := NewUser(conn)
	//用户上线，将用户加到onlineMap中
	this.mapLock.Lock()
	this.OnlineMap[user.Name] = user
	fmt.Println("[" + user.Addr + "]" + user.Name + ":" + "上线成功")
	this.mapLock.Unlock()

	//广播当前用户上线的消息
	this.BroadCast(user, "已上线")

	//当前handle阻塞·
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
	go this.ListenMessager()
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
