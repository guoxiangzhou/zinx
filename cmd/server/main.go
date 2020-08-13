package main

import (
	"fmt"
	"github.com/aceld/zinx/ziface"
	"github.com/aceld/zinx/znet"
)

type Router struct {
	znet.BaseRouter
}

func (router *Router) Handle(request ziface.IRequest) {
	fmt.Println("recv from client : msgId=", request.GetMsgID(), ", data=", string(request.GetData()))
	err := request.GetConnection().SendMsg(1, []byte("server: ping"))
	if err != nil {
		fmt.Println("Handle SendMsg err: ", err)
	}
}

func OnConnected(conn ziface.IConnection)  {
	fmt.Println("server: OnConnected ------------------ ", conn.GetConnID())
	conn.SendMsg(1, []byte("client: ping"))
}

func OnClosed(conn ziface.IConnection)  {
	fmt.Println("server: OnClosed ------------------ ", conn.GetConnID())
}

func main()  {
	s := znet.NewServer()

	// 多路由
	s.AddRouter(1, &Router{})
	//s.AddRouter(2, &HelloRouter{})

	go s.Start(OnConnected, OnClosed)

	select {
	}
}