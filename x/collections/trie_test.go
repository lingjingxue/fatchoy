// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

//go:build !ignore

package collections

import (
	"testing"
)

func createProfanityTrie(dict []string) *HashTrie {
	var trie = NewHashTrie()
	for _, word := range dict {
		trie.AddWord(word)
	}
	return trie
}

// 测试删除字段
func TestTrie_Remove(t *testing.T) {
	var dirtyWords = []string{
		"fuck",
		"fuckyou",
		"shit",
		"pussy",
		"dick",
		"eatdick",
	}

	var trie = createProfanityTrie(dirtyWords)
	println(trie.String())
	var toDel = []string{
		"fuck",
		"fuckyou",
		"eatdick",
		"dick",
	}
	for _, word := range toDel {
		ok := trie.Remove(word)
		if !ok {
			t.Fatalf("remove failed")
		}
	}
	trie = createProfanityTrie(dirtyWords)
	toDel = []string{
		"fuckyou",
		"fuck",
		"dick",
		"eatdick",
	}
	for _, word := range toDel {
		ok := trie.Remove(word)
		if !ok {
			t.Fatalf("remove failed")
		}
	}
}

// 测试精确匹配
func TestTrie_ExactMatch(t *testing.T) {
	var dirtyWords = []string{
		"fuck",
		"shit",
		"我操",
		"*你妈",
		"你妈*",
	}
	tests := []struct {
		input    string
		expected bool
	}{
		{"", false},
		{"fuck", true},
		{"fucku", false},
		{"ufuck", false},
		{"我操", true},
		{"我操她", false},
		{"哎哟我操", false},
		{"哎哟我操", false},
		{"曰你妈", true},
		{"你妈B", true},
	}
	var trie = createProfanityTrie(dirtyWords)
	for _, tc := range tests {
		output := trie.ExactMatch(tc.input)
		if output != tc.expected {
			t.Fatalf("unexpected output for [%s]: %v != %v", tc.input, output, tc.expected)
		}
	}
}

// 测试模糊匹配
func TestTrie_Match(t *testing.T) {
	var dirtyWords = []string{
		"sm",
		"台独",
		"*泽东",
	}
	tests := []struct {
		input    string
		expected bool
	}{
		{"", false},
		{"small", true},
		{"smith", true},
		{"台独", true},
		{"一台独立服务器", true},
		{"毛泽东", true},
		{"泽东影业", false},
	}
	var trie = createProfanityTrie(dirtyWords)
	for _, tc := range tests {
		output := trie.Contains(tc.input)
		if output != tc.expected {
			t.Fatalf("unexpected output for [%s]: %v != %v", tc.input, output, tc.expected)
		}
	}
}

// 测试*号过滤
func TestTrie_Filter(t *testing.T) {
	var dirtyWords = []string{
		"sm",
		"fuck",
		"毛泽东",
	}
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"quickfox", "quickfox"},
		{"sm", "**"},
		{"small", "**all"},
		{"jrsmith", "jr**ith"},
		{"go_fuck_u_smith", "go_****_u_**ith"},
		{"毛泽东", "***"},
		{"你好毛泽东", "你好***"},
		{"毛-泽-东", "毛-泽-东"},
	}
	var trie = createProfanityTrie(dirtyWords)
	for _, tc := range tests {
		output := trie.Filter(tc.input)
		if output != tc.expected {
			t.Fatalf("unexpected output for [%s]: %v != %v", tc.input, output, tc.expected)
		}
	}
}
