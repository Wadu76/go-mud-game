***Docker 操作手册***
这是以后的日常：

- 一键启动/更新代码 (核心命令)： 当你改了 Go 代码后，运行这个命令，它会重新编译并重启服务器，但保留数据库数据。

```Bash

docker compose up -d --build
```
- 查看服务器报错/日志：

```Bash

docker compose logs -f mud-server
```

- 彻底重置 (删库跑路)： 如果你想清空所有玩家数据从头开始：

```Bash

docker compose down -v
```