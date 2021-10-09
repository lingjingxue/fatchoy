# X Kit

可以跨多个项目使用的基础工具包

### 各子包说明

包名        |  描述
------------|-----------------------------
cipher      | 加密解密
collections | 算法和数据结构
datetime    | 日期相关
fsutil      | 文件相关
mathext     | 数学扩展
reflectutil | 反射相关
strutil     | 字符串相关



### 本包规范

[包名规范细节](https://blog.golang.org/package-names)

* API接口除非`Must`否则不使用panic抛出错误；
* 包命名要简短并准确，不使用下划线、驼峰，勿使用太笼统的名字（如common, base, misc）；
* 本包的各子包之间不相互引用；
* 本包的各子包不过多依赖第三方外部包；
