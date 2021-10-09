// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package collections

import "testing"

func TestConsistentExample(t *testing.T) {
	var c = NewConsistent()
	t.Logf("add node 1 2 3 4")
	c.AddNode("node1")
	c.AddNode("node2")
	c.AddNode("node3")
	c.AddNode("node4")
	key := "key1"
	node := c.GetNodeBy(key)
	t.Logf("get node %s by %s", node, key)
	c.RemoveNode("node1")
	c.RemoveNode("node2")
	t.Logf("remove node 1 2")
	node = c.GetNodeBy(key)
	t.Logf("get node %s by %s", node, key)
}
