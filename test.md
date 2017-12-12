# 测试报告

## 测试工具

|/|路径|描述|
|---|---|---|
|coap-server|coap/tools/coap-server|测试服务器|
|coap-curl|coap/tools/coap-curl|发送coap请求的工具|
|coap-mesg|coap/tools/coap-mesg|直接发送coap消息的工具|

测试流程：先启动coap-server，再运行`coap/tools/scripts`目录下的测试脚本。

## 协议测试

### 可靠请求测试

* 测试方法

```
$CURL -X POST --con --data 'ConRequest' coap://localhost/TestConRequest
```

* 预期输出

coap-curl输出
```
CON[true] POST coap://localhost:5683/TestConRequest
Uri-Host: localhost
Uri-Path: TestConRequest

ConRequest
ACK[true] Content

ConRequest
```

coap-server输出
```
CON[true] POST coap://localhost:5683/TestConRequest
Uri-Host: localhost
Uri-Path: TestConRequest

ConRequest
```

### 非可靠请求测试

* 测试方法

```
$CURL -X POST --data 'NonRequest' coap://localhost/TestNonRequest
```

* 预期输出

coap-curl输出
```
CON[false] POST coap://localhost:5683/TestNonRequest
Uri-Host: localhost
Uri-Path: TestNonRequest

NonRequest
ACK[false] Content

NonRequest
```

coap-server输出
```
CON[false] POST coap://localhost:5683/TestNonRequest
Uri-Host: localhost
Uri-Path: TestNonRequest

NonRequest
```

## 功能测试

### 去重及可靠性测试

* 测试方法

```
$CURL --verbose 2 -X POST --con --data '5s' coap://localhost/TestDeduplication
```

* 预期输出

coap-curl输出
```
CON[true] POST coap://localhost:5683/TestDeduplication
Uri-Host: localhost
Uri-Path: TestDeduplication

5s
2017/12/10 16:56:22.091865 send: Confirmable,POST,63555,0926d51dfe0244e1
Uri-Host: localhost
Uri-Path: TestDeduplication
2017/12/10 16:56:25.090696 send: Confirmable,POST,63555,0926d51dfe0244e1
Uri-Host: localhost
Uri-Path: TestDeduplication
2017/12/10 16:56:27.093624 recv: Acknowledgement,Content,63555,0926d51dfe0244e1
ACK[true] Content

count=1
```

coap-server输出
```
2017/12/10 16:56:22 [TestDeduplication] start: count=1
2017/12/10 16:56:27 [TestDeduplication] end: count=1
```

### Blockwise测试

* 测试方法

```
$CURL -X POST --con --in-file ietf-block.html --out-file output.html coap://localhost/TestBlock
md5sum ietf-block.html output.html
```

* 预期输出

md5sum输出
```
3c8b5de7b1583d3790d70cf92fa211e4  ietf-block.html
3c8b5de7b1583d3790d70cf92fa211e4  output.html
```

## 异常测试

### 异常消息码测试

* 测试方法

```
$MESG --read --code 6
```

* 预期输出

coap-mesg输出
```
coap server: localhost:5683
Confirmable,Unknown (6.00),0
Reset,Unknown (0.00),0
```

### 异常选项测试

#### 不可识别的重要选项

* 测试方法

```
$MESG --read --code 0.01 --empty-option "9"
```

* 预期输出

coap-mesg输出
```
coap server: localhost:5683
Confirmable,GET,0
9: <nil>

Acknowledgement,BadOption,0
Unrecognized options of class "critical" that occur in a Confirmable request
```

#### 不可识别的非重要选项

* 测试方法

```
$MESG --read --code 0.01 --empty-option "10"
```

* 预期输出

coap-mesg输出
```
coap server: localhost:5683
Confirmable,GET,0
10: <nil>

Acknowledgement,NotFound,0
"/" path not found
```
