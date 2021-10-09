// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package cipher

import (
	"golang.org/x/crypto/salsa20"
)

// https://en.wikipedia.org/wiki/Salsa20
type salsa20Crypt struct {
	key   [32]byte
	nonce [8]byte
}

func NewSalsa20(key, iv []byte) BlockCryptor {
	s := &salsa20Crypt{}
	copy(s.key[:], key)
	copy(s.nonce[:], iv)
	return s
}

func (c *salsa20Crypt) Key() []byte {
	return c.key[:]
}

func (c *salsa20Crypt) IV() []byte {
	return c.nonce[:]
}

func (c *salsa20Crypt) Encrypt(data []byte) []byte {
	salsa20.XORKeyStream(data, data, c.nonce[:], &c.key)
	return data
}

func (c *salsa20Crypt) Decrypt(data []byte) []byte {
	return c.Encrypt(data)
}
