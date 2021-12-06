// Copyright © 2021-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package codes

import (
	"testing"
)

func TestCode(t *testing.T) {
	tests := []struct {
		code Code
		name string
	}{
		{Unknown, "UNKNOWN"},
		{ServiceMaintenance, "SERVICE_MAINTENANCE"},
		{RequestTimeout, "REQUEST_TIMEOUT"},
	}

	for _, tc := range tests {
		var name = GetCodeName(tc.code)
		var val = GetCodeValue(tc.name)
		if name != tc.name || val != tc.code {
			t.Fatalf("%v", tc.code)
		}
	}
}
