package main

import (
	"encoding/json"
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"net/http"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var g_socket *zmq.Socket
var g_sendsocks = make(map[string]*zmq.Socket)
var g_msgId = int(0)
var selfAddr string
var enginType = "webserv"
var mutex sync.Mutex

//消息请求类型
var SEND_TYPE_REQ = "REQ"   //请求
var SEND_TYPE_RESP = "RESP" //返回请求

func StartZmq() {
	//对端zmq地址
	tarAddr := g_conf["game_ipc_bind_addr_linux_fmt"]
	tarAddr = fmt.Sprintf(tarAddr, 0)
	tarAddr = strings.TrimPrefix(tarAddr, "\"")
	tarAddr = strings.TrimSuffix(tarAddr, "\"")
	//本端地址
	selfAddr = g_conf["http_ipc_bind_addr_linux"]
	selfAddr = strings.TrimPrefix(selfAddr, "\"")
	selfAddr = strings.TrimSuffix(selfAddr, "\"")
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
			go doFunc(args_str, addr)
		}

	} else if sendType == SEND_TYPE_RESP {
		//go这一端暂不支持rpc返回值
	}

	return true
}

//go这一端暂不支持rpc返回值
//发送方作为 dealer,不用发送identity
func send(addr, rpcFuncName, args string) {
	defer mutex.Unlock()
	mutex.Lock()
	peerSock, ok := g_sendsocks[addr]
	if !ok {
		newSocket, err := zmq.NewSocket(zmq.DEALER)
		if err != nil {
			os.Exit(-1)
		}
		g_sendsocks[addr] = newSocket
		peerSock = newSocket
		fmt.Printf("%v, %s\n", peerSock.Connect(addr), addr)
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
	args := decode_luatable_argstr(args_str)
	reqType := args[0]
	openid := args[1]
	access_token := args[2]
	retStr := realDo(reqType,openid,access_token)
	if len(addr) > 0 {
		//需要加锁
		send(addr, "DOCMD:HandleWebServRet", retStr)
	}
}

func decode_luatable_argstr(args_str string) []string {
	var args []string
	if strings.HasPrefix(args_str, "{") { //是lua table格式转换来的str
		//格式举例, args_str = {[1]="asd",[2]="3",}
		args_str = strings.TrimPrefix(args_str, "{")
		args_str = strings.TrimSuffix(args_str, "}")
		elem := strings.Split(args_str, ",")
		for _, v := range elem {
			kv := strings.Split(v, "=")
			//去value部分
			args = append(args, kv[1])
		}
	} else {
		args = append(args, args_str)
	}
	return args
}

func encode_luatable_argstr(str ...string) string {
	num := 0
	var arr []string
	arr = append(arr, "{")
	var inside []string
	for _, v := range str {
		num++
		s := fmt.Sprintf("[%d]=%s", num, v)
		inside = append(inside, s)
	}
	if num > 0 {
		s := strings.Join(inside, ",")
		arr = append(arr, s)
	}
	arr = append(arr, "}")
	ret := strings.Join(arr, "")
	return ret
}

func realDo(reqType, openid, access_token string) string {
	var retStr string
	if reqType == "authcheck" {
		retStr = authCheck(openid, access_token)
	}
	return retStr
}

var ERR_1 = "AUTH_ERROR"
var ERR_2 = "USERINFO_ERROR"

// {
// "errcode":40003,"errmsg":"invalid openid"
// }

type authResp struct {
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
}

//access_token 认证
//https://api.weixin.qq.com/sns/auth?access_token=ACCESS_TOKEN&openid=OPENID
func authCheck(openid, access_token string) string {
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/auth?access_token=%s&openid=%s", access_token, openid)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("get error: %v\n", err)
		return ERR_1
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("body error: %v\n", err)
		return ERR_1
	}
	respBody := &authResp{}
	json.Unmarshal(body, respBody)
	if respBody.Errcode != 0 {
		fmt.Printf("auth error: %s\n", respBody.Errmsg)
		return ERR_1
	}
	retStr := getUserinfo(openid, access_token)
	return retStr
}

// {
// "openid":"OPENID",
// "nickname":"NICKNAME",
// "sex":1,
// "province":"PROVINCE",
// "city":"CITY",
// "country":"COUNTRY",
// "headimgurl": "http://wx.qlogo.cn/mmopen/g3MonUZtNHkdmzicIlibx6iaFqAc56vxLSUfpb6n5WKSYVY0ChQKkiaJSgQ1dZuTOgvLLrhJbERQQ4eMsv84eavHiaiceqxibJxCfHe/0",
// "privilege":[
// "PRIVILEGE1",
// "PRIVILEGE2"
// ],
// "unionid": " o6_bmasdasdsad6_2sgVt7hMZOPfL"

// }

type userinfoResp struct {
	Openid     string   `json:"openid"`
	Nickname   string   `json:"nickname"`
	Sex        int      `json:"sex"`
	Province   string   `json:"province"`
	City       string   `json:"city"`
	Country    string   `json:"country"`
	Headimgurl string   `json:"headimgurl"`
	Privilege  []string `json:"privilege"`
	Unionid    string   `json:"unionid"`
	Errcode    int      `json:"errcode"`
	Errmsg     string   `json:"errmsg"`
}

//获取用户信息
//https://api.weixin.qq.com/sns/userinfo?access_token=ACCESS_TOKEN&openid=OPENID
func getUserinfo(openid, access_token string) string {
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s", access_token, openid)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("get error: %v\n", err)
		return ERR_2
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("body error: %v\n", err)
		return ERR_2
	}
	respBody := &userinfoResp{}
	json.Unmarshal(body, respBody)
	if respBody.Errcode != 0 {
		fmt.Printf("auth error: %s\n", respBody.Errmsg)
		return ERR_2
	}

	var strs []string
	strs = append(strs, respBody.Openid)
	strs = append(strs, respBody.Nickname)
	strsex := strconv.Itoa(respBody.Sex)
	strs = append(strs, strsex)
	strs = append(strs, respBody.Headimgurl)
	strs = append(strs, respBody.Privilege...)
	retStr := strings.Join(strs, ",")
	return retStr
}
