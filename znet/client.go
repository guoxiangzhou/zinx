package znet

import (
	"fmt"
	"github.com/aceld/zinx/ziface"
	"net"
)

type Client struct {
	//tcp connection
	TCPConntion ziface.IConnection
	//tcp4 or other
	IPVersion string
	//服务绑定的IP地址
	RemoteIP string
	//服务绑定的端口
	RemotePort int
	//当前Client的消息管理模块，用来绑定MsgId和对应的处理方法
	msgHandler ziface.IMsgHandle
}

func NewClient(ip string, port int) ziface.IClient {

	c := &Client{
		IPVersion:  "tcp4",
		RemoteIP:   ip,
		RemotePort: port,
		msgHandler: NewMsgHandle(),
	}
	return c
}

func (c *Client) Connect(connected func(conn ziface.IConnection), closed func(conn ziface.IConnection)) {
	fmt.Println("Client Connect ...")

	conn, err := net.Dial(c.IPVersion, fmt.Sprintf("%s:%d", c.RemoteIP, c.RemotePort))
	if err != nil {
		fmt.Println("client connect err, exit!")
		return
	}

	c.msgHandler.StartWorkerPool()
	c.TCPConntion = NewConntion(nil, conn, c.msgHandler, connected, closed)
	go c.TCPConntion.Start()
}

func (c *Client) Close() {
	if c.TCPConntion != nil {
		c.TCPConntion.Stop()
		c.TCPConntion = nil
	}
}

func (c *Client) AddRouter(msgId uint32, router ziface.IRouter) {
	c.msgHandler.AddRouter(msgId, router)
}
