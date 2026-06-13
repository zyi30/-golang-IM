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
	flag       int
}

func NewClient(serverip string, serverport int) *Client {
	//创建客户端
	client := &Client{
		ServerIp:   serverip,
		ServerPort: serverport,
		flag:       99,
	}
	//连接
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverip, serverport))
	if err != nil {
		fmt.Printf("net.Dial error", err)
		return nil
	}
	client.conn = conn
	//返回客户端对象
	return client
}

// 处理server回应
func (client *Client) DealResponse() {
	io.Copy(os.Stdout, client.conn)
}
func (client *Client) menu() bool {

	var flag int
	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")

	fmt.Scanln(&flag)
	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println("不合法，请输入合法的数字")
		return false
	}
}

func (client *Client) PublicChat() {
	//提示用户发送消息
	var chatmsg string
	fmt.Println("请输入来聊天内容,exit退出")
	fmt.Scanln(&chatmsg)
	for chatmsg != "exit" {
		if len(chatmsg) != 0 {
			sendmsg := chatmsg + "\n"
			_, err := client.conn.Write([]byte(sendmsg))
			if err != nil {
				fmt.Println("conn Write err:", err)
				break
			}
		}
		chatmsg = ""
		fmt.Println("请输入来聊天内容,exit退出")
		fmt.Scanln(&chatmsg)
	}

}
func (client *Client) UpdateNmae() bool {
	fmt.Printf("请输入用户名")
	fmt.Scanln(&client.Name)
	sendMsg := "rename|" + client.Name + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return false
	}
	return true

}
func (client *Client) Selectusers() {
	sendmsg := "who\n"
	_, err := client.conn.Write([]byte(sendmsg))
	if err != nil {
		fmt.Println("conn Write err:", err)
		return
	}
}

func (client *Client) PrivateChat() {
	var remotename string
	var chatmsg string
	client.Selectusers()
	fmt.Println("请输入要聊天用户的用户名,exit退出")
	fmt.Scanln(&remotename)

	for remotename != "exit" {
		fmt.Println("请输入聊天内容")
		fmt.Scanln(&chatmsg)
		for chatmsg != "exit" {
			if len(chatmsg) != 0 {
				sendmsg := "to|" + remotename + "|" + chatmsg + "\n\n"
				_, err := client.conn.Write([]byte(sendmsg))
				if err != nil {
					fmt.Println("conn Write err:", err)
					break
				}
			}
			chatmsg = ""
			fmt.Println("请输入来聊天内容,exit退出")
			fmt.Scanln(&chatmsg)
		}
		client.Selectusers()
		fmt.Println("请输入要聊天用户的用户名,exit退出")
		fmt.Scanln(&remotename)
	}

}
func (client *Client) run() {
	for client.flag != 0 {
		for client.menu() != true {
		}
		//根据不同模式处理不同业务
		switch client.flag {
		case 1:
			client.PublicChat()
			break
		case 2:
			client.PrivateChat()
			break
		case 3:
			client.UpdateNmae()
			break
		}
	}
}

var ServerIp string
var ServerPort int

func init() {
	flag.StringVar(&ServerIp, "ip", "127.0.0.1", "设置服务器IP地址")
	flag.IntVar(&ServerPort, "port", 8888, "设置服务器端口")
}

func main() {
	//命令行解析
	flag.Parse()
	client := NewClient(ServerIp, ServerPort)
	if client == nil {
		fmt.Println(">>>>连接服务器失败")
		return
	}
	//接受消息
	go client.DealResponse()
	fmt.Println("连接服务器成功")

	//启动客户端业务(发送消息端)
	client.run()
}
