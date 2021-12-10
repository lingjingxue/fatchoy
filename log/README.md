# 日志API设计


### 日志API需求

1. 日志分级;
2. 日志存档（文件，ELK等）;
3. 日志API


### 日志API实现

1. qlog包提供一个对外部稳定的API，外部不必关注包内部是如何实现（内部目前使用uber zap）;
2. 业务package大部分场合只需要import qlog即可使用日志，不必在new其它对象；
3. 通过不同的API名称实现日志分级

接口        |  作用
------------|------------
qlog.Setup()  | 设置参数
qlog.Debugf() | 调试日志
qlog.Infof() | 信息日志
qlog.Warnf() | 警告日志
qlog.Errorf() | 错误日志
qlog.Fatalf() | 错误日志，程序不能继续执行
