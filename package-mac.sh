#! /bin/bash
## mac下编译linux平台的64位可执行程序
echo "start to package rds4go, please waiting..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o rds4go main.go
echo "package rds4go finished"