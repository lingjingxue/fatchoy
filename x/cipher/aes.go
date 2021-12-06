// Copyright © 2020-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package cipher

import (
	"crypto/aes"
	"crypto/cipher"
	"log"
)

// https://en.wikipedia.org/wiki/Advanced_Encryption_Standard
type aesCFBCrypt struct {
	encbuf  [aes.BlockSize]byte
	decbuf  [2 * aes.BlockSize]byte
	block   cipher.Block
	key, iv []byte
}

// key should be 16, 24, or 32 bytes
func NewAESCFB(key, iv []byte) BlockCryptor {
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Panicf("%v", err)
	}
	return &aesCFBCrypt{
		block: block,
		key:   key,
		iv:    iv,
	}
}

func (c *aesCFBCrypt) Key() []byte {
	return c.key
}

func (c *aesCFBCrypt) IV() []byte {
	return c.iv
}

func (c *aesCFBCrypt) Encrypt(src []byte) []byte {
	encrypt(c.block, c.iv, src, src, c.encbuf[:])
	return src
}

func (c *aesCFBCrypt) Decrypt(src []byte) []byte {
	decrypt(c.block, c.iv, src, src, c.decbuf[:])
	return src
}
