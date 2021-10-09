// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package collections

import (
	"testing"
)

func checkDupObject(zsl *ZSkipList, t *testing.T) {
	if zsl.Len() == 0 {
		return
	}
	var set = make(map[int64]bool, zsl.Len())
	var rank = zsl.Len()
	var node = zsl.HeadNode().Next()
	for node != nil {
		rank--
		var player = node.Ele.(*testPlayer)
		if _, found := set[player.Uid]; found {
			t.Fatalf("Duplicate rank object found: %d, %d", rank, player.Uid)
		}
		set[player.Uid] = true
		node = node.Next()
	}
}

func TestZSkipList(t *testing.T) {

}
