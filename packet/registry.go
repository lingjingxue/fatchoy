// Copyright © 2020-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package packet

import (
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

var (
	msgIdNames      = make(map[int32]string)        // 消息ID --> 消息名称
	msgNameIds      = make(map[string]int32)        // 消息名称 --> 消息ID
	msgTypeRegistry = make(map[string]reflect.Type) // 消息名称 --> reflect类型
)

// 消息协议规则:
//  1, 请求消息以Req结尾
//  2, 响应消息以Ack结尾
//  3, 通知消息以Ntf结尾
var validNameSuffix = []string{"Req", "Ack", "Ntf"}

func hasValidSuffix(name string) bool {
	for _, suffix := range validNameSuffix {
		if strings.HasSuffix(name, suffix) {
			return true
		}
	}
	return false
}

var wellKnownPkg = []string{"google/", "github.com/", "grpc/"}

func isWellKnown(name string) bool {
	for _, prefix := range wellKnownPkg {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

func isNil(c interface{}) bool {
	if c == nil {
		return true
	}
	return reflect.ValueOf(c).Kind() == reflect.Ptr && reflect.ValueOf(c).IsNil()
}

// 从message的option里获取消息ID
func getMsgIdByExtension(descriptor protoreflect.MessageDescriptor, xtName protoreflect.FullName) int32 {
	var ovi = descriptor.Options()
	if isNil(ovi) {
		return 0
	}
	var omi = ovi.ProtoReflect()
	var msgId int32
	omi.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		if !fd.IsExtension() {
			return true
		}
		if fd.FullName() == xtName {
			var ivs = v.String()
			n, _ := strconv.ParseInt(ivs, 10, 64)
			msgId = int32(n)
			return false
		}
		return true
	})
	return msgId
}

// 根据消息option指定的ID注册
func registerByExtension(fd protoreflect.FileDescriptor, xtName protoreflect.FullName) bool {
	if isWellKnown(fd.Path()) {
		return true
	}
	var descriptors = fd.Messages()
	for i := 0; i < descriptors.Len(); i++ {
		var descriptor = descriptors.Get(i)
		var fullname = string(descriptor.FullName())
		if !hasValidSuffix(fullname) {
			continue
		}
		var msgId = getMsgIdByExtension(descriptor, xtName)
		if msgId == 0 {
			continue
		}
		mt, err := protoregistry.GlobalTypes.FindMessageByName(descriptor.FullName())
		if err != nil {
			continue
		}
		rtype := reflect.TypeOf(proto.MessageV1(mt.Zero().Interface()))
		if s, found := msgIdNames[msgId]; found {
			log.Panicf("duplicate message hash %s %s, %d", s, fullname, msgId)
		}
		var name = rtype.Elem().String()
		msgTypeRegistry[name] = rtype.Elem()
		msgNameIds[name] = msgId
		msgIdNames[msgId] = name
	}
	return true
}

func RegisterMsgID(exName string) {
	var fullname = protoreflect.FullName(exName)
	_, err := protoregistry.GlobalTypes.FindExtensionByName(fullname)
	if err != nil {
		log.Panicf("%v", err)
	}
	protoregistry.GlobalFiles.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		return registerByExtension(fd, fullname)
	})
	log.Printf("%d messages registered\n", len(msgTypeRegistry))
}

// 根据名称创建消息
func CreateMessageByName(name string) proto.Message {
	if rtype, ok := msgTypeRegistry[name]; ok {
		msg := reflect.New(rtype).Interface()
		return msg.(proto.Message)
	}
	return nil
}

// 根据消息ID创建消
func CreateMessageByID(msgId int32) proto.Message {
	if name, found := msgIdNames[msgId]; found {
		return CreateMessageByName(name)
	}
	return nil
}

// 根据message获取消息ID
func GetMessageIDOf(msg proto.Message) int32 {
	var rtype = reflect.TypeOf(msg).Elem()
	var fullname = rtype.String()
	return msgNameIds[fullname]
}

func GetMessageIDByName(msgName string) int32 {
	return msgNameIds[msgName]
}

// 根据Req消息的名称，返回其对应的Ack消息名称
func GetPairingAckName(reqName string) string {
	if strings.HasSuffix(reqName, "Req") {
		return reqName[:len(reqName)-3] + "Ack"
	}
	return ""
}

// 如果消息的名字是XXXReq，则尝试创建与其名称对应的XXXAck消息
func CreatePairingAckBy(reqName string) proto.Message {
	var ackName = GetPairingAckName(reqName)
	if ackName != "" {
		return CreateMessageByName(ackName)
	}
	return nil
}

// 如果消息的名字是XXXReq，则尝试创建与其名称对应的XXXAck消息
func CreatePairingAck(req proto.Message) proto.Message {
	var rtype = reflect.TypeOf(req).Elem()
	var fullname = rtype.String()
	return CreatePairingAckBy(fullname)
}
