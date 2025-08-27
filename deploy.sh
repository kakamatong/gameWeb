#!/bin/bash

# 游戏管理后台部署脚本

echo "==================== 游戏管理后台部署脚本 ===================="

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "错误: Go环境未安装，请先安装Go 1.21+"
    exit 1
fi

echo "✓ Go环境检查通过"

# 检查配置文件
if [ ! -f "config/config.yaml" ]; then
    echo "警告: 配置文件 config/config.yaml 不存在，使用默认配置"
fi

# 创建日志目录
mkdir -p logs
echo "✓ 日志目录已创建"

# 编译项目
echo "正在编译项目..."
go build -o gameWeb main.go

if [ $? -ne 0 ]; then
    echo "错误: 编译失败"
    exit 1
fi

echo "✓ 项目编译成功"

# 检查数据库连接（可选）
echo "==================== 部署检查清单 ===================="
echo "请确保以下条件满足："
echo "1. MySQL服务器已启动"
echo "2. gameWeb库已创建，并执行了 sql/adminAccount.sql"
echo "3. game库已创建，并执行了用户相关SQL文件"
echo "4. gamelog库连接权限正常"
echo "5. Redis服务器已启动"
echo "6. 配置文件config/config.yaml已正确配置"
echo ""

echo "==================== 启动选项 ===================="
echo "1. 直接启动服务"
echo "2. 后台启动服务"
echo "3. 仅编译，不启动"
echo "4. 查看配置信息"

read -p "请选择操作 (1-4): " choice

case $choice in
    1)
        echo "正在启动服务..."
        ./gameWeb
        ;;
    2)
        echo "正在后台启动服务..."
        nohup ./gameWeb > logs/gameWeb.log 2>&1 &
        echo $! > pid.txt
        echo "✓ 服务已后台启动，PID: $(cat pid.txt)"
        echo "日志文件: logs/gameWeb.log"
        ;;
    3)
        echo "✓ 编译完成，可执行文件: ./gameWeb"
        ;;
    4)
        echo "==================== 配置信息 ===================="
        echo "服务端口: 8080"
        echo "配置文件: config/config.yaml" 
        echo "日志目录: logs/"
        echo "API文档: API_DOCUMENTATION.md"
        echo ""
        echo "主要API端点:"
        echo "- 管理员登录: POST /api/admin/login"
        echo "- 用户管理: GET /api/admin/users"
        echo "- 日志查询: GET /api/admin/logs/auth"
        echo "- 邮件发送: POST /api/admin/mails/send"
        ;;
    *)
        echo "无效选择"
        exit 1
        ;;
esac

echo ""
echo "==================== 部署完成 ===================="
echo "API文档: 查看 API_DOCUMENTATION.md"
echo "如需停止服务: 运行 ./stop.sh"