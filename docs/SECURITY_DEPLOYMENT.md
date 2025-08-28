# 🔒 安全部署指南

## 快速安全部署

### 1. 环境准备
```bash
# 克隆仓库（在清理git历史后）
git clone <repository-url>
cd gameWeb

# 复制配置模板
cp config.example.yaml config.yaml
cp set_env.example.sh set_env.sh
```

### 2. 配置敏感信息
编辑 `config.yaml` 或设置环境变量：

#### 方式A: 使用配置文件
```bash
# 编辑配置文件（推荐）
nano config.yaml
```

#### 方式B: 使用环境变量
```bash
# 编辑环境变量脚本
nano set_env.sh

# 应用环境变量
source set_env.sh
```

### 3. 必须修改的敏感配置

#### 🚨 立即修改（高优先级）
- `MYSQL_GAMELOG_PASSWORD`: 设置新的数据库密码
- `MYSQL_GAMELOG_HOST`: 设置实际的数据库主机
- `MYSQL_GAMELOG_USER`: 设置实际的数据库用户名

#### 🔑 建议修改（中优先级）
- `JWT_SECRET`: 设置强JWT密钥（至少32字符）
- `ADMIN_JWT_SECRET`: 设置管理员JWT密钥
- `MYSQL_PASSWORD`: 设置主数据库密码
- `MYSQL_GAMEWEB_PASSWORD`: 设置gameWeb数据库密码

### 4. 安全检查清单

#### 数据库安全
- [ ] 已更改所有数据库密码
- [ ] 数据库用户权限最小化
- [ ] 启用数据库访问日志
- [ ] 配置防火墙规则

#### 应用安全
- [ ] JWT密钥足够强（建议使用随机生成）
- [ ] 配置文件权限正确（600或更严格）
- [ ] 日志目录权限正确
- [ ] 禁用调试模式（生产环境）

#### 网络安全
- [ ] 使用HTTPS（生产环境）
- [ ] 配置适当的CORS策略
- [ ] 限制API访问频率
- [ ] 监控异常访问

### 5. 密钥生成命令

```bash
# 生成强JWT密钥
openssl rand -base64 32

# 生成随机密码
openssl rand -base64 16
```

### 6. 生产部署

```bash
# 构建应用
./build.sh

# 启动服务（使用环境变量）
source set_env.sh
./run.sh

# 或者直接使用systemd服务
sudo systemctl start gameWeb
```

### 7. 监控和维护

#### 日志监控
```bash
# 监控应用日志
tail -f logs/game.log

# 监控错误日志
grep -i error logs/game.log
```

#### 安全审计
```bash
# 检查配置文件权限
ls -la config.yaml

# 检查进程状态
ps aux | grep gameWeb

# 检查端口占用
netstat -tulpn | grep 8080
```

### 8. 故障排查

#### 常见问题
1. **数据库连接失败**
   - 检查数据库服务状态
   - 验证连接参数
   - 检查网络连通性

2. **JWT验证失败**
   - 确认密钥配置正确
   - 检查令牌过期时间
   - 验证令牌格式

3. **权限错误**
   - 检查文件权限
   - 验证用户权限
   - 确认目录结构

### 9. 应急联系

如果发现安全问题，请立即：
1. 停止服务
2. 更改所有相关密码
3. 检查访问日志
4. 联系系统管理员

---

**⚠️ 重要提醒**:
- 永远不要在代码中硬编码敏感信息
- 定期轮换密码和密钥
- 保持系统和依赖项更新
- 定期进行安全审计