// Copyright © 2020-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package strutil

import (
	"reflect"
	"testing"
)

func TestParseBool(t *testing.T) {
	v := ParseBool("")
	if v {
		t.Fatalf("parse bool failure")
	}

	v = ParseBool("on")
	if !v {
		t.Fatalf("parse bool failure")
	}

	v = ParseBool("YES")
	if !v {
		t.Fatalf("parse bool failure")
	}
}

func TestParseNumber(t *testing.T) {
	var s = "1234"
	n := MustParseI32(s)
	if int(n) != 1234 {
		t.Fatalf("unexpected result: %d != 1234", n)
	}
}

func TestParseSepKeyValues(t *testing.T) {
	tests := []struct {
		input    string
		expected map[string]string
	}{
		{"", map[string]string{}},
		{"a=", map[string]string{"a": ""}},
		{"a=1,b=2,c=3", map[string]string{"a": "1", "b": "2", "c": "3"}},
		{"a=1,b,c=3", map[string]string{"a": "1", "b": "", "c": "3"}},
		{"a=1,b=", map[string]string{"a": "1", "b": ""}},
		{"a=,b=2,c,d=3,e", map[string]string{"a": "", "b": "2", "c": "", "d": "3", "e": ""}},
		{"a='1,2,3',b=456", map[string]string{"a": "1,2,3", "b": "456"}},
		{"a=123, b=456", map[string]string{"a": "123", "b": "456"}},
		{"a = 123, b = 456", map[string]string{"a": "123", "b": "456"}},
		{"a=123 , b='4,5,6' , c = 789", map[string]string{"a": "123", "b": "4,5,6", "c": "789"}},
	}
	const sep1, sep2 = ',', '='
	for i, test := range tests {
		r := ParseKeyValuePairs(test.input, sep1, sep2)
		if len(r) == 0 && len(test.expected) == 0 {
			continue
		}
		if !reflect.DeepEqual(r, test.expected) {
			t.Fatalf("Test case %d failed, expect %v, got %v", i, test.expected, r)
		}
	}
}
