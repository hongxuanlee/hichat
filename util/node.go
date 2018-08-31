package util

import (
	"bufio"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// nodeList:
type Node struct {
	Username        string
	Address         string
	LastConnectTime int64
}

var nodeFilePath = os.Getenv("HOME") + "/.hichat_nodes"
var chatNodeList []*Node
var chatNodeMap = map[string]*Node{}

// line ("kino 0.0.0.0:2500 1535684617935") => Node
func generateNodeByLine(line string) *Node {
	vals := strings.Split(line, " ")
	if len(vals) < 2 {
		return nil
	}
	username := strings.TrimSpace(vals[0])
	addr := strings.TrimSpace(vals[1])
	node := &Node{Username: username, Address: addr}
	if len(vals) >= 3 {
		val2 := strings.TrimSpace(vals[2])
		lastTime, err := strconv.ParseInt(val2, 10, 64)
		if err != nil {
			log.Fatalf("lastConnectTime parse error %s", err)
		} else {
			node.LastConnectTime = lastTime
		}
	}
	return node
}

func nodeToLine(node *Node) string {
	strs := []string{
		node.Username,
		node.Address,
		strconv.FormatInt(node.LastConnectTime, 10),
	}
	return strings.Join(strs, " ")
}

func getNodelistFromFile() (nodeList []*Node, err error) {
	fh, err := os.Open(nodeFilePath)
	if err != nil {
		return
	}
	defer fh.Close()
	scanner := bufio.NewScanner(fh)
	for scanner.Scan() {
		line := scanner.Text()
		node := generateNodeByLine(line)
		if node != nil {
			nodeList = append(nodeList, node)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return
}

func writeNodeListToFile() (err error) {
	lines := make([]string, len(chatNodeList))
	for i, node := range chatNodeList {
		lines[i] = nodeToLine(node)
	}
	content := strings.Join(lines, "\n")
	err = ioutil.WriteFile(nodeFilePath, []byte(content), 0644)
	return
}

func ClearNodeList() (err error) {
	chatNodeList = []*Node{}
	chatNodeMap = map[string]*Node{}
	err = writeNodeListToFile()
	return
}

func GetNode(username string) *Node {
	if v, ok := chatNodeMap[username]; ok {
		return v
	}
	return nil
}

func listToMap(nodes []*Node) map[string]*Node {
	var m = make(map[string]*Node, len(nodes))
	for _, node := range nodes {
		m[node.Username] = node
	}
	return m
}

func mapToList(m map[string]*Node) []*Node {
	nodeList := make([]*Node, len(m))
	i := 0
	for _, node := range m {
		nodeList[i] = node
		i++
	}
	return nodeList
}

func reloadNodeList() (err error) {
	nodeList, err := getNodelistFromFile()
	if err == nil {
		chatNodeList = nodeList
		chatNodeMap = listToMap(chatNodeList)
	}
	return
}

func InitialNodeList() {
	err := reloadNodeList()
	if err != nil {
		log.Printf("no initial nodelist: %s", err)
	}
}

func GetNodeList() []*Node {
	if len(chatNodeList) == 0 {
		reloadNodeList()
	}
	return chatNodeList
}

func GetNodeListMap() map[string]*Node {
	if len(chatNodeMap) == 0 {
		reloadNodeList()
	}
	return chatNodeMap
}

// updateNode infomation
func UpdateNode(username string, addr string) (err error) {
	now := time.Now().Unix()
	if v, ok := chatNodeMap[username]; ok {
		v.Address = addr
		v.LastConnectTime = now
	} else {
		node := &Node{username, addr, now}
		chatNodeMap[username] = node
	}
	chatNodeList = mapToList(chatNodeMap)
	err = writeNodeListToFile()
	return
}
