#!/bin/bash

# 确保bin目录存在且应用程序可执行
if [ ! -x "bin/gameWeb" ]; then
  echo "错误: 应用程序不存在或不可执行，请先构建"
  exit 1
fi

# 检查是否已有实例在运行
if [ -f "pid.txt" ]; then
  PID=$(cat pid.txt)
  if ps -p $PID > /dev/null; then
    echo "应用程序已在运行，PID: $PID"
    exit 1
  else
    echo "发现旧的PID文件，但进程不存在，删除旧文件"
    rm -f pid.txt
  fi
fi

# 设置环境变量禁用日志文件
export GAMEWEB_LOG_PATH=""

# 启动应用程序，使用nohup在后台运行，输出重定向到/dev/null
nohup ./bin/gameWeb > /dev/null 2>&1 &

# 记录进程ID
echo $! > pid.txt

echo "应用程序已启动，PID: $(cat pid.txt)"
echo "日志已禁用，不会生成日志文件"