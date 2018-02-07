package main

import (
	"fmt"
	"net/http"
	"log"
	"sync"
)

var wg sync.WaitGroup

func reqSecret(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "asdo131asdn123-089")
}

func main() {
	ParseConf()
	//http服务
	wg.Add(2)
	go func() {
		defer wg.Done()

		http.HandleFunc("/reqSecret", reqSecret)
		err := http.ListenAndServe(":9090", nil)
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}()

	//开始zmq定时循环
	go func () {
		defer wg.Done()
		StartZmq()
	}()

	wg.Wait()
}
