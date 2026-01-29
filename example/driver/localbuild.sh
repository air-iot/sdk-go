#!/bin/bash

# 设置编译目标为Linux amd64
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64

# 编译Go程序，这里假设你的main.go位于cmd/ai/main.go，且你想给编译的程序设置版本号为4.0.0
go build -tags netgo -v -ldflags "-X main.VERSION=4.0.0" -o sdk main.go

# 压缩编译后的程序，这里假设编译后的程序名为main
# 确保你已经在mac上安装了upx，如果没有安装，可以通过brew install upx安装
#upx -4 -k main
#docker run --rm --name upx -v ./main:/main airiot/upx:4.0.2 upx -k -4 /main
#scp sdk sky@192.168.124.139:/home/sky
#scp -r etc sky@192.168.124.139:/home/sky


