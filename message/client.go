package message

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
)

// dial ip
func Dial(address string) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Printf("dial %s error, err: %s \n", address, err)
	}

	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(conn)

	sendConnect(encoder)

	msgChan := make(chan Message)

	go func() {
		handleMessage(msgChan)
	}()

	go func() {
		for {
			HandleRequest(conn, decoder, encoder, msgChan)
		}
	}()

	for {
		// read in input from stdin
		reader := bufio.NewReader(os.Stdin)
		//	fmt.Print("Text to send: ")
		text, _ := reader.ReadString('\n')
		// send to peer
		sendMessage(text, encoder)
	}
}
