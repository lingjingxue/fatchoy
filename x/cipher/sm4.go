// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package cipher

import (
	"crypto/cipher"
	"log"

	"github.com/tjfoc/gmsm/sm4"
)

// https://en.wikipedia.org/wiki/SM4_(cipher)
type sm4Crypt struct {
	encbuf  [sm4.BlockSize]byte
	decbuf  [2 * sm4.BlockSize]byte
	block   cipher.Block
	key, iv []byte
}

// key must be 16-byte
func NewSM4(key, iv []byte) BlockCryptor {
	block, err := sm4.NewCipher(key)
	if err != nil {
		log.Panicf("%v", err)
	}
	return &sm4Crypt{
		block: block,
		key:   key,
		iv:    iv,
	}
}

func (c *sm4Crypt) Key() []byte {
	return c.key
}

func (c *sm4Crypt) IV() []byte {
	return c.iv
}

func (c *sm4Crypt) Encrypt(src []byte) []byte {
	encrypt(c.block, c.iv, src, src, c.encbuf[:])
	return src
}

func (c *sm4Crypt) Decrypt(src []byte) []byte {
	decrypt(c.block, c.iv, src, src, c.decbuf[:])
	return src
}
