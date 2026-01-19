#第一阶段：构建阶段 (Builder)
#使用官方 Go 语言镜像作为基础，起个别名叫 builder
# 把 1.21 改成 latest
FROM golang:latest AS builder

#设置 Go 模块代理，这步决定下载依赖快不快
ENV GOPROXY=https://goproxy.cn,direct
#设置容器内的工作目录
WORKDIR /app

#1. 先复制依赖文件 (利用 Docker 缓存机制，加快构建速度)
COPY go.mod go.sum ./
#下载依赖
RUN go mod download

#2. 复制剩下的所有源代码
COPY . .

# 3. 编译 Go 程序
#-o mud_server 指定输出文件名为 mud_server
RUN CGO_ENABLED=0 GOOS=linux go build -o mud_server cmd/server/main.go


#第二阶段：运行阶段 (Runner)
#使用一个超小的 Linux 镜像 (Alpine)，为了让最终镜像尽可能小
FROM alpine:latest

#设置工作目录
WORKDIR /root/

#从第一阶段把编译好的 mud_server 文件复制过来
COPY --from=builder /app/mud_server .

#如果项目有配置文件（比如 config.yaml），记得也要复制
#COPY --from=builder /app/config.yaml . 

#把 gamedata 文件夹也从宿主机（你的电脑）复制进容器
COPY gamedata ./gamedata
#暴露端口 (虽然只是文档作用，但是个好习惯)
EXPOSE 8888

#容器启动时执行的命令
CMD ["./mud_server"]