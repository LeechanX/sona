package logic

type GlobalConf struct {
	BrokerPort int
	//连接个数限制
	ConnectionLimit int
}

var GConf GlobalConf