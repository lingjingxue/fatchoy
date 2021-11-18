// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

//go:build !ignore

package fsutil

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"
)

var showCompressRate = true

func testCase(t *testing.T, fn func(int) []byte) {
	var size = 1
	for size < 10000 {
		data := fn(size)
		size += rand.Int() % 100
		compressed, err := CompressBytes(data)
		if err != nil {
			t.Fatalf("compress: %v", err)
		}
		uncompressed, err := UncompressBytes(compressed)
		if err != nil {
			t.Fatalf("compress: %v", err)
		}
		if !bytes.Equal(data, uncompressed) {
			t.Fatalf("not equal")
		}
		if showCompressRate {
			var before, after = len(data), len(compressed)
			var rate = float64(after) / float64(before)
			if rate < 1.0 {
				t.Logf("compress rate: %d/%d = %f", before, after, rate)
			}
		}
	}
}

func TestCompressZlib(t *testing.T) {
	testCase(t, randBytes)
}

func TestCompressWriter(t *testing.T) {
	for i := 1; i <= 100; i++ {
		var data = randBytes(256 + rand.Int()%1024)
		compressed, err := CompressBytes(data)
		if err != nil {
			t.Fatalf("doCompressWithWriter: %v, turn: %d", err, i)
		}
		uncompressed, err := UncompressBytes(compressed)
		if err != nil {
			t.Fatalf("compress: %v, turn: %d", err, i)
		}
		if !bytes.Equal(data, uncompressed) {
			t.Fatalf("not equal")
		}
	}
}

func TestCompressRate(t *testing.T) {
	var stats = make([]float64, 0, 100)
	for i := 1; i <= 100; i++ {
		var data = randBytes(64 + rand.Int()%8192)
		compressed, err := CompressBytes(data)
		if err != nil {
			t.Fatalf("doCompressWithWriter: %v, turn: %d", err, i)
		}
		var rate = float64(len(compressed)) / float64(len(data))
		stats = append(stats, rate)
	}
	var sum float64
	for _, v := range stats {
		sum += v
	}
	var avg = sum / float64(len(stats))
	fmt.Printf("compress, total %v, avg: %v\n", sum, avg)
}

func BenchmarkZlibCompress1(b *testing.B) {
	var data = randBytes(8192)
	_, err := CompressBytes(data)
	if err != nil {
		b.Fatalf("compress: %v", err)
	}
	b.StopTimer()
}

func BenchmarkZlibUncompress(b *testing.B) {
	b.StopTimer()

	var data = randBytes(1024)
	compressed, err := CompressBytes(data)
	if err != nil {
		b.Fatalf("compress: %v", err)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, err = UncompressBytes(compressed)
		if err != nil {
			b.Fatalf("compress: %v", err)
		}
		b.StopTimer()
		b.StartTimer()
	}
}
