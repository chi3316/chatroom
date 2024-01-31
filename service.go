package main

import (
	"fmt"
	"net"
)

type Server struct {
	Ip   string
	Port int
}

// 创建一个Server的接口
func NewService(ip string, port int) *Server {
	server := &Server{
		Ip:   ip,
		Port: port,
	}
	return server
}

// Handler 处理接收后的数据
func (this *Server) Handler(conn net.Conn) {
	//TODO : 当前业务的逻辑
	fmt.Println("连接建立成功")
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
