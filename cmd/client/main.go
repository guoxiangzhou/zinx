package main

import (
	"fmt"
	"github.com/aceld/zinx/ziface"
	"github.com/aceld/zinx/znet"
	"time"
)

type PingRouter struct {
	znet.BaseRouter
}

func (router *PingRouter) Handle(request ziface.IRequest) {
	//fmt.Println("call helloRouter Handle")
	fmt.Printf("receive from server msgId=%d, data=%s\n", request.GetMsgID(), string(request.GetData()))

	err := request.GetConnection().SendMsg(1, []byte("client: ping"))
	if err != nil {
		fmt.Println(err)
	}
}

func OnConnected(conn ziface.IConnection)  {
	fmt.Println("client: OnConnected ------------------ ")
	conn.SendMsg(1, []byte("client: ping"))
}

func OnClosed(conn ziface.IConnection)  {
	fmt.Println("client: OnClosed ------------------ ")
}


func main()  {
	c := znet.NewClient("127.0.0.1", 8999)
	c.AddRouter(1, &PingRouter{})
	go c.Connect(OnConnected, OnClosed)

	select {
	case <-time.After(time.Second * 1):
		break
	}

	c.Close()

	time.Sleep(3 * time.Second)
}
