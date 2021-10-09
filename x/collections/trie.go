// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package collections

import (
	"strings"
)

// 用trie树实现敏感词过滤
// 	精确匹配：比如精确屏蔽【sm】，就只会挡【sm】，不会挡【small】
//  模糊匹配：比如模糊屏蔽【习主席】，输入【你好习主席】时，会挡【你好习主席】这个词，但是不会挡【你好习-主-席】

// 通配符
const WildCardStar = '*' // '\u002A'

// 哈希表实现的字典树
type HashTrie struct {
	root trieNode
	size int
}

func NewHashTrie() *HashTrie {
	return &HashTrie{}
}

type trieNode struct {
	children map[rune]*trieNode //
	isEnd    bool               // 是否单词结束
}

func (n *trieNode) reset() {
	n.children = make(map[rune]*trieNode)
	n.isEnd = false
}

func (n *trieNode) lazyInit() {
	if n.children == nil {
		n.children = make(map[rune]*trieNode)
	}
}

// 是否包含`r`和通配符
func (n *trieNode) contains(r rune) *trieNode {
	if child, found := n.children[r]; found {
		return child
	}
	if child, found := n.children[WildCardStar]; found {
		return child
	}
	return nil
}

func (n *trieNode) String() string {
	var tmp, sb strings.Builder
	traverse(n, &tmp, &sb, 0)
	return sb.String()
}

func traverse(node *trieNode, tmp, sb *strings.Builder, depth int) {
	if node == nil {
		return
	}
	for ch, child := range node.children {
		tmp.WriteString(string(ch))
		if child.isEnd {
			sb.WriteString(tmp.String())
			sb.WriteByte('\n')
		}
		if len(child.children) > 0 {
			traverse(child, tmp, sb, depth+1)
		} else {
			tmp.Reset()
		}
	}
}

func newTrieNode() *trieNode {
	return &trieNode{
		children: make(map[rune]*trieNode),
	}
}

func (t *HashTrie) Reset() {
	t.root.reset()
	t.size = 0
}

// 单词表数量
func (t *HashTrie) WordsCount() int {
	return t.size
}

func (t *HashTrie) String() string {
	return t.root.String()
}

// 添加到单词表
func (t *HashTrie) AddWord(word string) {
	if word == "" {
		return
	}
	var node = &t.root
	for _, ch := range word {
		node.lazyInit()
		if _, found := node.children[ch]; !found {
			node.children[ch] = newTrieNode()
		}
		node = node.children[ch]
	}
	if !node.isEnd {
		t.size++
	}
	node.isEnd = true
}

// 最后一个字符在树中的节点
func (t *HashTrie) getTailCharNode(word []rune) *trieNode {
	var node = &t.root
	for _, ch := range word {
		child := node.contains(ch)
		if child == nil {
			return nil
		}
		node = child
	}
	return node
}

// 单词表是否包含`word`
func (t *HashTrie) contains(word string) bool {
	var node = t.getTailCharNode([]rune(word))
	if node != nil {
		return node.isEnd
	}
	return false
}

func (t *HashTrie) remove(node *trieNode, word []rune, depth int) bool {
	if node == nil {
		return false
	}
	if depth == len(word) {
		if node.isEnd {
			node.isEnd = false
			return len(node.children) > 0 // 是否还有其它单词的路径
		}
	} else {
		node = node.children[word[depth]]
		if t.remove(node, word, depth+1) {
			delete(node.children, word[depth])
			return !node.isEnd && len(node.children) > 0
		}
	}
	return false
}

// 从单词表中删除一个单词
func (t *HashTrie) Remove(word string) bool {
	if t.contains(word) {
		t.remove(&t.root, []rune(word), 0)
		t.size--
		return true
	}
	return false
}

// 从`pos`开始查找单词`word`是否包含有单词表中的单词，成功返回找到的位置，否则返回-1
func (t *HashTrie) starts(word []rune, pos int) int {
	var node = &t.root
	for pos >= 0 && pos < len(word) {
		ch := word[pos]
		child := node.contains(ch)
		if child == nil {
			return -1 // 不在字典里
		}
		if child.isEnd {
			return pos // 在字典中
		}
		node = child
		pos++
	}
	if node.isEnd {
		return pos
	}
	return -1
}

// 返回匹配的开始位置和长度
func (t *HashTrie) find(word []rune, pos int) (int, int) {
	for i := pos; i < len(word); i++ {
		idx := t.starts(word, i)
		if idx >= 0 {
			return i, idx - i + 1
		}
	}
	return -1, 0
}

// 精确匹配
func (t *HashTrie) ExactMatch(word string) bool {
	if word == "" {
		return false
	}
	runes := []rune(word)
	pos := t.starts(runes, 0)
	return pos+1 == len(runes)
}

// 模糊匹配（包含通配符）
func (t *HashTrie) Contains(word string) bool {
	runes := []rune(word)
	i, n := t.find(runes, 0)
	return i >= 0 && n > 0
}

// 将敏感字符替换为星号
func (t *HashTrie) Filter(word string) string {
	if word == "" {
		return ""
	}
	var sb strings.Builder
	start := 0
	runes := []rune(word)
	for start < len(runes) {
		pos, n := t.find(runes, start)
		if pos < 0 {
			if start == 0 {
				return word // 没有找到任何需要过滤的词
			}
			break
		} else {
			for i := start; i < pos; i++ {
				sb.WriteRune(runes[i])
			}
			if n > 0 {
				sb.WriteString(strings.Repeat("*", n))
			}
			start = pos + n
		}
	}
	for i := start; i < len(runes); i++ {
		sb.WriteRune(runes[i])
	}
	return sb.String()
}
