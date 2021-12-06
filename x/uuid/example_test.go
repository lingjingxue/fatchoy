// Copyright © 2020-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package uuid

import (
	"testing"
)

func TestExampleUUID(t *testing.T) {
	var store = createCounterStorage("redis", "/uuid/eg11")
	if err := Init(0, store); err != nil {
		t.Fatalf("%v", err)
	}
	t.Logf("seq id: %d", NextID())
	t.Logf("uuid: %d", NextUUID())
	t.Logf("guid: %s", NextGUID())
}
