package logic

type GlobalConf struct {
	BrokerIp string
	BrokerPort int
	AgentPort int
}

var GConf GlobalConf
