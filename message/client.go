package message

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
)

// dial ip
func Dial(address string, session *Session) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Printf("dial %s error, err: %s \n", address, err)
		return errors.New("dial error")
	}

	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(conn)

	sendConnect(encoder)

	go func() {
		for {
			session.HandleRequest(conn, decoder, encoder)
		}
		conn.Close()
	}()

	go session.handleSendMessage(conn, encoder)
	return nil
}
