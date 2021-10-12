# fatchoy

Golang Game Server Kit

Gung Hay Fat Choy(恭喜發財)

`go get gopkg.in/qchencc/fatchoy.v1`

## 各目录说明

  目录       |  描述
------------|------------
log         | 日志
debug       | 调试API
codec       | 编解码
codes       | 错误码
sched       | 执行器
secure      | 密码生成
packet      | 消息结构
qnet        | 网络通信
discovery   | 服务发现
x           | 工具包

#### 包名规范

* API接口除非`Must`否则不使用panic抛出错误；
* 使用下划线分隔的小写字母，并使用有意义的缩写
* 包名要见文知义，让人看名字就知道这个包大体用来干什么的，避免太通用的名称（如common,util,misc)

参考 [package names](https://go.dev/blog/package-names)


## 规范

### Git提交规范

```
<scope>: subject
<blank line>
<body>
```

* scope: 是本次commit涉及了哪些模块
* subject：本次commit的简单描述，仅一行
* body: 对本次 commit 的详细描述，可以分成多行

大体是参考[Angular的规范](doc/Git Commit Message Conventions.docx)，但不必全套照搬，
大致要注意以下几点：

* 提交的commit message要考虑reviewer的阅读体验，能让其迅速浏览完这一篇提交大概干了些什么；
* 提交的描述文字一律使用中文，不要这次提交写英语下次提交又是中文，非外企实践全英文环境必要性不强；



### 版本号管理

一个版本号（如1.0.1）由`major.minor.patch`三部分组成

* `major`: 主版本号
* `minor`: 次版本号
* `patch`: 修订号

一些约定：

* 第一个初始开发版本使用`0.1.0`
* 第一个可以对外发布的版本使用`1.0.0`

参考[semantic version](https://semver.org/lang/zh-CN/)
