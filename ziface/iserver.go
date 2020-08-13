package ziface

//定义服务器接口
type IServer interface {
	//启动服务器方法
	Start(connected func(conn IConnection), closed func(conn IConnection))
	//停止服务器方法
	Stop()
	//路由功能：给当前服务注册一个路由业务方法，供客户端链接处理使用
	AddRouter(msgId uint32, router IRouter)
	//得到链接管理
	GetConnMgr() IConnManager
}
