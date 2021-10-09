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

const (
	PkgName = "pbapi"
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
func hasValidSuffix(name string) bool {
	nameSuffix := []string{"Req", "Ack", "Ntf"}
	for _, suffix := range nameSuffix {
		if strings.HasSuffix(name, suffix) {
			return true
		}
	}
	return false
}

func isWellKnown(name string) bool {
	return strings.HasPrefix(name, "google/") ||
		strings.HasPrefix(name, "github.com/") ||
		strings.HasPrefix(name, "grpc/")
}

func isNil(c interface{}) bool {
	if c == nil {
		return true
	}
	return reflect.ValueOf(c).Kind() == reflect.Ptr && reflect.ValueOf(c).IsNil()
}

// 从message的option里获取消息ID
func getMsgIdByExtension(descriptor protoreflect.MessageDescriptor, xtName protoreflect.FullName) int32 {
	ovi := descriptor.Options()
	if isNil(ovi) {
		return 0
	}
	omi := ovi.ProtoReflect()
	var msgId int32
	omi.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		if !fd.IsExtension() {
			return true
		}
		if fd.FullName() == xtName {
			ivs := v.String()
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
	if isWellKnown(fd.Path()) || fd.Package() != PkgName {
		return true
	}
	log.Printf("register %s %s\n", fd.Package(), fd.Path())
	msgDescriptors := fd.Messages()
	for i := 0; i < msgDescriptors.Len(); i++ {
		descriptor := msgDescriptors.Get(i)
		fullname := descriptor.FullName()
		if !hasValidSuffix(string(fullname)) {
			continue
		}
		name := string(fullname)
		msgid := getMsgIdByExtension(descriptor, xtName)
		if msgid == 0 {
			continue
		}
		rtype := proto.MessageType(string(fullname))
		if s, found := msgIdNames[msgid]; found {
			log.Panicf("duplicate message hash %s %s, %d", s, name, msgid)
		}
		msgTypeRegistry[name] = rtype.Elem()
		msgNameIds[name] = msgid
		msgIdNames[msgid] = name
	}
	return true
}

func RegisterV2(exName string) {
	fullname := protoreflect.FullName(exName)
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
	rtype := reflect.TypeOf(msg)
	fullname := rtype.String()
	if fullname == "" {
		return 0
	}
	for fullname[0] == '*' {
		fullname = fullname[1:]
	}
	return msgNameIds[fullname]
}

// 根据Req消息的ID，返回其对应的Ack消息ID
func GetPairingAckID(reqId int32) int32 {
	var reqName = msgIdNames[reqId]
	if !strings.HasSuffix(reqName, "Req") {
		return 0
	}
	var ackName = reqName[:len(reqName)-3] + "Ack"
	return msgNameIds[ackName]
}
