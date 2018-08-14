package main

import (
	"fmt"
	"net"
	"os"
	"strconv"

	msg "github.com/hongxuanlee/hichat/message"
)

const PORT = 2500

func handleErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	if len(os.Args) <= 2 || len(os.Args[2]) == 0 {
		fmt.Println("username required....")
		os.Exit(1)
	}
	username := os.Args[2]
	msg.InitUsername(username)
	var servePort int
	var err error
	if len(os.Args) > 3 {
		servePort, err = strconv.Atoi(os.Args[3])
		if err != nil {
			fmt.Printf("input serve port wrong: %s", err)
		}
	} else {
		servePort = PORT
	}

	serve, err := net.Listen("tcp", ":"+strconv.Itoa(servePort))
	handleErr(err)
	fmt.Printf("listen to port : %d \n", servePort)
	// server side

	//defer serve.Close()

	// whether dial other peer
	if len(os.Args) > 4 {
		go func() {
			for {
				conn, err := serve.Accept()
				handleErr(err)
				msg.ServeConn(conn)
			}
		}()
		dialAddr := os.Args[4]
		msg.Dial(dialAddr)
	} else {
		for {
			conn, err := serve.Accept()
			handleErr(err)
			msg.ServeConn(conn)
		}
	}
}
