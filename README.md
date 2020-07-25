# Respberry Pi

​		

树莓派的高级玩法，以钉钉作为入口和树莓派进行交互，使用到emqx作为消息中间件。go开发服务发布端，python来开发树莓派订阅端。

程序可以运行在任何平台。

- 钉钉机器人webhook信息处理。
- emq消息订阅发布。

## 环境

- emqx服务器。

## 下载源码

```shell
git clone https://github.com/ranzhendong/respberry.git
```

## 编译

```shell
go bulid -o respberry main.go
```

## 启动

```shell
./respberry
```

