# 游戏管理后台API文档

## 概述

本文档描述了游戏管理后台的所有API接口，包括管理员认证、用户管理、日志查询和系统邮件功能。

## 基础信息

- **基础URL**: `http://your-domain:8080/api/admin`
- **认证方式**: JWT Bearer Token（管理员专用）
- **数据格式**: JSON
- **字符编码**: UTF-8

## 数据库配置

项目使用多数据库架构：

1. **game库** - 用户游戏数据（localhost）
   - userData: 用户基础信息
   - userRiches: 用户财富信息
   - userStatus: 用户状态信息

2. **gameWeb库** - 管理员数据（localhost）
   - adminAccount: 管理员账户信息

3. **username库** - 日志数据（xxx.xxx.xxx.xxx）
   - logAuth: 用户登录认证日志
   - logResult10001: 用户对局结果日志

4. **邮件系统表**（game库）
   - mails: 邮件内容表
   - mailSystem: 邮件系统表
   - mailUsers: 用户邮件表

## 认证机制

管理后台使用独立的JWT认证系统，与客户端认证完全分离：

### 客户端认证（现有）
- 使用DES+Redis复杂验证机制
- JWT密钥: `config.JWT.SecretKey`
- 适用于游戏客户端API

### 管理后台认证（新增）
- 使用简化的JWT认证机制
- JWT密钥: `config.Admin.JWTSecretKey`
- 包含Redis会话验证
- 支持IP地址变更检测

## API接口

### 邮件系统架构

项目包含两套独立的邮件API系统：

1. **游戏客户端邮件API** - 供玩家使用
   - 基础路径: `/api/mail/`
   - 认证方式: 客户端JWT认证 (`AuthMiddlewareByJWT`)
   - 功能: 查看邮件、标记已读、领取奖励

2. **管理后台邮件API** - 供管理员使用
   - 基础路径: `/api/admin/mails/`
   - 认证方式: 管理员JWT认证 (`AdminJWTMiddleware`)
   - 功能: 发送邮件、邮件管理、统计信息

### 1. 管理员认证

#### 1.1 管理员登录
```http
POST /api/admin/login
Content-Type: application/json

{
    "username": "admin",
    "password": "password123"
}
```

**响应示例**:
```json
{
    "code": 200,
    "message": "登录成功",
    "data": {
        "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "adminInfo": {
            "id": 1,
            "username": "admin",
            "email": "admin@example.com",
            "realName": "系统管理员",
            "isSuperAdmin": true,
            "lastLoginTime": "2024-01-01T12:00:00Z"
        }
    }
}
```

#### 1.2 管理员登出
```http
POST /api/admin/logout
Authorization: Bearer <token>
```

#### 1.3 获取管理员信息
```http
GET /api/admin/info
Authorization: Bearer <token>
```

#### 1.4 创建管理员（仅超级管理员）
```http
POST /api/admin/create-admin
Authorization: Bearer <token>
Content-Type: application/json

{
    "username": "newadmin",
    "password": "password123",
    "email": "newadmin@example.com",
    "realName": "新管理员",
    "isSuperAdmin": 0
}
```

### 2. 用户管理

#### 2.1 获取用户列表
```http
GET /api/admin/users?page=1&pageSize=20&keyword=玩家昵称&userid=12345
Authorization: Bearer <token>
```

**查询参数**:
- `page`: 页码（默认1）
- `pageSize`: 每页大小（默认20，最大100）
- `keyword`: 搜索关键词（昵称）
- `userid`: 特定用户ID

**响应示例**:
```json
{
    "code": 200,
    "message": "获取成功",
    "data": {
        "total": 100,
        "users": [
            {
                "userid": 12345,
                "nickname": "玩家昵称",
                "headurl": "头像URL",
                "sex": 1,
                "province": "广东省",
                "city": "深圳市",
                "ip": "192.168.1.1",
                "status": 1,
                "gameid": 0,
                "roomid": 0,
                "riches": [
                    {"richType": 1, "richNums": 10000},
                    {"richType": 2, "richNums": 500}
                ],
                "createTime": "2024-01-01T12:00:00Z",
                "updateTime": "2024-01-01T12:00:00Z"
            }
        ]
    }
}
```

#### 2.2 获取用户详情
```http
GET /api/admin/users/{userid}
Authorization: Bearer <token>
```

#### 2.3 更新用户信息
```http
PUT /api/admin/users/{userid}
Authorization: Bearer <token>
Content-Type: application/json

{
    "nickname": "新昵称",
    "status": 1,
    "riches": [
        {"richType": 1, "richNums": 15000},
        {"richType": 2, "richNums": 600}
    ]
}
```

### 3. 用户日志查询

#### 3.1 获取用户登录日志
```http
GET /api/admin/logs/auth?userid=12345&startTime=2024-01-01T00:00:00Z&endTime=2024-01-31T23:59:59Z&page=1&pageSize=20
Authorization: Bearer <token>
```

**查询参数**:
- `userid`: 用户ID（可选）
- `startTime`: 开始时间
- `endTime`: 结束时间
- `page`: 页码
- `pageSize`: 每页大小

**响应示例**:
```json
{
    "code": 200,
    "message": "获取成功",
    "data": {
        "total": 50,
        "page": 1,
        "pageSize": 20,
        "data": [
            {
                "id": 1,
                "userid": 12345,
                "channel": "android",
                "ip": "192.168.1.1",
                "deviceId": "device123",
                "loginTime": "2024-01-01T08:00:00Z",
                "logoutTime": "2024-01-01T10:00:00Z",
                "duration": 7200,
                "status": 1
            }
        ]
    }
}
```

#### 3.2 获取用户对局日志
```http
GET /api/admin/logs/game?userid=12345&startTime=2024-01-01T00:00:00Z&endTime=2024-01-31T23:59:59Z&page=1&pageSize=20
Authorization: Bearer <token>
```

#### 3.3 获取用户登录统计
```http
GET /api/admin/logs/login-stats?userid=12345
Authorization: Bearer <token>
```

**响应示例**:
```json
{
    "code": 200,
    "message": "获取成功",
    "data": {
        "totalLogins": 150,
        "lastLoginTime": "2024-01-01T08:00:00Z",
        "todayLogins": 3,
        "weekLogins": 15,
        "avgDuration": 120.5
    }
}
```

#### 3.4 获取用户对局统计
```http
GET /api/admin/logs/game-stats?userid=12345
Authorization: Bearer <token>
```

### 4. 系统邮件管理

#### 4.1 发送系统邮件
```http
POST /api/admin/mails/send
Authorization: Bearer <token>
Content-Type: application/json

{
    "type": 0,
    "title": "邮件标题",
    "content": "邮件内容",
    "awards": "[{\"type\":1,\"count\":100}]",
    "startTime": "2024-01-01T00:00:00Z",
    "endTime": "2024-12-31T23:59:59Z",
    "targetUsers": []
}
```

**参数说明**:
- `type`: 邮件类型（0-全服邮件，1-个人邮件）
- `title`: 邮件标题
- `content`: 邮件内容
- `awards`: 奖励内容（JSON格式）
- `startTime`: 邮件有效开始时间
- `endTime`: 邮件有效结束时间
- `targetUsers`: 目标用户ID列表（个人邮件时使用）

#### 4.2 获取邮件列表
```http
GET /api/admin/mails?page=1&pageSize=20&status=1&type=0
Authorization: Bearer <token>
```

#### 4.3 获取邮件详情
```http
GET /api/admin/mails/{id}
Authorization: Bearer <token>
```

#### 4.4 更新邮件状态
```http
PUT /api/admin/mails/{id}/status
Authorization: Bearer <token>
Content-Type: application/json

{
    "status": 0
}
```

#### 4.5 获取邮件统计
```http
GET /api/admin/mails/stats
Authorization: Bearer <token>
```

## 5. 游戏客户端邮件API

### 5.1 获取邮件列表
```http
POST /api/mail/list
Authorization: Bearer <token>
Content-Type: application/json

{
    "userid": 12345
}
```

### 5.2 获取邮件详情
```http
POST /api/mail/detail/{id}
Authorization: Bearer <token>
Content-Type: application/json

{
    "userid": 12345
}
```

### 5.3 标记邮件为已读
```http
POST /api/mail/read/{id}
Authorization: Bearer <token>
Content-Type: application/json

{
    "userid": 12345
}
```

**响应示例**:
```json
{
    "code": 200,
    "message": "success",
    "data": {}
}
```

### 5.4 领取邮件奖励
```http
POST /api/mail/getaward/{id}
Authorization: Bearer <token>
Content-Type: application/json

{
    "userid": 12345
}
```

**响应示例**:
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "awards": [
            {"type": 1, "count": 100},
            {"type": 2, "count": 50}
        ]
    }
}
```

**特殊情况**:
- 已领取: `{"code": 200, "message": "Award already received", "data": {"alreadyReceived": true}}`
- 无奖励: `{"code": 200, "message": "No awards in this mail", "data": {}}`
- 操作过频: `{"code": 429, "message": "Operation too frequent"}`

**安全机制**:
- Redis分布式锁防止重复领取
- 数据库事务保证一致性
- JWT用户ID验证防止越权

## 错误码说明

| 错误码 | 说明 |
|-------|------|
| 200   | 成功 |
| 400   | 参数错误 |
| 401   | 未认证或认证失败 |
| 403   | 权限不足 |
| 404   | 资源不存在 |
| 500   | 服务器内部错误 |

## 数据模型

### 管理员账户模型
```go
type AdminAccount struct {
    ID              uint64    `json:"id"`
    Username        string    `json:"username"`
    Email           string    `json:"email"`
    RealName        string    `json:"realName"`
    IsSuperAdmin    int8      `json:"isSuperAdmin"`
    LastLoginTime   time.Time `json:"lastLoginTime"`
    // ... 其他字段
}
```

### 用户信息模型
```go
type UserInfo struct {
    UserID     int64       `json:"userid"`
    Nickname   string      `json:"nickname"`
    HeadURL    string      `json:"headurl"`
    Sex        int8        `json:"sex"`
    Province   string      `json:"province"`
    City       string      `json:"city"`
    IP         string      `json:"ip"`
    Status     int8        `json:"status"`
    GameID     int64       `json:"gameid"`
    RoomID     int64       `json:"roomid"`
    Riches     []UserRich  `json:"riches"`
    CreateTime time.Time   `json:"createTime"`
    UpdateTime time.Time   `json:"updateTime"`
}
```

### 用户财富模型
```go
type UserRich struct {
    RichType int8  `json:"richType"`
    RichNums int64 `json:"richNums"`
}
```

## 部署说明

1. **数据库准备**:
   - 在gameWeb库中执行`sql/adminAccount.sql`
   - 在game库中执行用户相关的SQL文件
   - 确保username库（xxx.xxx.xxx.xxx）的连接权限

2. **配置文件**:
   - 复制`config/config.yaml`到项目根目录
   - 根据实际环境修改数据库连接信息

3. **启动服务**:
   ```bash
   go build -o gameWeb main.go
   ./gameWeb
   ```

4. **创建初始管理员**:
   - 直接在数据库中插入管理员记录
   - 密码使用bcrypt加密

## 安全考虑

1. **JWT密钥管理**: 生产环境中务必更换默认的JWT密钥
2. **数据库权限**: 使用最小权限原则配置数据库账户
3. **HTTPS**: 生产环境中启用HTTPS传输
4. **IP白名单**: 可考虑为管理后台添加IP访问限制
5. **操作日志**: 所有管理操作都会记录日志

## 测试建议

1. **单元测试**: 为控制器和数据库操作函数编写单元测试
2. **集成测试**: 测试完整的API流程
3. **压力测试**: 验证多数据库连接的性能
4. **安全测试**: 验证JWT认证和权限控制