// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package cipher

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"math/rand"
	"testing"
)

func TestDESCrypt(t *testing.T) {
	iv := randBytes(16)
	key := randBytes(24)
	encryptor := NewTripleDES(key, iv)
	decryptor := NewTripleDES(key, iv)
	for i := 0; i < 100; i++ {
		payload := randBytes(100 + rand.Int()%1000)
		encrypted := encryptor.Encrypt(cloneBytes(payload))
		decrypted := decryptor.Decrypt(encrypted)
		if !bytes.Equal(payload, decrypted) {
			checksum1 := fmt.Sprintf("%x", md5.Sum(payload))
			checksum2 := fmt.Sprintf("%x", md5.Sum(decrypted))
			t.Fatalf("encryption mismatch %s != %s", checksum1, checksum2)
		}
	}
}
