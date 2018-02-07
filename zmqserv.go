package main

import (
	zmq "github.com/pebbe/zmq4"
	"fmt"
	"strings"
	"strconv"
)

func worker() {
	//对端zmq地址
	servPort := g_conf["servPort"]
	portInt,_ := strconv.ParseInt(servPort, 10, 32)
	tarAddr := g_conf["game_ipc_bind_addr_linux_fmt"]
	tarAddr = fmt.Sprint(tarAddr,portInt)

	//本端地址
	selfAddr := g_conf["http_ipc_bind_addr_linux"]

	w, _ := zmq.NewSocket(zmq.XREQ)
	defer worker.Close()
	worker.Connect(tarAddr)

	
}

func StartZmq() {

}
