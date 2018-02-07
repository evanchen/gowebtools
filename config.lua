return 
{
	--服务器网络ip
	servIp = "0.0.0.0",
	--服务器网络端口
	servPort = 7000,
	--服务器id
	hostId = 121,
	--game服务器数量(包括0进程)
	gameNum = 3,
	--是否开放gm指令
	gmopen = true,
	--是否记录协议日志
	printpto = true,
	--net进程协议通信地址,不同平台下配置不一样
	  --windows下zmq不支持ipc,用tcp格式代替
	net_ipc_bind_addr_win = "tcp://*:8888",
	net_ipc_bind_addr_linux = "ipc://net.ipc",
	--game进程通信地址格式
	--windows下的需要端口,port = servPort + gsId + 1
	game_ipc_bind_addr_win_fmt = "tcp://127.0.0.1:%d",
	game_ipc_bind_addr_linux_fmt = "ipc://game_%d.ipc",
	--日志进程ipc地址
	log_ipc_bind_addr_win = "tcp://127.0.0.1:6999",
	log_ipc_bind_addr_linux = "ipc://log.ipc",
	--db进程ipc地址
	db_ipc_bind_addr_win = "tcp://127.0.0.1:6998",
	db_ipc_bind_addr_linux = "ipc://db.ipc",
	--战斗进程
	battle_ipc_bind_addr_win = "tcp://127.0.0.1:6997",
	battle_ipc_bind_addr_linux = "ipc://battle.ipc",
}
