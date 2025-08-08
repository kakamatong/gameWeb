#!/bin/bash

# 检查pid文件是否存在
if [ ! -f "pid.txt" ]; then
  echo "未找到PID文件，应用程序可能未运行"
  exit 1
fi

# 读取PID
PID=$(cat pid.txt)

# 检查进程是否存在
if ps -p $PID > /dev/null; then
  # 终止进程
  kill $PID
  echo "应用程序已停止，PID: $PID"
  
  # 删除pid文件
  rm -f pid.txt
else
  echo "进程 $PID 不存在，应用程序可能已停止"
  rm -f pid.txt
fi