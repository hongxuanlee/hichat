package main

import (
	"fmt"
	"net"
	"os"
	"strconv"

	msg "github.com/hongxuanlee/hichat/message"
	ishell "gopkg.in/abiosoft/ishell.v2"
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

	// create new shell.
	// by default, new shell includes 'exit', 'help' and 'clear' commands.
	shell := ishell.New()

	// display welcome info.
	shell.Println("I am the cutest chatter")

	// register a function for "call" command.
	shell.AddCmd(&ishell.Cmd{
		Name: "call",
		Help: "call addr(x.x.x.x:2500)",
		Func: func(c *ishell.Context) {
			if len(c.Args) < 0 {
				c.Println("call addr required")
			}
			addr := c.Args[0]
			msg.Dial(addr, c)
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "wait",
		Help: "wait for another dial",
		Func: func(c *ishell.Context) {
			serve, err := net.Listen("tcp", ":"+strconv.Itoa(servePort))
			handleErr(err)
			c.Printf("listen to port : %d \n", servePort)
			// server side
			for {
				conn, err := serve.Accept()
				handleErr(err)
				txtChan := make(chan string)
				msg.ServeConn(conn, c, txtChan)
				c.Print("you: ")
				txt := c.ReadLine()
				txtChan <- txt
				if txt == "exit" {
					break
				}
			}
		},
	})

	// run shell
	shell.Run()
}
