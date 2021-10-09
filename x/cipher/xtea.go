// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package cipher

import (
	"crypto/cipher"
	"log"

	"golang.org/x/crypto/xtea"
)

// https://en.wikipedia.org/wiki/XTEA
type xteaCrypt struct {
	encbuf  [xtea.BlockSize]byte
	decbuf  [2 * xtea.BlockSize]byte
	block   cipher.Block
	key, iv []byte
}

// key must be 16 bytes
func NewXTEA(key, iv []byte) BlockCryptor {
	block, err := xtea.NewCipher(key)
	if err != nil {
		log.Panicf("%v", err)
	}
	return &xteaCrypt{
		block: block,
		key:   key,
		iv:    iv,
	}
}

func (c *xteaCrypt) Key() []byte {
	return c.key
}

func (c *xteaCrypt) IV() []byte {
	return c.iv
}

func (c *xteaCrypt) Encrypt(src []byte) []byte {
	encrypt(c.block, c.iv, src, src, c.encbuf[:])
	return src
}

func (c *xteaCrypt) Decrypt(src []byte) []byte {
	decrypt(c.block, c.iv, src, src, c.decbuf[:])
	return src
}
