package znet

import (
	"fmt"
	"github.com/aceld/zinx/ziface"
	"testing"
	"time"
)

// run in terminal:
// go test -v ./znet -run=TestServer

/*
	模拟客户端
*/
/*
func ClientTest(i uint32) {

	fmt.Println("Client Test ... start")
	//3秒之后发起测试请求，给服务端开启服务的机会
	time.Sleep(3 * time.Second)

	conn, err := net.Dial("tcp", "127.0.0.1:8999")
	if err != nil {
		fmt.Println("client start err, exit!")
		return
	}

	for {
		dp := NewDataPack()
		msg, _ := dp.Pack(NewMsgPackage(i, []byte("client test message")))
		_, err := conn.Write(msg)
		if err != nil {
			fmt.Println("client write err: ", err)
			return
		}

		//先读出流中的head部分
		headData := make([]byte, dp.GetHeadLen())
		_, err = io.ReadFull(conn, headData)
		if err != nil {
			fmt.Println("client read head err: ", err)
			return
		}

		// 将headData字节流 拆包到msg中
		msgHead, err := dp.Unpack(headData)
		if err != nil {
			fmt.Println("client unpack head err: ", err)
			return
		}

		if msgHead.GetDataLen() > 0 {
			//msg 是有data数据的，需要再次读取data数据
			msg := msgHead.(*Message)
			msg.Data = make([]byte, msg.GetDataLen())

			//根据dataLen从io中读取字节流
			_, err := io.ReadFull(conn, msg.Data)
			if err != nil {
				fmt.Println("client unpack data err")
				return
			}

			fmt.Printf("==> Client receive Msg: Id = %d, len = %d , data = %s\n", msg.Id, msg.DataLen, msg.Data)
		}

		time.Sleep(time.Second)
	}
}
*/
/*
	模拟服务器端
*/

//ping test 自定义路由
type ServerPingRouter struct {
	BaseRouter
}

//Test PreHandle
//func (this *PingRouter) PreHandle(request ziface.IRequest) {
//	fmt.Println("Call Router PreHandle")
//	err := request.GetConnection().SendMsg(1, []byte("before ping ....\n"))
//	if err != nil {
//		fmt.Println("preHandle SendMsg err: ", err)
//	}
//}

//Test Handle
func (this *ServerPingRouter) Handle(request ziface.IRequest) {
	//fmt.Println("Call PingRouter Handle")
	//先读取客户端的数据，再回写ping...ping...ping
	fmt.Println("recv from client : msgId=", request.GetMsgID(), ", data=", string(request.GetData()))

	err := request.GetConnection().SendMsg(1, []byte("server: ping"))
	if err != nil {
		fmt.Println("Handle SendMsg err: ", err)
	}
}

//Test PostHandle
//func (this *PingRouter) PostHandle(request ziface.IRequest) {
//	fmt.Println("Call Router PostHandle")
//	err := request.GetConnection().SendMsg(1, []byte("After ping .....\n"))
//	if err != nil {
//		fmt.Println("Post SendMsg err: ", err)
//	}
//}

type ClientRouter struct {
	BaseRouter
}

func (this *ClientRouter) Handle(request ziface.IRequest) {
	//fmt.Println("call helloRouter Handle")
	fmt.Printf("receive from server msgId=%d, data=%s\n", request.GetMsgID(), string(request.GetData()))

	err := request.GetConnection().SendMsg(1, []byte("client: ping"))
	if err != nil {
		fmt.Println(err)
	}
}

//func DoConnectionBegin(conn ziface.IConnection) {
//	fmt.Println("DoConnectionBegin is Called ... ")
//	err := conn.SendMsg(2, []byte("DoConnection BEGIN..."))
//	if err != nil {
//		fmt.Println(err)
//	}
//}
//
////连接断开的时候执行
//func DoConnectionLost(conn ziface.IConnection) {
//	fmt.Println("DoConnectionLost is Called ... ")
//}

func OnConnected(conn ziface.IConnection)  {
	fmt.Println("OnConnected ------------------ ")
	conn.SendMsg(1, []byte("client: ping"))
	//conn.SendMsg(2, []byte("client: hello"))
}

func OnClosed(conn ziface.IConnection)  {

}

func TestServer(t *testing.T) {
	//创建一个server句柄
	s := NewServer()

	// 多路由
	s.AddRouter(1, &ServerPingRouter{})
	//s.AddRouter(2, &HelloRouter{})

	go s.Start(OnConnected, nil)

	time.Sleep(1 * time.Second)

	c := NewClient("127.0.0.1", 8999)
	c.AddRouter(1, &ClientRouter{})
	//c.AddRouter(2, &HelloRouter{})
	go c.Connect(OnConnected, OnClosed)

	//	客户端测试
	//go ClientTest(1)
	//go ClientTest(2)

	//2 开启服务
	// go s.Serve()

	select {
	case <-time.After(time.Second * 1):
		return
	}

	c.Close()

	s.Stop()
}
