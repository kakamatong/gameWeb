# 邮件接口分离与数据库修复说明

## 修复历史

### 第一次修复：接口分离
原来的邮件接口中，客户端邮件接口和管理后台邮件接口混在一起，导致函数名冲突和逻辑混乱。现在已经将两部分逻辑完全分离，避免相互干扰。

### 第二次修复：数据库连接
发现客户端邮件接口使用了错误的数据库连接（db.MySQLDB指向game数据库），但邮件表在gameWeb数据库中。已修复为使用正确的数据库连接（db.MySQLDBGameWeb）。

### 第三次修复：数据库字段匹配
发现代码中使用的字段名与实际数据库表结构不匹配，如：
- `mailSystem.isGlobal` → 实际使用 `mailSystem.type`
- `mails.startTime/endTime/status` → 实际只有 `mails.created_at`
- `mailUsers.isRead/isReceived` → 实际使用 `mailUsers.status`

已根据实际数据库表结构修复所有SQL查询。

## 接口分离方案

### 客户端邮件接口 (游戏客户端使用)
路由前缀: `/api/mail/`
认证方式: JWT客户端认证 (`AuthMiddlewareByJWT`)

| 接口路径 | HTTP方法 | 控制器函数 | 功能描述 |
|---------|---------|-----------|---------|
| `/api/mail/list` | POST | `GetClientMailList` | 获取用户邮件列表 |
| `/api/mail/detail/:id` | POST | `GetClientMailDetail` | 获取邮件详情 |
| `/api/mail/read/:id` | POST | `MarkMailAsRead` | 标记邮件为已读 |
| `/api/mail/getaward/:id` | POST | `GetMailAward` | 领取邮件奖励 |

#### 客户端接口特点：
- 需要在请求体中传递 `userid` 参数
- 使用JWT验证用户身份，确保用户只能操作自己的邮件
- 自动同步全服邮件到用户邮件表
- 返回格式统一为 `{"code": 200, "message": "success", "data": {...}}`
- 只显示未过期且有效的邮件

### 管理后台邮件接口 (管理员使用)
路由前缀: `/api/admin/mails/`
认证方式: 管理员JWT认证 (`AdminJWTMiddleware`)

| 接口路径 | HTTP方法 | 控制器函数 | 功能描述 |
|---------|---------|-----------|---------|
| `/api/admin/mails/send` | POST | `SendSystemMail` | 发送系统邮件 |
| `/api/admin/mails/` | GET | `GetAdminMailList` | 获取邮件列表(管理员视图) |
| `/api/admin/mails/:id` | GET | `GetAdminMailDetail` | 获取邮件详情(管理员视图) |
| `/api/admin/mails/:id/status` | PUT | `UpdateMailStatus` | 更新邮件状态 |
| `/api/admin/mails/stats` | GET | `GetMailStats` | 获取邮件统计信息 |

#### 管理后台接口特点：
- 使用管理员JWT验证权限
- 支持分页查询和条件筛选
- 可以查看所有邮件（包括过期的）
- 支持邮件状态管理
- 返回格式为标准的API响应格式

## 数据库操作分离

### 客户端专用函数：
- `syncSystemMails(userID)` - 同步系统邮件到用户邮件表
- `getClientMailList(userID)` - 获取用户邮件列表
- `getClientMailDetail(mailID, userID)` - 获取用户邮件详情

### 管理后台专用函数：
- `createMail()` - 创建邮件记录
- `createMailSystem()` - 创建邮件系统记录
- `createMailUser()` - 创建用户邮件记录
- `getMailCount()` - 获取邮件总数（支持条件查询）
- `getMailList()` - 获取邮件列表（管理员视图，支持分页）
- `getMailByID()` - 根据ID获取邮件（管理员视图）
- `getMailSystemByMailID()` - 获取邮件系统信息
- `updateMailStatus()` - 更新邮件状态
- `getMailStats()` - 获取邮件统计信息

## 主要改进点

1. **函数名区分**: 
   - 客户端: `GetClientMailList`, `GetClientMailDetail`
   - 管理后台: `GetAdminMailList`, `GetAdminMailDetail`

2. **认证分离**: 
   - 客户端使用游戏用户JWT
   - 管理后台使用管理员JWT

3. **权限控制**: 
   - 客户端严格验证用户身份，只能操作自己的邮件
   - 管理后台具有全局邮件管理权限

4. **数据格式统一**: 
   - 客户端接口统一返回 `{"code", "message", "data"}` 格式
   - 管理后台保持原有的API响应格式

5. **业务逻辑分离**: 
   - 客户端自动处理系统邮件同步
   - 管理后台专注于邮件管理和统计

## 迁移影响

- **客户端**: 无需修改调用方式，接口路径和参数保持不变
- **管理后台**: 无需修改调用方式，接口路径和参数保持不变
- **向后兼容**: 完全兼容原有的调用方式

## 测试建议

1. 测试客户端邮件列表和详情获取
2. 测试管理后台邮件管理功能
3. 验证权限隔离是否生效
4. 确认邮件奖励领取功能正常