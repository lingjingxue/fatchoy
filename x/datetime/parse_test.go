// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

//go:build !ignore

package datetime

import (
	"testing"
	"time"
)

const timeFormat = "2006-01-02 15:04:05"

// 测试解析时间格式
func TestParseTimeFormat(t *testing.T) {
	type testCase struct {
		shouldPass bool
		text       string
		hour       int
		minute     int
		second     int
	}

	var cases = []testCase{
		{true, "12:34:56", 12, 34, 56},
		{true, "01:02:03", 1, 2, 3},
		{true, "12:34", 0, 12, 34},
		{true, "23:59:59", 23, 59, 59},
		{false, "-12:-34:-56", 0, 0, 0},
		{false, "24:60:60", 0, 0, 0},
	}
	for _, tcase := range cases {
		var parts [3]int
		err := ParseTimeParts(tcase.text, parts[:])
		if tcase.shouldPass {
			if err != nil {
				t.Fatalf("ParseTimeParts: failed %v, %v", tcase.text, err)
			}
		} else {
			if err == nil {
				t.Fatalf("ParseTimeParts: should not pass, %v, %v", tcase.text, err)
			} else {
				continue
			}
		}
		if parts[0] != tcase.hour {
			t.Fatalf("ParseTimeParts: %s, hour %d != %d", tcase.text, parts[0], tcase.hour)
		}
		if parts[1] != tcase.minute {
			t.Fatalf("ParseTimeParts: %s, min %d != %d", tcase.text, parts[1], tcase.minute)
		}
		if parts[2] != tcase.second {
			t.Fatalf("ParseTimeParts: %s, sec %d != %d", tcase.text, parts[2], tcase.second)
		}
	}
}

// 测试解析日期格式
func TestParseDateFormat(t *testing.T) {
	type testCase struct {
		shouldPass bool
		text       string
		year       int
		month      int
		day        int
	}

	var cases = []testCase{
		{true, "0001-01-01", 1, 1, 1},
		{true, "1234-01-01", 1234, 1, 1},
		{true, "12-31", 0, 12, 31},
		{false, "-12:-34:-56", 0, 0, 0},
		{false, "1234:0:1", 0, 0, 0},
		{false, "1234:-1:0", 0, 0, 0},
		{false, "1234:13:32", 0, 0, 0},
	}
	for _, tcase := range cases {
		var parts [3]int
		err := ParseDateParts(tcase.text, parts[:])
		if tcase.shouldPass {
			if err != nil {
				t.Fatalf("ParseDateParts: failed %v, %v", tcase.text, err)
			}
		} else {
			if err == nil {
				t.Fatalf("ParseDateParts: should not pass, %v, %v", tcase.text, err)
			} else {
				continue
			}
		}
		//println(tcase.text, year, mon, day)
		if parts[0] != tcase.year {
			t.Fatalf("ParseDateParts: %s, year %d != %d", tcase.text, parts[0], tcase.year)
		}
		if parts[1] != tcase.month {
			t.Fatalf("ParseDateParts: %s, month %d != %d", tcase.text, parts[1], tcase.month)
		}
		if parts[2] != tcase.day {
			t.Fatalf("ParseDateParts: %s, day %d != %d", tcase.text, parts[2], tcase.day)
		}
	}
}

func TestParseDateTime(t *testing.T) {
	type testCase struct {
		shouldPass       bool
		text             string
		year, month, day int
		hour, min, sec   int
	}
	var cases = []testCase{
		{true, "2001-01-02 12:34:56", 2001, 1, 2, 12, 34, 56},
		{true, "1-2-3 0:0:0", 1, 2, 3, 0, 0, 0},
	}
	for _, tcase := range cases {
		ts, err := ParseDateTime(tcase.text)
		if tcase.shouldPass {
			if err != nil {
				t.Fatalf("ParseDateTime: failed %v, %v", tcase.text, err)
			}
		} else {
			if err == nil {
				t.Fatalf("ParseDateTime: should not pass, %v, %v", tcase.text, err)
			} else {
				continue
			}
		}
		if ts.Year() != tcase.year {
			t.Fatalf("ParseDateTime: %s, year %d != %d", tcase.text, ts.Year(), tcase.year)
		}
		if int(ts.Month()) != tcase.month {
			t.Fatalf("ParseDateTime: %s, month %d != %d", tcase.text, ts.Month(), tcase.month)
		}
		if ts.Day() != tcase.day {
			t.Fatalf("ParseDateTime: %s, %d != %d", tcase.text, ts.Day(), tcase.day)
		}
		if ts.Hour() != tcase.hour {
			t.Fatalf("ParseDateTime: %s, %d != %d", tcase.text, ts.Hour(), tcase.hour)
		}
		if ts.Minute() != tcase.min {
			t.Fatalf("ParseDateTime: %s, %d != %d", tcase.text, ts.Minute(), tcase.min)
		}
		if ts.Second() != tcase.sec {
			t.Fatalf("ParseDateTime: %s, %d != %d", tcase.text, ts.Second(), tcase.sec)
		}
	}
}

func BenchmarkParseDateTime(b *testing.B) {
	s := "2001-01-02 12:34:56"
	for i := 0; i < 10000; i++ {
		if _, err := ParseDateTime(s); err != nil {
			b.Fatalf("%v", err)
		}
	}
	// Output:
	// BenchmarkParseDateTime-12    	1000000000	         0.00100 ns/op
}

func BenchmarkParseDateTimeStd(b *testing.B) {
	s := "2001-01-02 12:34:56"
	for i := 0; i < 10000; i++ {
		if _, err := time.Parse(DateFormat, s); err != nil {
			b.Fatalf("%v", err)
		}
	}
	// Output:
	// BenchmarkParseDateTime-12    	1000000000	         0.00200 ns/op
}

func TestParseMomentTime(t *testing.T) {
	s := "12:34:56"
	tm := time.Now()
	ts := MustParseMomentTime(tm, s)
	if ts.Hour() != 12 {
		t.Fatalf("hour %d != %d", ts.Hour(), 12)
	}
	if ts.Minute() != 34 {
		t.Fatalf("hour %d != %d", ts.Minute(), 34)
	}
	if ts.Second() != 56 {
		t.Fatalf("hour %d != %d", ts.Second(), 56)
	}
}

func TestParseMomentDate(t *testing.T) {
	s := "2001-01-02"
	tm := time.Now()
	ts := MustParseMomentDate(tm, s)
	if ts.Year() != 2001 {
		t.Fatalf("hour %d != %d", ts.Year(), 2001)
	}
	if ts.Month() != 1 {
		t.Fatalf("hour %d != %d", ts.Month(), 1)
	}
	if ts.Day() != 2 {
		t.Fatalf("hour %d != %d", ts.Day(), 2)
	}
}
