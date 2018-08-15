package message

import (
	"encoding/json"
	"fmt"
	"net"

	ishell "gopkg.in/abiosoft/ishell.v2"
)

// dial ip
func Dial(address string, c *ishell.Context) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Printf("dial %s error, err: %s \n", address, err)
	}

	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(conn)

	sendConnect(encoder)

	msgChan := make(chan Message)

	go func() {
		handleMessage(msgChan, c)
	}()

	go func() {
		for {
			HandleRequest(conn, decoder, encoder, msgChan)
		}
	}()

	for {
		c.Print("you: ")
		txt := c.ReadLine()
		sendMessage(txt, encoder)
		if txt == "exit" {
			break
		}
		c.Println("Hello", txt)
	}
}
