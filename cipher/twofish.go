// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package cipher

import (
	"crypto/cipher"
	"log"

	"golang.org/x/crypto/twofish"
)

type twofishCrypt struct {
	encbuf  [twofish.BlockSize]byte
	decbuf  [2 * twofish.BlockSize]byte
	block   cipher.Block
	key, iv []byte
}

// key should be 16, 24 or 32 bytes
func NewTwofish(key, iv []byte) BlockCryptor {
	block, err := twofish.NewCipher(key)
	if err != nil {
		log.Panicf("%v", err)
	}
	return &twofishCrypt{
		block: block,
		key:   key,
		iv:    iv,
	}
}

func (c *twofishCrypt) Key() []byte {
	return c.key
}

func (c *twofishCrypt) IV() []byte {
	return c.iv
}

func (c *twofishCrypt) Encrypt(src []byte) []byte {
	encrypt(c.block, c.iv, src, src, c.encbuf[:])
	return src
}

func (c *twofishCrypt) Decrypt(src []byte) []byte {
	decrypt(c.block, c.iv, src, src, c.decbuf[:])
	return src
}
