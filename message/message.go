package message

import (
	"encoding/json"
	"log"
	"net"
	"sync"
)

/**
*  send and receive message
**/
const (
	MessageType_Connect    = 0
	MessageType_Connected  = 1
	MessageType_Disconnect = 2
	MessageType_Private    = 3
	MessageType_Recieved   = 4
	MessageType_Error      = 5
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

//HandleRquest: handle request from client to server
func HandleRequest(conn net.Conn, decoder *json.Decoder, encoder *json.Encoder) {
	var msg Message
	decoder.Decode(&msg)
	receiveMessage(&msg, conn, encoder)
	return
}

func ServConn(conn net.Conn) {
	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(conn)
	for {
		HandleRequest(conn, decoder, encoder)
	}
	// close until interupt
	conn.Close()
}

func (msg *Message) sendMessage(receiver string) {
	peerConnection := listConnections[receiver]
	enc := json.NewEncoder(peerConnection)
	enc.Encode(msg)
}

func receiveMessage(msg *Message, conn net.Conn, encoder *json.Encoder) {
	switch msg.Type {
	case MessageType_Connect:
		handleNewConnect(*msg, conn, encoder)
	case MessageType_Disconnect:
		disconnect(*msg)
	case MessageType_Private:
		received := Message{
			MessageType_Recieved,
			username,
			"",
		}
		encoder.Encode(received)
	default:
		log.Println("unrecongnized type: %d", msg.Type)
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
		response.MsgContent = "Username already taken, choose another one that is not in the list"
		encoder.Encode(response)
		return false
	}
	mutex.Lock()
	listIPs[msg.Username] = conn.RemoteAddr().(*net.TCPAddr).IP.String()
	listConnections[msg.Username] = conn
	mutex.Unlock()
	log.Println(listConnections)
	response.Type = MessageType_Connected
	response.Username = username
	encoder.Encode(response)
	return true
}

//disconnect user by deleting him/her from list
func disconnect(msg Message) {
	mutex.Lock()
	delete(listIPs, msg.Username)
	delete(listConnections, msg.Username)
	mutex.Unlock()
	// update list
}
