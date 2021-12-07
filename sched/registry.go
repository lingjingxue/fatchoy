// Copyright © 2020 qchen.fun All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package sched

import (
	"sort"
	"strings"
	"sync"

	"qchen.fun/fatchoy"
	"qchen.fun/fatchoy/l0g"
)

var (
	guard           sync.RWMutex
	serviceRegistry = make(map[string]fatchoy.Service)
	serviceIdMap    = make(map[uint8]string)
)

// 注册服务
func Register(service fatchoy.Service) {
	guard.Lock()
	defer guard.Unlock()
	var name = strings.ToUpper(service.Name())
	var typ = service.Type()
	if _, dup := serviceRegistry[name]; dup {
		l0g.Panicf("duplicate registration of service %x", name)
	}
	if _, dup := serviceIdMap[typ]; dup {
		l0g.Panicf("duplicate service type of service %x", typ)
	}
	serviceRegistry[name] = service
	serviceIdMap[typ] = name
}

// 根据服务ID获取Service对象
func GetServiceByID(srvType uint8) fatchoy.Service {
	guard.RLock()
	var v fatchoy.Service
	if name, ok := serviceIdMap[srvType]; ok {
		v = serviceRegistry[name]
	}
	guard.RUnlock()
	return v
}

// 根据名称获取Service对象
func GetServiceByName(name string) fatchoy.Service {
	guard.RLock()
	v := serviceRegistry[strings.ToUpper(name)]
	guard.RUnlock()
	return v
}

// 所有服务类型名
func GetServiceNames() []string {
	guard.RLock()
	var names = make([]string, 0, len(serviceRegistry))
	for s, _ := range serviceRegistry {
		names = append(names, s)
	}
	guard.RUnlock()
	sort.Strings(names)
	return names
}
