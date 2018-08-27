package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

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
	shell := ishell.New()

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
			conn, err := msg.Dial(addr)
			if err != nil {
				c.Print("dial tcp address %s error, %s", addr, err)
			}

			session := msg.InitSession(username, conn)
			session.SendConnect()
			keepReading := true
			go session.ServeConn()
			go func() {
				for {
					received := <-session.ReceivedMsg
					if received == "exit" {
						c.Printf("%s quit session, you could start another call...\n", session.Username)
						keepReading = false
						return
					}
					c.Print(received)
				}
			}()

			for keepReading {
				txt := c.ReadLine()
				session.InputMsg <- txt
				if txt == "exit" {
					break
				}
			}
		},
	})

	// register a function for "wait" command.
	shell.AddCmd(&ishell.Cmd{
		Name: "wait",
		Help: "wait for another dial",
		Func: func(c *ishell.Context) {
			serve, err := net.Listen("tcp", ":"+strconv.Itoa(servePort))
			handleErr(err)
			c.Printf("listen to port : %d \n", servePort)
			keepWaiting := true
			// server side wait for connect
			var session *msg.Session
			go func() {
				for keepWaiting {
					conn, err := serve.Accept()
					handleErr(err)
					session = msg.InitSession(username, conn)
					go session.ServeConn()
					// wait for receive msg
					go func() {
						for {
							received := <-session.ReceivedMsg
							if received == "exit" {
								c.Printf("%s quit session, keep waiting for others call...\n", session.Username)
								return
							}
							c.Print(received)
						}
					}()
				}
			}()

			// wait for input msg
			for {
				//c.Print("you: ")
				txt := c.ReadLine()
				session.InputMsg <- txt
				if txt == "exit" {
					time.Sleep(2 * time.Second)
					keepWaiting = false
					session.Close()
					serve.Close()
					break
				}
			}
		},
	})

	// run shell
	shell.Run()
}
