#!/bin/bash

# 确保bin目录存在
mkdir -p bin

# 执行go build，将输出文件放到bin目录
# -o 参数指定输出文件路径和名称
# 假设可执行文件名为gameWeb
GOOS=linux GOARCH=amd64 go build -o bin/gameWeb main.go

# 检查构建是否成功
if [ $? -eq 0 ]; then
  echo "构建成功！可执行文件已生成到bin目录"
  ls -l bin/
else
  echo "构建失败！"
  exit 1
fi