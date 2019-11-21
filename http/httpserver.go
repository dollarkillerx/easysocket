package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"strings"
)

func main() {
	list, e := net.Listen("tcp", ":8080")
	if e != nil {
		log.Fatalln(e)
	}

	// 監控請求
	for {
		conn, e := list.Accept()
		if e != nil {
			log.Fatalln(e)
		}
		go handleClientRequest(conn)
		select {}
	}
}

// 處理每一次請求
func handleClientRequest(coon net.Conn) {
	var b [1024]byte
	n, err := coon.Read(b[:])
	if err != nil {
		log.Println(err)
		return
	}
	var method, host, address string
	fmt.Sscanf(string(b[:bytes.IndexByte(b[:], '\n')]), "%s%s", &method, &host)
	hostPortUrl, err := url.Parse(host)
	if err != nil {
		log.Println(err)
		return
	}

	// 進一步解析
	if hostPortUrl.Opaque == "443" {
		// https
		address = hostPortUrl.Host + ":443"
	} else {
		// http
		if strings.Index(hostPortUrl.Host, ":") == -1 {
			address = hostPortUrl.Host + ":80"
		} else {
			address = hostPortUrl.Host
		}
	}

	// 開始撥號
	server, err := net.Dial("tcp", address)
	if err != nil {
		log.Println(err)
		return
	}
	if method == "CONNECT" {
		fmt.Fprint(coon, "HTTP/1.1 200 Connection established\r\n\r\n")
	} else {
		server.Write(b[:n])
	}

	// 進行轉發
	go io.Copy(server, coon)
	io.Copy(coon, server)
}
