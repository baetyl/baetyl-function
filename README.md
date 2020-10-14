baetyl-function
========

[![Build Status](https://travis-ci.org/baetyl/baetyl-function.svg?branch=master)](https://travis-ci.org/baetyl/baetyl-function)
[![Go Report Card](https://goreportcard.com/badge/github.com/baetyl/baetyl-function)](https://goreportcard.com/report/github.com/baetyl/baetyl-function) 
[![codecov](https://codecov.io/gh/baetyl/baetyl-function/branch/master/graph/badge.svg)](https://codecov.io/gh/baetyl/baetyl-function)
[![License](https://img.shields.io/github/license/baetyl/baetyl-function.svg)](./LICENSE)

## 简介

[Baetyl-function](https://github.com/baetyl/baetyl-function) 是 [Baetyl](https://github.com/baetyl/baetyl) 框架端侧的函数计算框架。端侧函数计算框架由前端代理和后端函数运行时两部分组成。

[Baetyl-function](https://github.com/baetyl/baetyl-function) 提供前端代理，是函数入口，通过暴露 HTTP 接口供其他服务调用，然后将请求透传给后端的函数运行时模块。

后端函数运行时提供多种选择：

- [baetyl-function-python36](https://github.com/baetyl/baetyl-function/tree/master/python36) 提供 Python3.6 函数运行时；
- [baetyl-function-node10](https://github.com/baetyl/baetyl-function/tree/master/node10) 提供 Node 10 函数运行时；
- baetyl-function-sql 提供 SQL 函数运行时，兼容 SQL92 语法。

用户可以编写 python、node、sql 脚本来构建自己的业务逻辑，进行消息的过滤、转换和转发等，使用非常灵活。

函数计算基于事件驱动编程模型设置。用户可以使用 Baetyl 端侧规则引擎 [baetyl-rule](https://github.com/baetyl/baetyl-rule) 模块设定规则，当由消息触发某条规则的时候，可以在该条规则内调用相关函数。Baetyl-rule 会按照如下格式请求 baetyl-function 模块：

```
https://[baetyl-function-service]/[function-service]/[function]
```

其中 `baetyl-function-service` 是 baetyl-function 的服务地址，例如 baetyl-function:50011，`function-service` 是后端函数运行时服务的名称，`function` 表示函数入口，如果 `funciton` 字段不指定的话，后端函数运行时会默认选择自身函数列表中的第一个函数。函数入口表示执行函数，对于 Python/Node 运行时来说，由函数脚本和处理函数名组成，对于 SQL 运行时来说，只有函数脚本组成，函数脚本内即是用户编写的 sql 语句。

具体使用可以参考最佳实践 [Baetyl 边缘规则引擎实践](https://github.com/baetyl/baetyl-docs-cn/blob/master/docs/practice/message-rule-practice.md) 。

## 配置

[Baetyl-function](https://github.com/baetyl/baetyl-function) 的全量配置文件如下，并对配置字段做了相应解释：

```yaml
server: # server 相关设置，由于把请求代理到后端的 Runtimes 模块
  address: ":50011" # 监听地址
  concurrency: # 服务端并发连接数，如果不设置的话将使用默认值
  disableKeepalive: true # 是否启用 keep-alive 连接,默认值为 false
  tcpKeepalive: false # 是否主动发送 keep-alive 消息，默认为 false
  maxRequestBodySize: # Body 最大数据量，默认为 4 * 1024 * 1024 Byte
  readTimeout: 1h # 服务端连接的读超时时间，默认值为无限制
  writeTimeout: 1h # 服务端连接的写超时时间，默认值为无限制
  idleTimeout: 1h # 在 keep alive 启动条件下，服务端等待下次消息的空闲超时时间，如果值为0，复用读超时时间
  ca: example/var/lib/baetyl/testcert/ca.crt # Server 的 CA 证书路径
  key: example/var/lib/baetyl/testcert/server.key # Server 的服务端私钥路径
  cert: example/var/lib/baetyl/testcert/server.crt # Server 的服务端公钥路径

client: # 请求后端 Runtimes 模块的客户端相关设置
  grpc: # Grpc 客户端设置
    port: 80 # 后端 Runtimes 端口
    timeout: 5m # 请求超时时间
    retries: 3 # 请求重试次数

logger: # 日志
  level: info # 日志等级
```