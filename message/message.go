package message

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
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

type Session struct {
	Myname      string
	username    string
	ReceivedMsg chan string
	InputMsg    chan string
	Connection  *Connection
}

type Connection struct {
	conn    net.Conn
	encoder *json.Encoder
	decoder *json.Decoder
}

// dial ip
func Dial(address string) (conn net.Conn, err error) {
	conn, err = net.Dial("tcp", address)
	return
}

func (msg *Message) desp() string {
	return fmt.Sprintf("type: %d, username: %s, content: %s", msg.Type, msg.Username, msg.MsgContent)
}

func InitSession(name string, conn net.Conn) *Session {
	connection := &Connection{
		conn:    conn,
		encoder: json.NewEncoder(conn),
		decoder: json.NewDecoder(conn),
	}
	session := &Session{
		Myname:      name,
		ReceivedMsg: make(chan string),
		InputMsg:    make(chan string),
		Connection:  connection,
	}
	username = name
	return session
}

func (session *Session) ServeConn(conn net.Conn) {
	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(conn)
	go func() {
		for {
			session.HandleRequest(conn, decoder, encoder)
		}
	}()

	go session.handleSendMessage(conn, encoder)
}

//HandleRquest: handle request from client to server
func (session *Session) HandleRequest(conn net.Conn, decoder *json.Decoder, encoder *json.Encoder) {
	var msg Message
	decoder.Decode(&msg)
	session.receiveMessage(&msg, conn, encoder)
}

func (session *Session) handleReceivedMessage(msg Message) {
	session.ReceivedMsg <- fmt.Sprintf("%s: %s \n", msg.Username, msg.MsgContent)
}

func (session *Session) handleSendMessage(conn net.Conn, encoder *json.Encoder) {
	for {
		txt := <-session.InputMsg
		if txt == "exit" {
			session.ReceivedMsg <- fmt.Sprintf("you exit session")
			disconMsg := Message{
				MessageType_Disconnect,
				username,
				"",
			}
			encoder.Encode(disconMsg)
			conn.Close()
			break
		}
		sendMessage(txt, encoder)
	}
}

func sendMessage(txt string, encoder *json.Encoder) {
	msg := Message{MessageType_Private, username, txt}
	encoder.Encode(msg)
}

func handleError(msg *Message) {
	log.Print(msg.desp())
}

func (session *Session) receiveMessage(msg *Message, conn net.Conn, encoder *json.Encoder) {
	switch msg.Type {
	case MessageType_Error:
		handleError(msg)
	case MessageType_Connect:
		session.handleNewConnect(*msg, conn, encoder)
	case MessageType_Connected:
		session.addConnect(*msg, conn)
	case MessageType_Disconnect:
		session.disconnect(*msg, conn)
	case MessageType_Recieved:
	//	fmt.Printf("%s received \n", msg.Username)
	case MessageType_Private:
		//fmt.Println("receive private msg", msg.desp())
		session.handleReceivedMessage(*msg)
		received := Message{
			MessageType_Recieved,
			username,
			"",
		}
		encoder.Encode(received)
	default:
		if msg.Type != 0 {
			fmt.Printf("unrecongnized type: %d \n", msg.Type)
		}
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

func (session *Session) handleNewConnect(msg Message, conn net.Conn, encoder *json.Encoder) bool {
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
	fmt.Printf("new connected request from username: %s. ip: %s \n", msg.Username, listIPs[msg.Username])
	response.Type = MessageType_Connected
	response.Username = username
	encoder.Encode(response)
	return true
}

func (session *Session) SendConnect() {
	fmt.Println("send connect, username: ", username)
	msg := Message{MessageType_Connect, username, ""}
	session.Connection.encoder.Encode(msg)
}

func (session *Session) addConnect(msg Message, conn net.Conn) {
	mutex.Lock()
	listIPs[msg.Username] = conn.RemoteAddr().(*net.TCPAddr).IP.String()
	listConnections[msg.Username] = conn
	log.Printf("connected confirm from username: %s. ip: %s \n", msg.Username, listIPs[msg.Username])
	mutex.Unlock()
	session.ReceivedMsg <- "connected"
}

//disconnect user by deleting him/her from list
func (session *Session) disconnect(msg Message, conn net.Conn) {
	// update list
	session.ReceivedMsg <- fmt.Sprintf("%s exit session, session close.", msg.Username)
	mutex.Lock()
	delete(listIPs, msg.Username)
	delete(listConnections, msg.Username)
	mutex.Unlock()
	conn.Close()
}
