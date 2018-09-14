package message

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
)

// Session_Signal
const (
	Session_Signal_Exit = -1
)

/**
*  Send and receive message
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
	Username    string
	ReceivedMsg chan string
	InputMsg    chan string
	CurConn     *Connection
	Signal      chan int
	close       bool
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
		CurConn:     connection,
	}
	username = name
	return session
}

func (session *Session) ServeConn() {
	go func() {
		for !session.close {
			session.HandleRequest()
		}
	}()

	go session.handleSendMessage()
}

func (session *Session) GetCurconn() (curConn *Connection, err error) {
	if session.CurConn != nil {
		curConn = session.CurConn
		return
	}
	fmt.Println("connection lose, exit session")
	//TODO retry logic in some situation
	session.Close()
	err = errors.New("session close")
	return
}

// Close Session
func (session *Session) Close() {
	if session.close {
		return
	}
	session.close = true
	// close conn
	if session.CurConn != nil {
		session.CurConn.conn.Close()
	}
	session.Signal <- Session_Signal_Exit
}

//HandleRquest: handle request from client to server
func (session *Session) HandleRequest() {
	var msg Message
	curConn, err := session.GetCurconn()
	if err != nil {
		return
	}
	decoder := curConn.decoder
	decoder.Decode(&msg)
	session.receiveMessage(&msg)
}

func (session *Session) handleReceivedMessage(msg Message) {
	Notify(msg.MsgContent)
	session.ReceivedMsg <- fmt.Sprintf("%s: %s \n", msg.Username, msg.MsgContent)
}

func (session *Session) handleSendMessage() {
	curConn, err := session.GetCurconn()
	if err != nil {
		fmt.Println("lose Connection")
		return
	}
	encoder := curConn.encoder
	for !session.close {
		txt := <-session.InputMsg
		session.sendMessage(txt)
		if txt == "exit" {
			session.ReceivedMsg <- fmt.Sprintf("you exit session\n")
			disconMsg := Message{
				MessageType_Disconnect,
				username,
				"",
			}
			encoder.Encode(disconMsg)
			session.Close()
			break
		}
	}
}

func (session *Session) sendMessage(txt string) {
	curConn, err := session.GetCurconn()
	encoder := curConn.encoder
	if err != nil {
		fmt.Println("lose Connection")
		return
	}
	msg := Message{MessageType_Private, username, txt}
	encoder.Encode(msg)
}

func handleError(msg *Message) {
	log.Print(msg.desp())
}

func (session *Session) receiveMessage(msg *Message) {
	curConn, err := session.GetCurconn()
	encoder := curConn.encoder
	if err != nil {
		fmt.Println("lose Connection")
		return
	}
	switch msg.Type {
	case MessageType_Error:
		handleError(msg)
	case MessageType_Connect:
		session.handleNewConnect(*msg)
	case MessageType_Connected:
		session.addConnect(*msg)
	case MessageType_Disconnect:
		session.disconnect(*msg)
	case MessageType_Recieved:
	//	fmt.Printf("%s received \n", msg.Username)
	case MessageType_Private:
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

func (session *Session) handleNewConnect(msg Message) bool {
	curConn, err := session.GetCurconn()
	if err != nil {
		fmt.Println("lose Connection")
		return false
	}
	encoder, conn := curConn.encoder, curConn.conn
	response := Message{}

	if userExist(msg.Username) {
		response.Type = MessageType_Error
		response.Username = username
		response.MsgContent = "Username already taken"
		encoder.Encode(response)
		return false
	}
	session.Username = msg.Username
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
	curConn, err := session.GetCurconn()
	if err != nil {
		fmt.Println("lose Connection")
		return
	}
	encoder := curConn.encoder
	msg := Message{MessageType_Connect, username, ""}
	encoder.Encode(msg)
}

func (session *Session) addConnect(msg Message) {
	fmt.Println("add connect")
	curConn, err := session.GetCurconn()
	if err != nil {
		fmt.Println("lose Connection")
		return
	}
	conn := curConn.conn
	session.Username = msg.Username
	mutex.Lock()
	listIPs[msg.Username] = conn.RemoteAddr().(*net.TCPAddr).IP.String()
	listConnections[msg.Username] = conn
	fmt.Printf("connected confirm from username: %s. ip: %s \n", msg.Username, listIPs[msg.Username])
	mutex.Unlock()
}

//disconnect user by deleting him/her from list
func (session *Session) disconnect(msg Message) {
	// update list
	session.ReceivedMsg <- "exit"
	mutex.Lock()
	delete(listIPs, msg.Username)
	delete(listConnections, msg.Username)
	mutex.Unlock()
	session.Close()
}
