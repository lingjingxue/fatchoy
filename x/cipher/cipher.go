// Copyright © 2020-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package cipher

type BlockCryptor interface {
	Key() []byte
	IV() []byte

	Encrypt(src []byte) []byte
	Decrypt(src []byte) []byte
}

//
func NewCrypt(name string, key, iv []byte) BlockCryptor {
	switch name {
	case "aes-128":
		return NewAESCFB(key[:16], iv)
	case "aes-192":
		return NewAESCFB(key[:24], iv)
	case "sm4":
		return NewSM4(key[:16], iv)
	case "twofish":
		return NewTwofish(key, iv)
	case "3des":
		return NewTripleDES(key[:24], iv)
	case "xtea":
		return NewXTEA(key[:16], iv)
	case "salsa20":
		return NewSalsa20(key[:32], iv)
	case "none":
		return NewNoneCrypt(key, iv)
	default:
		return NewAESCFB(key[:32], iv)
	}
}
