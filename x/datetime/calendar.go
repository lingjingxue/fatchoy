// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package datetime

import "time"

var (
	DefaultLoc       = time.Local
	FirstDayIsMonday = true
)

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

// 当日零点
func MidnightTimeOf(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// N天后的这个时候
func ThisMomentAfterDays(this time.Time, days int) time.Time {
	if days == 0 {
		return this
	}
	return this.Add(time.Duration(days) * time.Hour * 24)
}

// 本周的起点
func StartingOfWeek(t time.Time) time.Time {
	t2 := MidnightTimeOf(t)
	weekday := int(t2.Weekday())
	if FirstDayIsMonday {
		if weekday == 0 {
			weekday = 7
		}
		weekday = weekday - 1
	}
	d := time.Duration(-weekday) * 24 * time.Hour
	return t2.Add(d)
}

// 本周的最后一天
func EndOfWeek(t time.Time) time.Time {
	begin := StartingOfWeek(t)
	end := ThisMomentAfterDays(begin, 7)
	return end.Add(-time.Second) // 23:59:59
}

// 年度第一天
func FirstDayOfYear(year int) time.Time {
	return time.Date(year, 1, 1, 0, 0, 0, 0, DefaultLoc)
}

// 年度最后一天
func LastDayOfYear(year int) time.Time {
	return time.Date(year, 12, 31, 0, 0, 0, 0, DefaultLoc)
}

// 获取两个时间中经过的天数
func ElapsedDaysBetween(start, end time.Time) int {
	var negative = false
	if start.After(end) {
		start, end = end, start
		negative = true
	}
	var days = 0
	if start.Year() != end.Year() {
		t := LastDayOfYear(start.Year())
		days = t.YearDay() - start.YearDay() // start年份的天数
		for i := start.Year() + 1; i < end.Year(); i++ {
			var t1 = LastDayOfYear(i)
			days += t1.YearDay() // start-end中间每年的天数
		}
		days += end.YearDay() // end年份的天数
	} else {
		days = end.YearDay() - start.YearDay()
	}
	if negative {
		days = -days
	}
	return days
}
