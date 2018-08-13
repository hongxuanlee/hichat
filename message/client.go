package message

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

type Client {
     
}

func connectPeer(ip string, encoder json.Encoder) bool {
   
}

// dial ip
func Dial(address string) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Printf("dial %s error, err: %s", address, err)
	}
  encoder := json.NewEncoder(conn)
  decoder := json.NewDecoder(conn)
	for {
		// read in input from stdin
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Text to send: ")
		text, _ := reader.ReadString('\n')
		// send to socket
		fmt.Fprintf(conn, text+"\n")
		// listen for reply
		message, _ := bufio.NewReader(conn).ReadString('\n')
		fmt.Print("Message from server: " + message)
	}
}
