package ziface

//定义客户端接口
type IClient interface {
	//启动服务器方法
	Connect(func(conn IConnection), func(conn IConnection))
	//停止服务器方法
	Close()
	//路由功能：供客户端链接处理使用
	AddRouter(msgId uint32, router IRouter)
}
