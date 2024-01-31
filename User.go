package main

import (
	"fmt"
	"net"
	"strings"
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

// 给用户发送消息
func (this *User) sendMsg(msg string) {
	this.conn.Write([]byte(msg))
}

// 用户处理消息的业务
func (this *User) DoMessage(msg string) {
	//fmt.Println(msg)
	//msgTrimSpace := strings.TrimSpace(msg)
	//查询在线用户
	if msg == "who" {
		this.server.mapLock.Lock()
		for _, usre := range this.server.OnlineMap {
			onlineMsg := "[" + usre.Addr + "]" + usre.Name + ":" + "在线...\n"
			this.sendMsg(onlineMsg)
		}
		this.server.mapLock.Unlock()
	} else if len(msg) >= 7 && msg[:7] == "rename|" {
		//获取新的用户名
		newName := strings.Split(msg, "|")[1]
		//判断newName在用户表中是否存在
		//存在返回错误信息,不存在则更新用户名并返回信息给客户端
		if _, exists := this.server.OnlineMap[newName]; exists {
			this.sendMsg("用户名已被占用")
		} else {
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap, this.Name)
			this.server.OnlineMap[newName] = this
			this.server.mapLock.Unlock()

			this.Name = newName
			this.sendMsg("您已更新用户名:" + this.Name + "\n")
		}

	} else if len(msg) > 4 && msg[:3] == "to|" {
		//消息格式 ： to|张三|内容
		//1. 获取对方用户名
		name := strings.Split(msg, "|")[1]
		if name == "" {
			this.sendMsg("消息格式不正确，请使用:\"to|迟迟|hello\"的格式")
			return
		}
		//2. 获取用户名对应的user对象
		remoteUser, exists := this.server.OnlineMap[name]
		if !exists {
			this.sendMsg("用户不存在\n")
			return
		}

		//3. 给user发消息
		content := strings.Split(msg, "|")[2]
		if content == "" {
			this.sendMsg("无消息内容，请重发")
			return
		}
		remoteUser.sendMsg(this.Name + "发来了一条消息：" + content)
	} else {
		this.server.BroadCast(this, msg)
	}
}

// 监听当前User channel的方法，一旦有消息,就直接发送给对端客户端
func (this *User) ListenMessage() {
	for {
		msg := <-this.C
		this.conn.Write([]byte(msg + "\n"))
	}
}
