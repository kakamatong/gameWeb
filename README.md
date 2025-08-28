# gameWeb

游戏Web服务，提供外部的WebAPI，如游戏配置拉取，游戏邮件拉取。

请求验签在middleware中处理，采用JWT验证，可根据路由自行配置是否要验签。

## 📁 项目结构

```
gameWeb/
├── app/controller/     # 控制器层
├── config/             # 配置管理
├── db/                 # 数据库连接
├── docs/               # 📚 项目文档
├── log/                # 日志系统
├── middleware/         # 中间件
├── models/             # 数据模型
├── routes/             # 路由定义
├── sql/                # 数据库表结构
├── test/               # 🧪 测试脚本
└── main.go             # 主程序入口
```

## 🚀 快速开始

### 环境要求
- Go 1.24.1+
- MySQL 5.7+
- Redis 6.0+

### 安装与运行
```bash
# 1. 安装依赖
go mod download

# 2. 配置数据库
cp config.example.yaml config.yaml
# 编辑 config.yaml 配置数据库连接

# 3. 启动服务
./run.sh
```

## 📚 文档

- [📋 完整文档列表](./docs/README.md)
- [🔌 API接口文档](./docs/API_DOCUMENTATION.md)
- [🔐 安全部署指南](./docs/SECURITY_DEPLOYMENT.md)

## 🧪 测试

- [🧪 测试脚本说明](./test/README.md)
- [⚡ 快速测试](./test/)

## 🛠️ 开发

- 构建: `go build -o gameWeb main.go`
- 部署: `./deploy.sh`
- 停止: `./stop.sh`

---

*详细信息请参考 [docs](./docs/) 目录中的文档*
