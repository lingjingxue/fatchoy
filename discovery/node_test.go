// Copyright © 2021-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package discovery

import (
	"testing"
)

func TestNewNode(t *testing.T) {
	var node = NewNode("GATE", 1)
	t.Logf("%v", node)
}

func TestNodeMap(t *testing.T) {
	var nm = NewNodeMap()
	nm.InsertNode(NewNode("GATE", 1))
	nm.InsertNode(NewNode("GATE", 2))
	nm.InsertNode(NewNode("GAME", 1))
	nm.InsertNode(NewNode("GAME", 2))
	nm.InsertNode(NewNode("LOGIN", 1))
	nm.InsertNode(NewNode("LOGIN", 2))
	t.Logf("initial nodes: %v", nm.String())

	var r = nm.GetNodes("GATE")
	t.Logf("get nodes: %v", r)

	nm.DeleteNode("LOGIN", 1)
	t.Logf("after del 1: %v", nm.String())

}
