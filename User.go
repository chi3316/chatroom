package main

import (
	"fmt"
	"net"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

// 创建一个用户
func NewUser(conn net.Conn, server *Server) *User {
	//获取客户端远程地址
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}
	//启动监听当前user channel 消息的goroutine
	go user.ListenMessage()
	return user
}

// 用户上线功能
func (this *User) Online() {
	//用户上线，将用户加到onlineMap中
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	fmt.Println("[" + this.Addr + "]" + this.Name + ":" + "上线成功")
	this.server.mapLock.Unlock()
	//广播当前用户上线的消息
	this.server.BroadCast(this, "已上线")
}

// 用户下线功能
func (this *User) Offline() {
	//用户上线，将用户从onlineMap中删除
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	fmt.Println("[" + this.Addr + "]" + this.Name + ":" + "已下线")
	this.server.mapLock.Unlock()
	//广播当前用户上线的消息
	this.server.BroadCast(this, "已下线")
}

// 用户处理消息的业务
func (this *User) DoMessage(msg string) {
	this.server.BroadCast(this, msg)
}

//监听当前User channel的方法，一旦有消息,就直接发送给对端客户端

func (this *User) ListenMessage() {
	for {
		msg := <-this.C
		this.conn.Write([]byte(msg + "\n"))
	}
}
