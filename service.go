package main

import (
	"fmt"
	"io"
	"net"
	"runtime"
	"sync"
	"time"
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
	//监听用户是否活跃
	isLive := make(chan bool)
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
			//代表当前用户是活跃的
			isLive <- true
		}
	}()

	//当前handle阻塞
	for {
		select {
		case <-isLive:
			//当前用户是活跃的，重置定时器。不需要做任何处理，为了激活select，更新下面的定时器
		case <-time.After(time.Second * 60 * 10):
			//已经超时，将当前user强制关闭
			user.sendMsg("你被踢了")
			//销毁资源
			close(user.C)
			//关闭连接·
			conn.Close()
			//退出当前的Handler
			runtime.Goexit()
		}
	}

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
