// Copyright © 2020-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package cipher

type noneCrypt struct{}

func NewNoneCrypt(key, iv []byte) BlockCryptor {
	return new(noneCrypt)
}

func (c *noneCrypt) Key() []byte {
	return nil
}

func (c *noneCrypt) IV() []byte {
	return nil
}

func (c *noneCrypt) Encrypt(src []byte) []byte {
	return src
}

func (c *noneCrypt) Decrypt(src []byte) []byte {
	return src
}
