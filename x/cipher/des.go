// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package cipher

import (
	"crypto/cipher"
	"crypto/des"
	"log"
)

// https://en.wikipedia.org/wiki/Triple_DES
type tripleDESCrypt struct {
	encbuf  [des.BlockSize]byte
	decbuf  [2 * des.BlockSize]byte
	block   cipher.Block
	key, iv []byte
}

// key must be 24-byte
func NewTripleDES(key, iv []byte) BlockCryptor {
	block, err := des.NewTripleDESCipher(key)
	if err != nil {
		log.Panicf("%v", err)
	}
	return &tripleDESCrypt{
		block: block,
		key:   key,
		iv:    iv,
	}
}

func (c *tripleDESCrypt) Key() []byte {
	return c.key
}

func (c *tripleDESCrypt) IV() []byte {
	return c.iv
}

func (c *tripleDESCrypt) Encrypt(src []byte) []byte {
	encrypt(c.block, c.iv, src, src, c.encbuf[:])
	return src
}

func (c *tripleDESCrypt) Decrypt(src []byte) []byte {
	decrypt(c.block, c.iv, src, src, c.decbuf[:])
	return src
}
