package main

import (
	"bufio"
	"io"
	"log"
	"os"
	"strings"
)

var confpath = "config.lua"
var g_conf = make(map[string]string)

func ParseConf() {
	file, err := os.Open(confpath)
	if err != nil {
		log.Fatalf("%v", err.Error())
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	for {
		lineBytes, _, err := reader.ReadLine()
		if lineBytes != nil && err == nil {
			line := string(lineBytes)
			line = strings.TrimSpace(line)
			if len(line) > 0 {
				if strings.HasPrefix(line, "return") {
					continue
				}
				if strings.Contains(line, "--") || strings.Contains(line, "{") || strings.Contains(line, "}") {
					continue
				}

				sps1 := strings.Split(line, ",")
				sps2 := strings.Split(sps1[0], "=")
				key := sps2[0]
				key = strings.TrimSpace(key)
				val := sps2[1]
				val = strings.TrimSpace(val)
				g_conf[key] = val
				//println(key, val)
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("%v", err.Error())
		}
	}
}
