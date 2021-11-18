// Copyright © 2021-present ichenq@outlook.com All rights reserved.
// Distributed under the terms and conditions of the BSD License.
// See accompanying files LICENSE.

package codes

// 通用错误码定义(0-100)
type Code int32

const (
	OK                    Code = 0
	Unknown               Code = 1  // 未知错误
	Canceled              Code = 2  // 操作被取消
	Aborted               Code = 3  // 操作被终止
	Unavailable           Code = 4  // 服务不可用
	DeadlineExceeded      Code = 5  // 过期超时
	InvalidArgument       Code = 6  // 参数错误
	OutOfRange            Code = 7  // 请求数据超出限制
	RequestTimeout        Code = 8  // 请求超时
	BadRequest            Code = 9  // 错误的请求
	NotFound              Code = 10 // 请求资源未找到
	PermissionDenied      Code = 11 // 权限不足
	Unauthenticated       Code = 12 // 未认证
	ResourceExhausted     Code = 13 // 资源已耗尽
	NotImplemented        Code = 14 // 请求未实现
	AlreadyExists         Code = 15 // entity已经存在
	OperationNotSupported Code = 16 // 不支持的操作
	OperationTooFrequent  Code = 17 // 操作过于频繁
	TransportFailure      Code = 18 // 传输错误
	DataLoss              Code = 19 // unrecoverable data loss or corruption
	DatabaseError         Code = 20 // 数据库异常
	RuntimeException      Code = 21 // 运行时错误
	ServiceMaintenance    Code = 22 // 服务维护中
	InternalError         Code = 23 // 内部错误
)

func (c Code) String() string {
	if s, found := codeName[int32(c)]; found {
		return s
	}
	return "??"
}

var codeName = map[int32]string{
	0:  "OK",
	1:  "UNKNOWN",
	2:  "CANCELED",
	3:  "ABORTED",
	4:  "UNAVAILABLE",
	5:  "DEADLINE_EXCEEDED",
	6:  "INVALID_ARGUMENT",
	7:  "OUT_OF_RANGE",
	8:  "REQUEST_TIMEOUT",
	9:  "BAD_REQUEST",
	10: "NOT_FOUND",
	11: "PERMISSION_DENIED",
	12: "UNAUTHENTICATED",
	13: "RESOURCE_EXHAUSTED",
	14: "NOT_IMPLEMENTED",
	15: "ALREADY_EXISTS",
	16: "OPERATION_NOT_SUPPORTED",
	17: "OPERATION_TOO_FREQUENT",
	18: "TRANSPORT_FAILURE",
	19: "DATA_LOSS",
	20: "DATABASE_ERROR",
	21: "RUNTIME_EXCEPTION",
	22: "SERVICE_MAINTENANCE",
	23: "INTERNAL_ERROR",
}

var (
	maxCode   int32
	codeValue = buildCodeValue()
)

func buildCodeValue() map[string]int32 {
	var codeValue = make(map[string]int32, len(codeName))
	for k, v := range codeName {
		codeValue[v] = k
		if k > maxCode {
			maxCode = k
		}
	}
	return codeValue
}

func GetCodeName(code Code) string {
	return codeName[int32(code)]
}

func GetCodeValue(name string) Code {
	if code, found := codeValue[name]; found {
		return Code(code)
	}
	return -1
}
