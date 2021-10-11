// Copyright © 2016-present ichenq@outlook.com. All Rights Reserved.
//
// Any redistribution or reproduction of part or all of the contents in any form
// is prohibited.
//
// You may not, except with our express written permission, distribute or commercially
// exploit the content. Nor may you transmit it or store it in any other website or
// other form of electronic retrieval system.

// +build !ignore

package stats

import (
	"sync"
	"testing"
)

func TestStatsSet(t *testing.T) {
	count := 10000
	s := New(count)
	for i := 0; i < count; i++ {
		s.Set(i, int64(i*10))
	}
	for i := 0; i < count; i++ {
		if v := s.Get(i); v != int64(i*10) {
			t.Fatalf("index: %v, expect %v, but got %v", i, i*10, v)
		}
	}
}

func TestStatsConcurrentSet(t *testing.T) {
	count := 1000
	var wg sync.WaitGroup
	s := New(count)

	t.Logf("Stats multiple Set")
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			s.Set(idx, int64(idx*10))
		}(i)
	}
	// waiting for Set()
	wg.Wait()
	t.Logf("Stats multiple Get")

	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if v := s.Get(idx); v != int64(idx*10) {
				t.Fatalf("index: %v, expect %v, but got %v", idx, idx*10, v)
			}
		}(i)
	}
	// waiting for Get()
	wg.Wait()
}

func TestStatsConcurrentAdd(t *testing.T) {
	count := 100000
	var wg sync.WaitGroup
	s := New(count)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < count; i++ {
			s.Add(i, int64(i*10+1))
		}
	}()
	wg.Wait()
	slice := s.Copy()
	for i := 0; i < count; i++ {
		if v := slice[i]; v != int64(i*10+1) {
			t.Fatalf("index: %v, expect %v, but got %v", i, i*10+1, v)
		}
	}
}
