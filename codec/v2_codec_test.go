// Copyright © 2021-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package codec

import (
	"bytes"
	"crypto/rand"
	"testing"

	"gopkg.in/qchencc/fatchoy.v1"
	"gopkg.in/qchencc/fatchoy.v1/x/cipher"
	"gopkg.in/qchencc/fatchoy.v1/x/strutil"
)

func isEqualPacket(t *testing.T, a, b fatchoy.IPacket) bool {
	if a.Command() != b.Command() || (a.Seq() != b.Seq()) {
		return false
	}
	data1 := a.BodyToBytes()
	data2 := b.BodyToBytes()
	if len(data1) > 0 && len(data2) > 0 {
		if !bytes.Equal(data1, data2) {
			println("msg a md5sum", md5Sum(data1))
			println("msg b md5sum", md5Sum(data2))
			t.Fatalf("packet not equal, %v != %v", a, b)
			return false
		}
	}
	return true
}

func newTestPacket(bodyLen int) fatchoy.IPacket {
	var pkt testPacket
	pkt.SetType(fatchoy.PTypePacket)
	pkt.SetCommand(1234)
	pkt.SetSeq(5678)
	pkt.SetRefers([]fatchoy.NodeID{1234567, 7654321})
	if bodyLen > 0 {
		s := strutil.RandString(bodyLen)
		pkt.SetBody([]byte(s))
	}
	return &pkt
}

func createCryptor(method string) (cipher.BlockCryptor, error) {
	var key = make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	var iv = make([]byte, 32)
	if _, err := rand.Read(iv); err != nil {
		return nil, err
	}
	encrypt := cipher.NewCrypt(method, key, iv)
	return encrypt, nil
}

func testProtoCodec(t *testing.T, size int, msgSent fatchoy.IPacket) {
	encrypt, _ := createCryptor("aes-192")
	decrypt := cipher.NewCrypt("aes-192", encrypt.Key(), encrypt.IV())
	// 如果加密方式是原地加密，会导致packet的body是加密后的内容
	clone := append([]byte(nil), msgSent.BodyToBytes()...)
	encoded, err := MarshalV2(msgSent, encrypt)
	if err != nil {
		t.Fatalf("Encode with size %d: %v", size, err)
	}
	msgSent.SetBody(nil)
	var msgRecv testPacket
	if err := ReadPacketV2(bytes.NewBuffer(encoded), decrypt, &msgRecv); err != nil {
		t.Fatalf("Decode with size %d: %v", size, err)
	}
	msgSent.SetBody(clone)
	if !isEqualPacket(t, msgSent, &msgRecv) {
		t.Fatalf("Encode and Decode not equal: %d\n%v\n%v", size, msgSent, msgRecv)
	}
}

func TestCodecEncode(t *testing.T) {
	var sizeList = []int{404, 1012, 2014, 4018, 8012, 40487, 1024 * 31} //
	for _, n := range sizeList {
		var pkt = newTestPacket(n)
		testProtoCodec(t, n, pkt)
	}
}

func BenchmarkCodecMarshal(b *testing.B) {
	b.StopTimer()
	var size = 4096
	b.Logf("benchmark with message size %d\n", size)
	var msg = newTestPacket(int(size))
	b.StartTimer()

	if _, err := MarshalV1(msg, nil); err != nil {
		b.Logf("Encode: %v", err)
	}

	var msg2 testPacket
	if _, err := MarshalV1(&msg2, nil); err != nil {
		b.Logf("Decode: %v", err)
	}
}
