package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int

	//在线用户列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	Message chan string
}

// 创建server的接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg

	this.Message <- sendMsg
}

// 监听Message广播消息channel的协程，有消息就发送给全部在线user
func (this *Server) ListenMessager() {
	for {
		msg := <-this.Message
		//将msg发送给全部在线user
		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()

	}
}

func (this *Server) Hander(conn net.Conn) {
	//当前连接的业务
	//fmt.Println("当前业务success")
	user := NewUser(conn, this)
	//当前用户上线，将用户加入表中
	user.Online()
	//监听用户是否活跃的channel
	isLive:=make(chan bool)
	//接受客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Printf("Conn Read  err:", err)
				return
			}
			msg := string(buf[:n-1])

			//用户针对msg进行消息处理
			user.DoMessage(msg)

			isLive<-true

		}
	}()
	//当前hander阻塞
	for{
		select {
		case <-isLive:
			//当前用户活跃,应该重置定时器

		case <-time.After(time.Second*100):
			//已经超时，将当前user强制关闭
			user.sendMsg("你被踢了")

			close(user.C)

			conn.Close()

			//退出
			return
		}
	}
	
}

// 启动服务器的接口
func (this *Server) Start() {
	//socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}
	defer listener.Close()
	//启动监听Message的协程
	go this.ListenMessager()
	for {
		//accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener  accept err:", err)
			continue
		}

		//do hander
		go this.Hander(conn)
	}

	// close listen socket
}
