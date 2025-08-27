#!/bin/bash

# 环境变量配置脚本
# 使用方法: source set_env.sh

echo "Setting up environment variables for gameWeb..."

# MySQL 主数据库配置
export MYSQL_HOST="localhost"
export MYSQL_PORT="3306"
export MYSQL_USER="root"
export MYSQL_PASSWORD="your-password-here"
export MYSQL_DATABASE="game_db"

# gameWeb 数据库配置
export MYSQL_GAMEWEB_HOST="localhost"
export MYSQL_GAMEWEB_PORT="3306"
export MYSQL_GAMEWEB_USER="root"
export MYSQL_GAMEWEB_PASSWORD="your-password-here"
export MYSQL_GAMEWEB_DATABASE="gameWeb"

# gamelog 数据库配置（敏感）
export MYSQL_GAMELOG_HOST="localhost"           # 请更改为实际主机
export MYSQL_GAMELOG_PORT="3306"
export MYSQL_GAMELOG_USER="root"                # 请更改为实际用户名
export MYSQL_GAMELOG_PASSWORD="your-new-password"  # 请设置新密码
export MYSQL_GAMELOG_DATABASE="gamelog"

# Redis 配置
export REDIS_HOST="localhost"
export REDIS_PORT="6379"
export REDIS_PASSWORD=""
export REDIS_DATABASE="0"

# JWT 密钥
export JWT_SECRET="your-jwt-secret-key-here"
export ADMIN_JWT_SECRET="your-admin-jwt-secret-here"

# 游戏服务器配置
export GAMESERVER_HOST="localhost"
export GAMESERVER_PORT="9000"

echo "Environment variables set successfully!"
echo "Please update the actual values before using in production."
echo ""
echo "⚠️  重要提醒:"
echo "1. 请立即更改 MYSQL_GAMELOG_PASSWORD 为新密码"
echo "2. 请更新实际的主机地址和用户名"
echo "3. 请设置强密码和密钥"