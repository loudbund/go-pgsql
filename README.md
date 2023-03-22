# go-mysql
go-pool是基于database/sql和github.com/go-sql-driver/mysql提供的一组数据库快速操作的函数库，使用前需要作些准备
1. 需要数据库的配置文件
2. 初始化配置
3. 获取数据库句柄
4. 调用函数操作数据表

## 安装
go get github.com/loudbund/go-mysql

## 引入
```golang
import "github.com/loudbund/go-mysql/mysql_v1"
```

## 配置文件示例 test.conf
```db.conf
# 默认数据库
[db_default]
host        = 127.0.0.1     ; 数据库ip
port        = 3306          ; 数据库端口
db          = test          ; 数据库名
username    = root          ; 连接账号
password    = root123456    ; 连接密码
charset     = utf8          ; 编码: utf8/utf8mb4
maxIdle     = 7             ; 空闲连接数
maxConn     = 19            ; 最大连接数

# 指定数据库
[db_test]
host        = 127.0.0.1 
port        = 3306
db          = test
username    = root
password    = root123456
charset     = utf8       # utf8/utf8mb4
maxIdle     = 7
maxConn     = 19
```

## 指定配置文件和初始化
1. 可以直接在main.go的init里初始
2. 初始化的时候，所有配置了的数据库都会连接检测，连不上就抛出panic
```golang
func init() {
	mysql_v1.Init("test.conf")
}
```

## 获取数据库句柄
1. 使用默认配置 mysql_v1.Handle() , 将读取 [db_default] 段配置
2. 指定数据库配置  mysql_v1.Handle("test") , 将读取 [db_test] 段配置
3. 指定数据库配置，指定数据库  mysql_v1.Handle("test","user") , 将读取 [db_test] 段配置, **数据库名换成user库**
```golang
handle := mysql_v1.Handle()
handle1 := mysql_v1.Handle("test")
handle2 := mysql_v1.Handle("test", "user")
```

## 数据库常规操作-表内容调整 函数
Insert将返回自增id和异常，其他的只返回异常
```golang
mysql_v1.Handle().Insert
mysql_v1.Handle().InsertManyTransaction
mysql_v1.Handle().Update
mysql_v1.Handle().Replace
mysql_v1.Handle().Delete
```

## 数据库常规操作-数据检索 函数
1. 批量读取返回格式都是[]map[string]interface{}，
2. QueryTableOne读取单条数据，返回格式map[string]interface{}，未取到时为nil
```golang
mysql_v1.Handle().Query
mysql_v1.Handle().QueryRaw
mysql_v1.Handle().QueryTable
mysql_v1.Handle().QueryTableOne
```

## 表信息获取 函数
NameAllDbs返回的数据库过滤掉了 mysql、information_schema、test 三个库名
```golang
mysql_v1.Handle().NameAllDbs
mysql_v1.Handle().NameAllTablesOneDb
mysql_v1.Handle().ShowCreateTable
mysql_v1.Handle().DescTable
```

## 特殊函数 函数
```golang
mysql_v1.Handle().Exec
mysql_v1.Handle().QueryAllCircle
```
## 特殊函数 QueryAllCircle
快速遍历数据表的特殊封装，其原理是按主键排序快速取出数据，取数据的条件只有主键id，所以读取速度非常快，可以达到10万/秒
```golang

Len :=0
if err := mysql_v1.Handle().QueryAllCircle(mysql_v1.UFastQuery{
    Table:    "user",        // 数据表名称
    Fields:   "*",           // 读取字段，默认将加上主键字段
    PriField: "userid",      // 主键字段名
    PriSort:  "asc",         // 遍历顺序 (asc/desc)
    RowLimit: 2000,          // 单次读取行数，针对有大数据字段的表，该值适当减小
    // BeginVal:        3,   // 主键起点值，不设置则程序自动识别
    // beginValIgnore: true, // 是否包含主键起点值，默认不包含
}, func(V map[string]interface{}) bool {
    // 这里处理一条数据V
    Len++

    // 返回true则会继续回调下一条，false则终止回调
    return true
}); err != nil {
    log.Error(err)
}
```
## 关于 example.go
1. 示例代码运行，需要一个可操作的数据库。 请修改 test.conf 的 [db_defaut] 配置
2. 运行示例代码，将会在配置的数据库里创建一张 demo表，并产生测试数据
3. 执行 go run example.go 运行示例代码
4. 详见 example.go 源码
