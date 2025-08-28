# gameWeb 项目目录结构说明

## 📁 完整目录结构

```
gameWeb/
├── .git/                           # Git 版本控制
├── .gitignore                      # Git 忽略文件配置
├── README.md                       # 项目主要说明文档
├── go.mod                          # Go 模块依赖管理
├── go.sum                          # Go 依赖版本锁定
├── main.go                         # 主程序入口文件
├── pid.txt                         # 进程ID文件
│
├── 📂 app/                         # 应用程序代码
│   └── controller/                 # 控制器层
│       ├── adminController.go      # 管理员控制器
│       └── gameController.go       # 游戏控制器
│
├── 📂 config/                      # 配置管理
│   ├── config.go                   # 配置加载逻辑
│   └── config.example.yaml         # 配置文件模板
│
├── 📂 db/                          # 数据库连接
│   ├── mysql.go                    # MySQL 数据库连接
│   └── redis.go                    # Redis 数据库连接
│
├── 📂 docs/                        # 📚 项目文档目录
│   ├── README.md                   # 文档总览
│   ├── API_DOCUMENTATION.md        # 完整API文档
│   ├── API_SEPARATION.md           # API分离设计
│   ├── admin_update_api.md         # 管理员更新接口文档
│   ├── get_admin_info_api.md       # 获取管理员信息接口文档
│   ├── DATABASE_FIX.md             # 数据库修复记录
│   ├── DATABASE_STRUCTURE_FIX.md   # 数据库结构修复
│   ├── FEATURE_CHECKLIST.md        # 功能开发清单
│   ├── SECURITY_CLEANUP_COMPLETE.md # 安全清理记录
│   └── SECURITY_DEPLOYMENT.md      # 安全部署指南
│
├── 📂 log/                         # 日志系统
│   ├── ginLogger.go                # Gin 日志中间件
│   └── zapLogger.go                # Zap 日志配置
│
├── 📂 logs/                        # 日志文件存储
│   └── game_*.log                  # 按日期分割的日志文件
│
├── 📂 middleware/                  # 中间件
│   └── authMiddleware.go           # 认证中间件
│
├── 📂 models/                      # 数据模型
│   └── models.go                   # 数据结构定义
│
├── 📂 routes/                      # 路由定义
│   └── routes.go                   # 路由配置
│
├── 📂 sql/                         # 数据库表结构
│   ├── adminAccount.sql            # 管理员账户表
│   ├── logAuth.sql                 # 认证日志表
│   ├── logResult10001.sql          # 游戏结果日志表
│   ├── mailSystem.sql              # 邮件系统表
│   ├── mailUsers.sql               # 用户邮件表
│   ├── mails.sql                   # 邮件模板表
│   ├── userData.sql                # 用户数据表
│   ├── userRiches.sql              # 用户财富表
│   └── userStatus.sql              # 用户状态表
│
├── 📂 test/                        # 🧪 测试脚本目录
│   ├── README.md                   # 测试说明文档
│   ├── test_admin_update.sh        # 管理员更新接口测试
│   └── test_get_admin_info.sh      # 获取管理员信息接口测试
│
├── 📂 bin/                         # 编译后的二进制文件
│
└── 📄 运维脚本
    ├── build.sh                    # 构建脚本
    ├── deploy.sh                   # 部署脚本
    ├── run.sh                      # 运行脚本
    ├── stop.sh                     # 停止脚本
    └── set_env.example.sh          # 环境变量配置示例
```

## 📋 目录功能说明

### 核心应用目录

#### `app/controller/`
- **作用**: 业务逻辑控制器
- **文件**: 
  - `adminController.go` - 管理员相关接口
  - `gameController.go` - 游戏配置相关接口

#### `config/`
- **作用**: 配置文件管理
- **特点**: 支持环境变量覆盖，使用 Viper 库

#### `db/`
- **作用**: 数据库连接管理
- **支持**: MySQL 和 Redis 数据库

#### `middleware/`
- **作用**: HTTP 中间件
- **功能**: JWT 认证、日志记录、CORS 等

#### `models/`
- **作用**: 数据模型定义
- **包含**: 数据库模型、API 请求/响应模型

#### `routes/`
- **作用**: 路由配置
- **设计**: RESTful API 路由

### 文档与测试目录

#### `docs/` 📚
- **作用**: 项目技术文档
- **分类**: API文档、数据库文档、安全文档、项目管理文档
- **维护**: 随代码更新同步维护

#### `test/` 🧪
- **作用**: 自动化测试脚本
- **类型**: API 接口测试、功能测试
- **工具**: Bash + curl + jq

#### `sql/`
- **作用**: 数据库表结构定义
- **用途**: 数据库初始化、表结构参考

### 日志与运行时

#### `log/`
- **作用**: 日志处理逻辑
- **特点**: 结构化日志、按级别输出

#### `logs/`
- **作用**: 日志文件存储
- **规则**: 按日期分割、自动轮转

### 运维脚本

#### 构建部署
- `build.sh` - 编译项目
- `deploy.sh` - 部署到服务器
- `run.sh` - 本地开发运行
- `stop.sh` - 停止服务

## 🚀 使用指南

### 开发者
1. 查看 `README.md` 了解项目概况
2. 阅读 `docs/` 目录下的技术文档
3. 使用 `test/` 目录下的脚本进行测试

### 运维人员
1. 参考 `docs/SECURITY_DEPLOYMENT.md` 进行部署
2. 使用运维脚本进行日常管理
3. 查看 `logs/` 目录进行问题排查

### 项目管理
1. 查看 `docs/FEATURE_CHECKLIST.md` 了解开发进度
2. 参考 API 文档进行需求分析
3. 使用测试脚本进行质量验证

## 📝 维护建议

### 文档维护
- 新功能开发必须更新相应文档
- API 变更同步更新接口文档
- 定期检查文档的准确性

### 测试维护
- 新接口开发必须编写测试脚本
- 修改接口时同步更新测试用例
- 定期运行测试确保功能正常

### 代码组织
- 遵循现有的目录结构
- 按功能模块组织代码
- 保持代码和文档的同步

---

*文档创建时间: 2024-08-28*
*最后更新时间: 2024-08-28*