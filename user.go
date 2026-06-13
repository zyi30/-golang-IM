package main

import (
	"net"
	"strings"
)

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn

	server *Server
}

// 创建一个用户API
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}
	//启动监听当前user channel的协程
	go user.ListenMessage()

	return user
}

// 监听当前User channel 的方法
func (this *User) ListenMessage() {
	for {
		msg := <-this.C
		this.conn.Write([]byte(msg + "\n"))
	}
}

// 用户上线业务
func (this *User) Online() {
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()
	//广播上线消息
	this.server.BroadCast(this, "已上线")
}

// 用户下线业务
func (this *User) Offline() {
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()
	//广播下线消息
	this.server.BroadCast(this, "下线")
}
//发送消息给user对应的客户端
func(this*User)sendMsg(msg string){
	this.conn.Write([]byte(msg))
}
// 用户处理消息的业务
func (this *User) DoMessage(msg string) {
	if msg == "who"{
		//查询当前在线用户都有哪些
		this.server.mapLock.Lock()
		for _,user:=range this.server.OnlineMap{
			onlineMsg:="["+user.Addr+"]"+user.Name+":"+"在线...\n"
			this.sendMsg(onlineMsg)
		}
		this.server.mapLock.Unlock()
	}else if len(msg)>7&&msg[:7]=="rename|"{
		newName:=strings.Split(msg,"|")[1]
		//判断name是否存在
		_,ok:=this.server.OnlineMap[newName]
		if ok{
			this.sendMsg("当前用户名被使用\n")
		}else{
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap,this.Name)
			this.server.OnlineMap[newName]=this
			this.server.mapLock.Unlock()

			this.Name=newName
			this.sendMsg("您已经更新用户名:"+this.Name+"\n")
		}
	}else if len(msg)>4&&msg[:3]=="to|"{
		//1.获取对方应户名
		remoteName:=strings.Split(msg,"|")[1]
		//2.获取user对象
		if remoteName==""{
			this.sendMsg("消息格式不正确")
			return
		}
		remoteUser,ok:=this.server.OnlineMap[remoteName]
		if !ok{
			this.sendMsg("该对象不存在")
			return 
		}
		//3.获取消息内容，，通过user对象发送
		content:=strings.Split(msg,"|")[2]
		if content==""{
			this.sendMsg("无消息内容,请重发\n")
			return
		}
		remoteUser.sendMsg(this.Name+"    speak   :"+content)
	}else{
		this.server.BroadCast(this, msg)
	}
	
}

