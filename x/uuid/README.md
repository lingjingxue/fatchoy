# uuid

分布式ID生成


# Usage

```go
// 初始化传入workerID, etcd地址，和命名空间前缀
uuid.Init(workerID, "localhost:2379", "/keyspace/")

// 返回一个ID，使用发号器算法，可用于角色、工会ID
id := NextID()

// 返回一个UUID，使用雪花算法，可用于事件行为、日志ID，通常值比较大
uuid := NextUUID()

// GUID string, standard universally unique identifier version 4
guid := NextGUID()

```

### 雪花算法

  时间戳 | WorkerID | SeqID
--------|--------|-------

* 把一个64位的整数分成3个区间，分别用于`时间戳`、`workerID`和`seqID`；
* 时间戳代表时间单元需要一直递增，不同的时间单元实现有毫秒、秒、厘秒等，这里依赖时钟不回调；
* `workerID`用来标识分布式环境下的每个service；
* `seqID`在一个时间单元内持续自增，如果在单个时间单元内`seqID`溢出了，需要sleep等待进入下一个时间单元；

不同的雪花算法实现的差异主要集中在3个区间的分配，和workerID自动还是手动分配上。

* sony的实现 `https://github.com/sony/sonyflake`
* 百度的实现 `https://github.com/baidu/uid-generator`
* 美团的实现 `https://github.com/Meituan-Dianping/Leaf`

### 发号器算法

* 算法把一个64位的整数按step划分为N个号段；
* 每个service向发号器申请领取一个代表号段的counter；
* service内部使用这个号段向业务层分配ID
* service重启或者号段分配完，向发号器申请下一个号段

发号器依赖存储组件，对存储组件的需求是能实现整数自增

本包提供多种存储选择，etcd、redis和mongodb

* etcd使用单个key的revision来实现自增
* redis使用incr命令实现自增
* mongodb使用findOneAndUpdate
