package util

import (
	"fmt"
	"testing"
)

func TestNode(t *testing.T) {
	InitialNodeList()
	UpdateNode("kino", "127.0.0.1:2500")

	node := GetNode("kino")
	assertEqual(t, node.Address, "127.0.0.1:2500", "")
	list := GetNodeList()
	assertEqual(t, list[0], node, "")

	UpdateNode("kino1", "127.0.0.1:2800")
	list = GetNodeList()
	assertEqual(t, len(list), 2, "")

	reloadNodeList()

	assertEqual(t, len(list), 2, "")
	node = GetNode("kino1")

	assertEqual(t, node.Address, "127.0.0.1:2800", "")
	ClearNodeList()
	list = GetNodeList()
	assertEqual(t, len(list), 0, "")
}

func assertEqual(t *testing.T, a interface{}, b interface{}, message string) {
	if a == b {
		return
	}
	if len(message) == 0 {
		message = fmt.Sprintf("%v != %v", a, b)
	}
	t.Fatal(message)
}
