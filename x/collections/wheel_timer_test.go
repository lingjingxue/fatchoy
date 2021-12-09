// Copyright © 2020 simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package collections

import (
	"testing"
)

func TestNewHashedWheelTimer(t *testing.T) {
	var timer = NewHashedWheelTimer()
	timer.Start()
	defer timer.Shutdown()
}
