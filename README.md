# WaduWorld (Go-MUD): 分布式架构的多人在线文字网游

![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)
![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)
![MySQL](https://img.shields.io/badge/MySQL-8.0-4479A1?style=flat&logo=mysql)
![License](https://img.shields.io/badge/License-MIT-green)

> 一个基于 Golang 原生 TCP Socket 实现的高并发 MUD (Multi-User Dungeon) 游戏引擎。
> 摒弃了传统的 Telnet 纯文本流，采用 **Client-Server 架构**，客户端实现了基于 **TUI (终端图形界面)** 的富交互体验。

## 项目演示 (Screenshots)

| **战斗与死亡系统 (Combat & Death)** | **背包与物品系统 (Inventory GUI)** | **日志系统** |
|:---:|:---:|:---:|
| <img src="/Server/images/死亡复活系统.png" width="400" alt="Death Screen"> | <img src="/Server/images/背包系统.png" width="400" alt="Inventory"> | <img src="/Server/images/日志交流.png" width="400" alt="Log System"> |
| *支持实时战斗日志流与全屏死亡反馈* | *支持表格化渲染与可视化装备管理* | *支持多玩家日志流与日志持久化* |

## 技术架构 (Architecture)

本项目采用经典的分层架构设计，实现了前后端分离与容器化部署。

* **服务端 (Server)**:
    * **核心语言**: Go (Golang)
    * **并发模型**: Goroutine-per-Connection + Channel 广播机制
    * **网络协议**: 自定义应用层协议 (`|CMD:OP_CODE:PAYLOAD`) 解决 TCP 粘包问题
    * **持久化**: MySQL 8.0 + GORM (ORM)
    * **同步机制**: `sync.Mutex` 保证怪物/玩家状态的并发安全
* **客户端 (Client)**:
    * **UI 框架**: Bubble Tea (基于 ELM 架构的 TUI 框架)
    * **渲染引擎**: Lipgloss (终端 CSS 风格渲染)
    * **特性**: 实现了状态管理 (State Management)、异步网络监听、动态血条渲染
* **运维 (DevOps)**:
    * **Docker**: 多阶段构建 (Multi-stage Build) 压缩镜像体积
    * **Docker Compose**: 一键编排 App 与 DB 容器，实现环境一致性

##  核心特性 (Key Features)

1.  **高性能并发**: 利用 Go 协程轻量级特性，为每个连接维护独立 Goroutine，支持多人同屏实时战斗与聊天。
2.  **富客户端体验 (Rich TUI)**:
    * **动态 HUD**: 实时解析服务端 `|CMD:HP` 协议，动态渲染血条。
    * **可视化背包**: 拦截 `|CMD:INC` JSON 数据，渲染交互式表格界面。
    * **沉浸式反馈**: 实现了全屏死亡警告（红色弹窗）与操作指引。
3.  **完备的游戏循环**:
    * 登录/注册/鉴权系统。
    * 基于房间 (Room) 的地图移动系统。
    * 基于所在区域 (Room) 一致的玩家交流系统。
    * 战斗系统 (Attack/Heal) 与怪物 AI 反击。
    * 物品系统 (Pick/Drop/Equip/Unequip) 与属性加成计算。
4.  **工程化落地**: 完整的 Docker 化部署方案，支持开箱即用。

## 快速开始 (Quick Start)

### 前置要求
* 安装 [Docker Desktop](https://www.docker.com/)

### 1. 一键启动服务器
无需安装 Go 或 MySQL，直接运行：

```bash
#构建镜像并启动容器(Server + MySQL)
docker compose up -d --build

```

### 2. 运行客户端

在另一个终端窗口中运行客户端：

```bash
#进入客户端目录运行
go run cmd/client/main.go

```

### 3. 游戏指令

* `register <name> <pwd>`: 注册
* `login <name> <pwd>`: 登录
* `attack`: 攻击当前房间的怪物
* `status`: 查看自己与boss的状态
* `say <内容>`: 在当前房间广播内容
* `inventory`: 打开可视化背包
* `look`: 查看周围环境
* `heal`: 治疗自己
* `pick <地上的物品名>`: 捡起地上的物品
* `drop <身上的物品名>`: 丢下身上的物品
*  `equip <装备名>`: 装上装备
*  `unequip <装备名>`: 卸下装备
* `talk <NPC名称> <想说的话>`: 与NPC对话
* `go north/south/east/west`: 移动
* `save`: 保存游戏（当然退出的时候也会自动保存）
* `exit`: 退出游戏

## 目录结构

```text
WaduWorld/
├── cmd/
│   ├── client/     # TUI 客户端入口 (Bubble Tea)
│   └── server/     # 游戏服务端入口
├── game/           # 核心游戏逻辑 (Player, Monster, Item, Logic)
├── network/        # 网络层封装 (TCP Handler, Protocol, World Broadcast)
├── database/       # 数据库连接与 ORM 模型
├── gamedata/       # 静态配置文件 (JSON)
├── Dockerfile      # 服务端镜像构建文件
├── images/         # 游戏图片 
└── docker-compose.yml # 容器编排文件

```

---

*Created by Wadu76 - 2026*



