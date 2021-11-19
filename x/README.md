# X Kit

一些基础工具包

### 各子包说明

包名        |  描述
------------|-----------------------------
cipher      | 加密解密
collections | 算法和数据结构
datetime    | 日期相关
fsutil      | 文件相关
mathext     | 数学扩展
reflext     | 反射相关
strutil     | 字符串相关
geom        | 几何算法
uuid        | 分布式ID




### 本包规范

[包名规范细节](https://blog.golang.org/package-names)

* API接口除非`Must`否则不使用panic抛出错误；
* 本包的各子包之间不相互引用；
* 本包的各子包不过多依赖第三方外部包；
* 包名要见文知义，让人看名字就知道这个包大体用来干什么的，避免太通用的名称（如common,util,misc)
