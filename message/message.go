package message

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"

	ishell "gopkg.in/abiosoft/ishell.v2"
)

/**
*  send and receive message
**/
const (
	MessageType_Connect    = 1
	MessageType_Connected  = 2
	MessageType_Disconnect = 3
	MessageType_Private    = 4
	MessageType_Recieved   = 5
	MessageType_Error      = 6
)

const BUFFER_SIZE = 4096

var (
	output          chan string         = make(chan string)         //channel waitin on the user to type something
	listIPs         map[string]string   = make(map[string]string)   // username -> ip
	listConnections map[string]net.Conn = make(map[string]net.Conn) // username -> conn
	username        string
	ip              string
	mutex           = new(sync.Mutex)
)

type Message struct {
	Type       int // type of message ("CONNECT","PRIVATE","DISCONNECT")
	Username   string
	MsgContent string
}

func (msg *Message) desp() string {
	return fmt.Sprintf("type: %d, username: %s, content: %s", msg.Type, msg.Username, msg.MsgContent)
}

//HandleRquest: handle request from client to server
func HandleRequest(conn net.Conn, decoder *json.Decoder, encoder *json.Encoder, c chan Message) {
	var msg Message
	decoder.Decode(&msg)
	receiveMessage(&msg, conn, encoder, c)
}

func InitUsername(name string) {
	fmt.Println("Myname is", name)
	username = name
}

func ServeConn(conn net.Conn, c *ishell.Context, sendTxt chan string) {
	msgChan := make(chan Message)
	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(conn)
	go func() {
		for {
			HandleRequest(conn, decoder, encoder, msgChan)
		}
	}()

	go func() {
		handleMessage(msgChan, c)
	}()

	for {
		txt := <-sendTxt
		if txt == "exit" {
			break
		}
		sendMessage(txt, encoder)
	}

	conn.Close()
}

func handleMessage(c chan Message, ctx *ishell.Context) {
	var err error
	var received Message
	for err == nil {
		received = <-c
		ctx.Printf("%s: %s \n", received.Username, received.MsgContent)
	}
}

func sendMessage(txt string, encoder *json.Encoder) {
	msg := Message{MessageType_Private, username, txt}
	encoder.Encode(msg)
}

func handleError(msg *Message) {
	log.Print(msg.desp())
}

func receiveMessage(msg *Message, conn net.Conn, encoder *json.Encoder, c chan Message) {
	switch msg.Type {
	case MessageType_Error:
		handleError(msg)
	case MessageType_Connect:
		handleNewConnect(*msg, conn, encoder)
	case MessageType_Connected:
		addConnect(*msg, conn)
	case MessageType_Disconnect:
		disconnect(*msg, conn)
	case MessageType_Recieved:
	//	fmt.Printf("%s received \n", msg.Username)
	case MessageType_Private:
		//fmt.Println("receive private msg", msg.desp())
		c <- *msg
		received := Message{
			MessageType_Recieved,
			username,
			"",
		}
		encoder.Encode(received)
	default:
		log.Printf("unrecongnized type: %d \n", msg.Type)
	}
}

func userExist(username string) bool {
	for k, _ := range listIPs {
		if k == username {
			return true
		}
	}
	return false
}

func handleNewConnect(msg Message, conn net.Conn, encoder *json.Encoder) bool {
	response := Message{}
	if userExist(msg.Username) {
		response.Type = MessageType_Error
		response.Username = username
		response.MsgContent = "Username already taken"
		encoder.Encode(response)
		return false
	}
	mutex.Lock()
	listIPs[msg.Username] = conn.RemoteAddr().(*net.TCPAddr).IP.String()
	listConnections[msg.Username] = conn
	mutex.Unlock()
	log.Printf("new connected request from username: %s. ip: %s \n", msg.Username, listIPs[msg.Username])
	response.Type = MessageType_Connected
	response.Username = username
	encoder.Encode(response)
	return true
}

func sendConnect(encoder *json.Encoder) {
	fmt.Println("send connect, username: ", username)
	msg := Message{MessageType_Connect, username, ""}
	encoder.Encode(msg)
}

func addConnect(msg Message, conn net.Conn) {
	mutex.Lock()
	listIPs[msg.Username] = conn.RemoteAddr().(*net.TCPAddr).IP.String()
	listConnections[msg.Username] = conn
	log.Printf("connected confirm from username: %s. ip: %s \n", msg.Username, listIPs[msg.Username])
	mutex.Unlock()
}

//disconnect user by deleting him/her from list
func disconnect(msg Message, conn net.Conn) {
	// update list
	mutex.Lock()
	delete(listIPs, msg.Username)
	delete(listConnections, msg.Username)
	mutex.Unlock()
	conn.Close()
}
