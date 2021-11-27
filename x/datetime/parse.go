// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package datetime

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

const (
	DateFormat      = "2006-01-02 15:04:05"
	TimestampFormat = "2006-01-02 15:04:05.999"
)

var (
	DummyTime     time.Time
	DateFormatSep = "-"
	TimeFormatSep = ":"
	DateTimeSep   = " "
)

//
// 实现一个对时间日期的解析， 要求日期格式如：2001-3-4 12:34:56
// 自己实现可以方便的定制，更加灵活

// 解析年月日格式，如:2001-3-4, 12-4
func ParseDateParts(s string, seg []int) error {
	i := strings.Index(s, DateFormatSep)
	j := strings.LastIndex(s, DateFormatSep)
	if i <= 0 || j <= 0 {
		return fmt.Errorf("invalid date format '%s'", s)
	}

	var part1, part2, part3 string
	if i == j { // [月-日] 格式
		part2 = s[:i]
		part3 = s[i+1:]
	} else { // [年-月-日]格式
		part1 = s[:i]
		part2 = s[i+1 : j]
		part3 = s[j+1:]
	}

	day, _ := strconv.Atoi(part3)
	if day < 1 || day > 31 {
		return fmt.Errorf("invalid date format '%s'", s)
	}
	seg[2] = day

	month, _ := strconv.Atoi(part2)
	if month < 1 || month > 12 {
		return fmt.Errorf("invalid date format '%s'", s)
	}
	seg[1] = month

	if part1 != "" {
		year, _ := strconv.Atoi(part1)
		if year <= 0 {
			return fmt.Errorf("invalid date format '%s'", s)
		}
		if max := DaysCountOfMonth(year, month); day > max {
			return fmt.Errorf("invalid date format '%s'", s)
		}
		seg[0] = year
	}
	return nil
}

// 解析时分秒格式，如:12:34:56
func ParseTimeParts(s string, seg []int) error {
	i := strings.Index(s, TimeFormatSep)
	j := strings.LastIndex(s, TimeFormatSep)
	if i <= 0 || j <= 0 {
		return fmt.Errorf("invalid date format '%s'", s)
	}
	var part1, part2, part3 string
	if i == j { // [分:秒] 格式
		part2 = s[:i]
		part3 = s[i+1:]
	} else { // [时:分:秒 格式
		part1 = s[:i]
		part2 = s[i+1 : j]
		part3 = s[j+1:]
	}
	sec, _ := strconv.Atoi(part3)
	if sec < 0 || sec > 60 {
		return fmt.Errorf("invalid time format '%s'", s)
	}
	seg[2] = sec

	min, _ := strconv.Atoi(part2)
	if min < 0 || min >= 60 {
		return fmt.Errorf("invalid time format '%s'", s)
	}
	seg[1] = min

	if part1 != "" {
		hour, _ := strconv.Atoi(part1)
		if hour < 0 || hour > 23 {
			return fmt.Errorf("invalid time format '%s'", s)
		}
		seg[0] = hour
	}
	return nil
}

// 根据当前时间和字符串日期，生成一个新的time.Time对象
func ParseMomentTime(tm time.Time, timeText string) (time.Time, error) {
	var parts = []int{-1, 0, 0}
	if err := ParseTimeParts(timeText, parts); err != nil {
		return DummyTime, err
	}
	if parts[0] < 0 {
		parts[0] = tm.Hour()
	}
	moment := time.Date(tm.Year(), tm.Month(), tm.Day(), parts[0], parts[1], parts[2], 0, tm.Location())
	return moment, nil
}

func MustParseMomentTime(tm time.Time, s string) time.Time {
	t, err := ParseMomentTime(tm, s)
	if err != nil {
		log.Panicf("MustParseMomentTime: %s, %v", s, err)
	}
	return t
}

// 根据当前时间和字符串日期，生成一个新的time.Time对象
func ParseMomentDate(tm time.Time, dateText string) (time.Time, error) {
	var parts = []int{-1, 0, 0}
	if err := ParseDateParts(dateText, parts); err != nil {
		return DummyTime, err
	}
	if parts[0] < 0 {
		parts[0] = tm.Year()
	}
	moment := time.Date(parts[0], time.Month(parts[1]), parts[2], tm.Hour(), tm.Minute(), tm.Second(), 0, tm.Location())
	return moment, nil
}

func MustParseMomentDate(tm time.Time, s string) time.Time {
	t, err := ParseMomentDate(tm, s)
	if err != nil {
		log.Panicf("MustParseMomentDate: %s, %v", s, err)
	}
	return t
}

// 解析时间字符串 [2001-5-7 12:34:56]
func ParseDateTime(s string) (time.Time, error) {
	i := strings.Index(s, DateTimeSep)
	if i <= 0 {
		return DummyTime, fmt.Errorf("invalid time format '%s'", s)
	}
	var dp, tp [3]int
	if er := ParseDateParts(s[:i], dp[:]); er != nil {
		return DummyTime, fmt.Errorf("invalid time format '%s'", s)
	}
	if er := ParseTimeParts(s[i+1:], tp[:]); er != nil {
		return DummyTime, fmt.Errorf("invalid time format '%s'", s)
	}
	ts := time.Date(dp[0], time.Month(dp[1]), dp[2], tp[0], tp[1], tp[2], 0, DefaultLoc)
	// println(s, "==>",ts.Format(DateFormat))
	return ts, nil
}

func MustParseDateTime(s string) time.Time {
	t, err := ParseDateTime(s)
	if err != nil {
		log.Panicf("MustParseDateTime: %s, %v", s, err)
	}
	return t
}

// 是否闰年
func IsLeapYear(year int) bool {
	return (year%4 == 0 && year%100 != 0) ||
		year%400 == 0
}

// 一个月的天数
func DaysCountOfMonth(year, month int) int {
	switch time.Month(month) {
	case time.January:
		return 31
	case time.February:
		if year > 0 && IsLeapYear(year) {
			return 29
		}
		return 28
	case time.March:
		return 31
	case time.April:
		return 30
	case time.May:
		return 31
	case time.June:
		return 30
	case time.July:
		return 31
	case time.August:
		return 31
	case time.September:
		return 30
	case time.October:
		return 31
	case time.November:
		return 30
	case time.December:
		return 31
	}
	return 0
}

// 解析时间字符串
func MustParseTime(s string) time.Time {
	t, err := time.ParseInLocation(TimestampFormat, s, time.Local)
	if err != nil {
		log.Panicf("MustParseTime: %s, %v", s, err)
	}
	return t
}

// 格式化字符串
func FormatTime(t time.Time) string {
	return t.Format(TimestampFormat)
}

// 格式化时间字符串
func FormatUnixTime(v int64) string {
	return time.Unix(v, 0).Format(DateFormat)
}

// 转换unix毫秒时间戳
func TimeToMillis(t time.Time) int64 {
	return t.UnixNano() / int64(1000_000)
}

// 当前unix毫秒时间戳
func CurrentTimeMillis() int64 {
	return time.Now().UnixNano() / int64(1000_000)
}
