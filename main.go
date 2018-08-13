package main

import (
	"fmt"
	"net"
	"os"

	msg "github.com/hongxuanlee/hichat/message"
)

const PORT = "2500"

func handleErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	username = os.Args[2]
	// server side
	serve, err := net.Listen("tcp", ":"+PORT)
	handleErr(err)
	defer serve.Close()
	fmt.Println("listening on :" + PORT)
	for {
		conn, err := serve.Accept()
		handleErr(err)
		msg.ServeConn(conn)
	}
}
