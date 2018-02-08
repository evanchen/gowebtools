package main

import (
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"strings"
	//"strconv"
	"time"
	"os"
)

var g_socket *zmq.Socket
var g_sendsocks = make(map[string]*zmq.Socket)
var g_msgId = int(0)
var selfAddr string
var enginType = "webserv"

//消息请求类型
var SEND_TYPE_REQ = "REQ"   //请求
var SEND_TYPE_RESP = "RESP" //返回请求

func StartZmq() {
	//对端zmq地址
	//servPort := g_conf["servPort"]
	//portInt, _ := strconv.ParseInt(servPort, 10, 32)
	tarAddr := g_conf["game_ipc_bind_addr_linux_fmt"]
	tarAddr = fmt.Sprintf(tarAddr, 0)
	tarAddr = strings.TrimPrefix(tarAddr,"\"")
	tarAddr = strings.TrimSuffix(tarAddr,"\"")
	//本端地址
	selfAddr = g_conf["http_ipc_bind_addr_linux"]
	selfAddr = strings.TrimPrefix(selfAddr,"\"")
	selfAddr = strings.TrimSuffix(selfAddr,"\"")
	println(tarAddr, selfAddr)

	socket, _ := zmq.NewSocket(zmq.ROUTER)
	
	g_socket = socket
	socket.Bind(selfAddr)
	defer closingAllSocks()

	//先向游戏帐号服务器注册
	send(tarAddr, "DISPATCH:RegisterWebServ", "{[1]=0}") //参数只有一个,gsId = 0

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
func recv() bool {
	identity, err1 := g_socket.Recv(zmq.DONTWAIT)
	if err1 != nil {
		//fmt.Printf("identity: %v\n", err1)
		return false
	}
	if len(identity) == 0 { //没消息
		return false
	}

	sendType, err2 := g_socket.Recv(zmq.DONTWAIT)
	if err2 != nil {
		fmt.Printf("sendType: %v\n", err2)
		return false
	}

	if sendType == SEND_TYPE_REQ {
		_, err3 := g_socket.Recv(zmq.DONTWAIT)
		if err3 != nil {
			fmt.Printf("msgId2str: %v\n", err3)
			return false
		}

		rpcFuncName, err4 := g_socket.Recv(zmq.DONTWAIT)
		if err4 != nil {
			fmt.Printf("rpcFuncName: %v\n", err4)
			return false
		}

		args_str, err5 := g_socket.Recv(zmq.DONTWAIT)
		if err5 != nil {
			fmt.Printf("args_str: %v\n", err5)
			return false
		}

		addr, err6 := g_socket.Recv(zmq.DONTWAIT)
		if err6 != nil {
			fmt.Printf("addr: %v\n", err6)
			return false
		}

		if rpcFuncName == "doFunc" {
			doFunc(args_str, addr)
		}

	} else if sendType == SEND_TYPE_RESP {
		//go这一端暂不支持rpc返回值
	}

	return true
}

//go这一端暂不支持rpc返回值
//发送方作为 dealer,不用发送identity
func send(addr, rpcFuncName, args string) {
	peerSock, ok := g_sendsocks[addr]
	if !ok {
		newSocket, err := zmq.NewSocket(zmq.DEALER)
		if err != nil {
			os.Exit(-1)
		}
		g_sendsocks[addr] = newSocket
		peerSock = newSocket
		fmt.Printf("%v, %s\n",peerSock.Connect(addr),addr)
	}

	g_msgId++
	msgId2str := fmt.Sprintf("%d_%s%d", g_msgId, enginType, 0)

	peerSock.Send(SEND_TYPE_REQ, zmq.SNDMORE)
	peerSock.Send(msgId2str, zmq.SNDMORE)
	peerSock.Send(rpcFuncName, zmq.SNDMORE)
	peerSock.Send(args, zmq.SNDMORE)
	peerSock.Send("", 0) //addr 为空,不用rpc返回
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
	g_socket.Close()
	for _, v := range g_sendsocks {
		v.Close()
	}
}

//完成对端的rpc操作,并返回操作结果
func doFunc(args_str, addr string) {
	println("doFunc:", args_str, addr)

	if len(addr) > 0 {
		send(addr, "DOCMD:HandleWebServRet", args_str)
	}
}
