package main

import (
	zmq "github.com/pebbe/zmq4"
	"fmt"
	//"strings"
	"strconv"
	"time"
)

var g_socket *zmq.Socket
var g_sendsocks = make(map[string]*zmq.Socket)

//消息请求类型
var SEND_TYPE_REQ = "REQ"   //请求
var SEND_TYPE_RESP = "RESP" //返回请求

func StartZmq() {
	//对端zmq地址
	servPort := g_conf["servPort"]
	portInt, _ := strconv.ParseInt(servPort, 10, 32)
	tarAddr := g_conf["game_ipc_bind_addr_linux_fmt"]
	tarAddr = fmt.Sprintf(tarAddr, portInt)

	//本端地址
	selfAddr := g_conf["http_ipc_bind_addr_linux"]

	println(tarAddr, selfAddr)

	socket, _ := zmq.NewSocket(zmq.XREP)
	g_socket = socket
	defer closingAllSocks()

	timer := time.NewTicker(100 * time.Millisecond)
	for {
		select {
		case <-timer.C:
			update()
		}
	}

	timer.Stop()
}

//对应zmq的send,根据 send 的类型有对应数量的 partial msg
func recv() {
	identity, err1 := g_socket.Recv(zmq.DONTWAIT)
	if err1 {
		fmt.Printf("identity: %v\n", err1)
		return
	}
	if len(identity) == 0 { //没消息
		return
	}

	sendType, err2 := g_socket.Recv(zmq.DONTWAIT)
	if err2 {
		fmt.Printf("sendType: %v\n", err2)
		return
	}

	if sendType == SEND_TYPE_REQ {
		msgId2str, err3 := g_socket.Recv(zmq.DONTWAIT)
		if err3 {
			fmt.Printf("msgId2str: %v\n", err3)
			return
		}

		rpcFuncName, err4 := g_socket.Recv(zmq.DONTWAIT)
		if err4 {
			fmt.Printf("rpcFuncName: %v\n", err4)
			return
		}

		args_str, err5 := g_socket.Recv(zmq.DONTWAIT)
		if err5 {
			fmt.Printf("args_str: %v\n", err5)
			return
		}

		addr, err6 := g_socket.Recv(zmq.DONTWAIT)
		if err6 {
			fmt.Printf("addr: %v\n", err6)
			return
		}

		if rpcFuncName == "doFunc" {
			doFunc(args_str, addr)
		}

	} else if sendType == SEND_TYPE_RESP {
		//go这一端暂不支持rpc返回值
	}
}

func update() {
	count := 100
	for {

		if !recv() {
			break
		}

		count--
		if count <= 0 {
			break
		}
	}
}

func closingAllSocks() {
	socket.Close()
	for k, v := range g_sendsocks {
		v.Close()
	}
}

func doFunc(args_str, addr string) {
	println("doFunc:", args_str, addr)
}
