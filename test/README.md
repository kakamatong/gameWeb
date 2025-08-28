# gameWeb 测试脚本

本目录包含 gameWeb 项目的所有自动化测试脚本。

## 📋 测试脚本分类

### API 接口测试
- [`test_admin_update.sh`](./test_admin_update.sh) - 管理员信息更新接口测试
- [`test_get_admin_info.sh`](./test_get_admin_info.sh) - 获取管理员信息接口测试
- [`test_auth_logs_api.sh`](./test_auth_logs_api.sh) - 登入认证日志API测试
- [`test_game_logs_api.sh`](./test_game_logs_api.sh) - 对局结果日志API测试
- [`test_all_logs_api.sh`](./test_all_logs_api.sh) - 综合日志API测试套件

## 🚀 测试脚本使用

### 环境要求
- Linux/macOS 环境
- curl 命令行工具
- jq JSON处理工具
- 运行中的 gameWeb 服务

### 安装依赖
```bash
# Ubuntu/Debian
sudo apt-get install curl jq

# macOS
brew install curl jq

# CentOS/RHEL
sudo yum install curl jq
```

### 运行测试

#### 单个接口测试
```bash
# 进入项目根目录
cd /root/gameWeb

# 测试管理员信息更新接口
./test/test_admin_update.sh

# 测试获取管理员信息接口
./test/test_get_admin_info.sh

# 测试登入认证日志API
./test/test_auth_logs_api.sh

# 测试对局结果日志API
./test/test_game_logs_api.sh

# 运行综合日志API测试套件
./test/test_all_logs_api.sh
```

#### 批量测试
```bash
# 运行所有API测试
for test_file in test/*.sh; do
    echo "运行测试: $test_file"
    ./"$test_file"
    echo "------------------------"
done
```

## 📊 测试内容说明

### test_admin_update.sh
测试管理员信息更新接口的完整功能：
- ✅ 管理员登录获取token
- ✅ 正常更新个人信息
- ✅ 部分字段更新测试
- ✅ 邮箱格式验证测试
- ✅ 权限控制测试
- ✅ 身份验证测试
- ✅ 更新结果验证

### test_get_admin_info.sh
测试获取管理员信息接口的功能：
- ✅ 管理员登录获取token
- ✅ 获取完整管理员信息
- ✅ 字段完整性验证
- ✅ 数据类型验证
- ✅ 身份验证测试
- ✅ JWT令牌验证测试

### test_auth_logs_api.sh
测试登入认证日志查询接口的完整功能：
- ✅ 管理员登录获取token
- ✅ 获取所有用户登录日志
- ✅ 按用户ID过滤查询
- ✅ 按时间范围过滤查询
- ✅ 分页参数测试
- ✅ 用户登录统计查询
- ✅ 参数验证测试
- ✅ 权限认证测试
- ✅ 性能测试
- ✅ 数据格式验证

### test_game_logs_api.sh
测试对局结果日志查询接口的完整功能：
- ✅ 管理员登录获取token
- ✅ 获取所有用户对局日志
- ✅ 按用户ID过滤查询
- ✅ 按时间范围过滤查询
- ✅ 复合条件查询测试
- ✅ 用户对局统计查询
- ✅ 数据一致性验证
- ✅ 权限认证测试
- ✅ 并发查询测试
- ✅ 数据格式验证

### test_all_logs_api.sh
综合日志API测试套件，包含：
- ✅ 登入认证日志API测试
- ✅ 对局结果日志API测试
- ✅ 统计API测试
- ✅ 数据一致性测试
- ✅ 权限认证测试
- ✅ 性能测试
- ✅ 自动化测试报告生成

## 🔧 测试配置

### 服务器配置
默认测试配置：
- **服务器地址**: `http://localhost:8080`
- **API路径**: `/api/admin`

如需修改，请编辑脚本中的配置变量：
```bash
BASE_URL="http://localhost:8080"
API_ENDPOINT="/api/admin"
```

### 测试账户
测试脚本使用以下默认账户：
- **用户名**: `admin`
- **密码**: `password123`

请确保数据库中存在该测试账户，或修改脚本中的登录信息。

## 📝 测试结果

### 成功标识
- ✅ `绿色对勾` - 测试通过
- ○ `黄色圆圈` - 可选项未设置
- ℹ️ `蓝色信息` - 信息提示

### 失败标识
- ✗ `红色叉号` - 测试失败
- ⚠️ `黄色警告` - 警告信息

### 示例输出
```
=== 管理员信息更新接口测试 ===

测试1: 管理员登录获取token
✓ 登录成功，获取到token: eyJhbGciOiJIUzI1NiIs...

测试2: 更新自己的基本信息
✓ 更新成功

测试3: 只更新部分字段（部门和头像）
✓ 部分更新成功
```

## 🔍 故障排除

### 常见问题

#### 1. curl命令未找到
```bash
# 解决方案
sudo apt-get install curl  # Ubuntu/Debian
brew install curl          # macOS
```

#### 2. jq命令未找到
```bash
# 解决方案
sudo apt-get install jq    # Ubuntu/Debian
brew install jq            # macOS
```

#### 3. 连接服务器失败
- 检查服务器是否正在运行：`ps aux | grep gameWeb`
- 检查端口是否被占用：`netstat -an | grep 8080`
- 检查防火墙设置

#### 4. 认证失败
- 确认测试账户存在于数据库中
- 检查用户名和密码是否正确
- 验证JWT配置是否正确

#### 5. 数据库连接问题
- 检查MySQL服务是否运行
- 验证数据库配置文件
- 确认数据库表结构正确

### 调试模式
在脚本中添加调试信息：
```bash
# 在脚本开头添加
set -x  # 显示执行的命令
set -e  # 遇到错误立即退出
```

## 📋 测试最佳实践

### 编写新测试
1. **命名规范**: `test_<功能名>_<接口名>.sh`
2. **脚本结构**: 包含测试说明、环境检查、测试用例、结果验证
3. **错误处理**: 合理的错误提示和故障恢复
4. **输出格式**: 统一的颜色标识和格式

### 测试用例设计
1. **正常流程**: 测试接口的正常功能
2. **边界条件**: 测试参数边界值
3. **错误处理**: 测试各种错误情况
4. **安全验证**: 测试认证和权限控制

### 维护原则
1. **定期运行**: 确保测试脚本的有效性
2. **及时更新**: API变更后同步更新测试
3. **文档同步**: 测试脚本与API文档保持一致

## 📚 参考资料

- [API接口文档](../docs/README.md)
- [项目部署指南](../docs/SECURITY_DEPLOYMENT.md)
- [数据库结构说明](../sql/)

---

*最后更新时间: 2024-08-28*