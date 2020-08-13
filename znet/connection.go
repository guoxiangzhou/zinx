package znet

import (
	"errors"
	"fmt"
	"github.com/aceld/zinx/utils"
	"github.com/aceld/zinx/ziface"
	"io"
	"net"
	"sync"
)

type Connection struct {
	//当前Conn属于哪个Server
	TcpServer ziface.IServer
	//当前连接的socket TCP套接字
	Conn net.Conn
	//当前连接的关闭状态
	isClosed bool
	//消息管理MsgId和对应处理方法的消息管理模块
	MsgHandler ziface.IMsgHandle
	//告知该链接已经退出/停止的channel
	ExitBuffChan chan bool
	//无缓冲管道，用于读、写两个goroutine之间的消息通信
	msgChan chan []byte
	//有缓冲管道，用于读、写两个goroutine之间的消息通信
	msgBuffChan chan []byte

	//链接属性
	property map[string]interface{}
	//保护链接属性修改的锁
	propertyLock sync.RWMutex

	//连接成功的回调
	onConnectedCallback func(conn ziface.IConnection)
	//连接关闭的回调
	onClosedCallback    func(conn ziface.IConnection)
}

//创建连接的方法
func NewConntion(server ziface.IServer, conn net.Conn, msgHandler ziface.IMsgHandle,
	connectedCB func(ziface.IConnection), closeCB func(ziface.IConnection)) *Connection {
	//初始化Conn属性
	c := &Connection{
		TcpServer:           server,
		Conn:                conn,
		isClosed:            false,
		MsgHandler:          msgHandler,
		ExitBuffChan:        make(chan bool, 1),
		msgChan:             make(chan []byte),
		msgBuffChan:         make(chan []byte, utils.GlobalObject.MaxMsgChanLen),
		property:            make(map[string]interface{}),
		onConnectedCallback: connectedCB,
		onClosedCallback:    closeCB,
	}

	//将新创建的Conn添加到链接管理中
	if c.TcpServer != nil {
		c.TcpServer.GetConnMgr().Add(c)
	}

	return c
}

/*
	写消息Goroutine， 用户将数据发送给客户端
*/
func (c *Connection) StartWriter() {
	fmt.Println("[Writer Goroutine is running]")
	defer fmt.Println(c.RemoteAddr().String(), "[conn Writer exit!]")

	for {
		select {
		case data := <-c.msgChan:
			//有数据要写给客户端
			if _, err := c.Conn.Write(data); err != nil {
				fmt.Println("Send Data error:, ", err, " Conn Writer exit")
				return
			}
			//fmt.Printf("Send data succ! data = %+v\n", data)
		case data, ok := <-c.msgBuffChan:
			if ok {
				//有数据要写给客户端
				if _, err := c.Conn.Write(data); err != nil {
					fmt.Println("Send Buff Data error:, ", err, " Conn Writer exit")
					return
				}
			} else {
				fmt.Println("msgBuffChan is Closed")
				break
			}
		case <-c.ExitBuffChan:
			return
		}
	}
}

/*
	读消息Goroutine，用于从客户端中读取数据
*/
func (c *Connection) StartReader() {
	fmt.Println("[Reader Goroutine is running]")
	defer fmt.Println(c.RemoteAddr().String(), "[conn Reader exit!]")
	defer c.Stop()

	for {
		// 创建拆包解包的对象
		dp := NewDataPack()

		//读取客户端的Msg head
		headData := make([]byte, dp.GetHeadLen())
		if _, err := io.ReadFull(c.Conn, headData); err != nil {
			fmt.Println("read msg head error ", err)
			break
		}
		//fmt.Printf("read headData %+v\n", headData)

		//拆包，得到msgid 和 datalen 放在msg中
		msg, err := dp.Unpack(headData)
		if err != nil {
			fmt.Println("unpack error ", err)
			break
		}

		//根据 dataLen 读取 data，放在msg.Data中
		var data []byte
		if msg.GetDataLen() > 0 {
			data = make([]byte, msg.GetDataLen())
			if _, err := io.ReadFull(c.Conn, data); err != nil {
				fmt.Println("read msg data error ", err)
				break
			}
		}
		msg.SetData(data)

		//得到当前客户端请求的Request数据
		req := Request{
			conn: c,
			msg:  msg,
		}

		if utils.GlobalObject.WorkerPoolSize > 0 {
			//已经启动工作池机制，将消息交给Worker处理
			c.MsgHandler.SendMsgToTaskQueue(&req)
		} else {
			//从绑定好的消息和对应的处理方法中执行对应的Handle方法
			go c.MsgHandler.DoMsgHandler(&req)
		}
	}
}

//启动连接，让当前连接开始工作
func (c *Connection) Start() {
	//1 开启用户从客户端读取数据流程的Goroutine
	go c.StartReader()
	//2 开启用于写回客户端数据流程的Goroutine
	go c.StartWriter()
	//按照用户传递进来的创建连接时需要处理的业务，执行钩子方法
	if c.onConnectedCallback != nil {
		c.onConnectedCallback(c)
	}
}

//停止连接，结束当前连接状态M
func (c *Connection) Stop() {
	//如果当前链接已经关闭
	if c.isClosed == true {
		return
	}
	c.isClosed = true

	fmt.Println("Conn Stop()...ConnID = ", c.GetNetConn())

	//如果用户注册了该链接的关闭回调业务，那么在此刻应该显示调用
	if c.onClosedCallback != nil {
		c.onClosedCallback(c)
	}

	// 关闭socket链接
	c.Conn.Close()
	//关闭Writer
	c.ExitBuffChan <- true

	//将链接从连接管理器中删除
	if c.TcpServer != nil {
		c.TcpServer.GetConnMgr().Remove(c)
	}

	//关闭该链接全部管道
	close(c.ExitBuffChan)
	close(c.msgBuffChan)
}

//从当前连接获取原始的socket TCPConn
func (c *Connection) GetNetConn() net.Conn {
	return c.Conn
}

//获取远程客户端地址信息
func (c *Connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

//直接将Message数据发送数据给远程的TCP客户端
func (c *Connection) SendMsg(msgId uint32, data []byte) error {
	if c.isClosed == true {
		return errors.New("connection closed when send msg")
	}
	//将data封包，并且发送
	dp := NewDataPack()
	msg, err := dp.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		fmt.Println("Pack error msg id = ", msgId)
		return errors.New("Pack error msg ")
	}

	//写回客户端
	c.msgChan <- msg

	return nil
}

func (c *Connection) SendBuffMsg(msgId uint32, data []byte) error {
	if c.isClosed == true {
		return errors.New("Connection closed when send buff msg")
	}
	//将data封包，并且发送
	dp := NewDataPack()
	msg, err := dp.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		fmt.Println("Pack error msg id = ", msgId)
		return errors.New("Pack error msg ")
	}

	//写回客户端
	c.msgBuffChan <- msg

	return nil
}

//设置链接属性
func (c *Connection) SetProperty(key string, value interface{}) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	c.property[key] = value
}

//获取链接属性
func (c *Connection) GetProperty(key string) (interface{}, error) {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()

	if value, ok := c.property[key]; ok {
		return value, nil
	} else {
		return nil, errors.New("no property found")
	}
}

//移除链接属性
func (c *Connection) RemoveProperty(key string) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	delete(c.property, key)
}
