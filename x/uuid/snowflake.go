// Copyright © 2020-present simon@qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package uuid

import (
	"errors"
	"log"
	"net"
	"sync"
	"time"
)

const (
	SequenceBits       = 10
	MachineIDBits      = 14
	TimeUnitBits       = 37
	MaxSeqID           = (1 << SequenceBits) - 1
	MachineIDMask      = 0x3FFFF
	TimestampShift     = MachineIDBits + SequenceBits
	MaxTimeUnits       = (1 << TimeUnitBits) - 1
	BackwardsMaskShift = TimeUnitBits + MachineIDBits + SequenceBits

	TimeUnit    = int64(time.Second / 100)        // 厘秒（10毫秒）
	CustomEpoch = int64(1577836800 * time.Second) // 起始纪元 2020-01-01 00:00:00 UTC
)

var (
	ErrClockGoneBackwards = errors.New("clock gone backwards")
	ErrTimeUnitOverflow   = errors.New("clock time unit overflow")
	ErrUUIDIntOverflow    = errors.New("uuid integer overflow")
)

// UUID的生成依赖系统时钟，如果系统时钟被回拨，会有潜在的生成重复ID的情况
// 	1，系统时钟被人为回拨（目前已经有`timeOffset`提供逻辑时钟机制）
// 	2，NTP同步和UTC闰秒(https://en.wikipedia.org/wiki/Leap_second)
// 设计中增加了时钟回拨标记位，可以让系统在时钟被回拨时仍正确工作
func currentTimeUnit() int64 {
	return (time.Now().UTC().UnixNano() - CustomEpoch) / TimeUnit // to centi-seconds
}

// sequence expired, tick to next time unit
func waitUntilNextTimeUnit(ts int64) int64 {
	for {
		time.Sleep(time.Millisecond)
		var now = currentTimeUnit()
		if now > ts {
			return now
		}
	}
}

// 一个64位UUID由以下部分组成
//	  1位符号位
//	 2位时钟回拨标记，支持时钟被回拨3次
//	 37位时间戳（厘秒），可以表示到2063-07-21
//	 14位服务器ID
//	 10位序列号，单个时间单位的最大分配数量
type Snowflake struct {
	machineID      int64      // id of this machine(process)
	guard          sync.Mutex //
	seq            int64      // last sequence ID
	lastTimeUnit   int64      // last time unit
	lastID         int64      // last generated id
	backwardsCount int64      // 允许时钟被回拨3次
}

func NewSnowflake(machineId uint16) *Snowflake {
	if machineId == 0 {
		machineId = privateIP4()
		log.Printf("snowflake auto set machine id to %d", machineId)
	}
	var sf = &Snowflake{
		machineID:    int64(machineId) & MachineIDMask,
		lastTimeUnit: currentTimeUnit(),
	}
	return sf
}

// Next generate an ID, panic if time gone backwards or integer overflow
func (sf *Snowflake) Next() (int64, error) {
	sf.guard.Lock()
	defer sf.guard.Unlock()

	var currentTs = currentTimeUnit()
	if currentTs > MaxTimeUnits {
		log.Printf("Snowflake: time unit overflow")
		return 0, ErrTimeUnitOverflow
	}
	if currentTs < sf.lastTimeUnit {
		log.Printf("Snowflake: time has gone backwards")
		if sf.backwardsCount > 3 {
			return 0, ErrClockGoneBackwards
		}
		sf.backwardsCount++
	}
	if currentTs == sf.lastTimeUnit {
		sf.seq++
		if sf.seq > MaxSeqID {
			sf.seq = 0
			currentTs = waitUntilNextTimeUnit(currentTs)
		}
	} else {
		sf.seq = 0
	}

	sf.lastTimeUnit = currentTs
	var backwardsMask = sf.backwardsCount << BackwardsMaskShift
	var uuid = backwardsMask | (currentTs << TimestampShift) | (sf.machineID << SequenceBits) | sf.seq
	if uuid <= sf.lastID {
		log.Printf("Snowflake: uuid int64 overflow")
		return 0, ErrUUIDIntOverflow
	}
	sf.lastID = uuid
	return uuid, nil
}

func (sf *Snowflake) MustNext() int64 {
	if n, err := sf.Next(); err != nil {
		panic(err)
	} else {
		return n
	}
}

// lower 16-bits of IPv4
func privateIP4() uint16 {
	addr, err := net.InterfaceAddrs()
	if err != nil {
		return 0
	}
	var isPrivateIPv4 = func(ip net.IP) bool {
		return ip[0] == 10 || ip[0] == 172 && (ip[1] >= 16 &&
			ip[1] < 32) || ip[0] == 192 && ip[1] == 168
	}
	for _, a := range addr {
		ipnet, ok := a.(*net.IPNet)
		if !ok || ipnet.IP.IsLoopback() {
			continue
		}
		var ip = ipnet.IP.To4()
		if ip != nil && isPrivateIPv4(ip) {
			return uint16(ip[2])<<8 + uint16(ip[3])
		}
	}
	return 0
}
