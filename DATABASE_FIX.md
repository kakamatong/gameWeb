# 数据库连接问题修复说明

## 问题描述

从日志错误可以看出：
```
Error 1146 (42S02): Table 'game.mailSystem' doesn't exist
Error 1146 (42S02): Table 'game.mails' doesn't exist
```

客户端邮件接口试图在 `game` 数据库中查找邮件相关的表，但这些表实际上在 `gameWeb` 数据库中。

## 根本原因

项目采用了多数据库架构：
- **MySQLDB** - 连接 `game` 数据库（用户游戏数据）
- **MySQLDBGameWeb** - 连接 `gameWeb` 数据库（管理员数据和邮件系统）
- **MySQLDBGameLog** - 连接 `username` 数据库（日志数据）

但邮件相关的函数错误地使用了 `db.MySQLDB`（game数据库）而不是 `db.MySQLDBGameWeb`（gameWeb数据库）。

## 修复内容

### 1. 客户端邮件接口函数
以下函数已修复数据库连接，从 `db.MySQLDB` 改为 `db.MySQLDBGameWeb`：

- `syncSystemMails()` - 同步系统邮件
- `getClientMailList()` - 获取客户端邮件列表
- `getClientMailDetail()` - 获取客户端邮件详情
- `MarkMailAsRead()` - 标记邮件已读

### 2. 管理后台邮件接口函数
以下函数已修复数据库连接，从 `db.MySQLDB` 改为 `db.MySQLDBGameWeb`：

- `SendSystemMail()` - 发送系统邮件
- `getMailCount()` - 获取邮件总数
- `getMailList()` - 获取邮件列表
- `getMailByID()` - 根据ID获取邮件
- `getMailSystemByMailID()` - 获取邮件系统记录
- `checkMailExists()` - 检查邮件是否存在
- `updateMailStatus()` - 更新邮件状态
- `getMailStats()` - 获取邮件统计

### 3. 跨数据库事务处理

**GetMailAward() 函数特殊处理**：
此函数需要操作两个数据库：
- 查询邮件信息 → `gameWeb` 数据库
- 更新用户财富 → `game` 数据库
- 更新邮件状态 → `gameWeb` 数据库

由于无法使用跨数据库事务，采用分步处理：
1. 查询邮件信息并验证（gameWeb数据库）
2. 发放奖励到用户财富（game数据库事务）
3. 更新邮件状态（gameWeb数据库事务）

**注意**：这种处理方式确保了用户财富的发放，即使邮件状态更新失败也会记录日志便于后续处理。

## 数据库表分布

### gameWeb 数据库
- `mails` - 邮件主表
- `mailSystem` - 邮件系统表
- `mailUsers` - 用户邮件状态表
- `adminAccount` - 管理员账户表

### game 数据库  
- `userRiches` - 用户财富表
- `userData` - 用户数据表
- `userStatus` - 用户状态表

### username 数据库
- `logAuth` - 认证日志表
- `logResult10001` - 游戏结果日志表

## 验证修复

修复后，客户端邮件接口应该能够正常工作：
- `/api/mail/list` - 获取邮件列表
- `/api/mail/detail/:id` - 获取邮件详情  
- `/api/mail/read/:id` - 标记邮件已读
- `/api/mail/getaward/:id` - 领取邮件奖励

## 注意事项

1. **数据一致性**：跨数据库操作无法保证ACID事务特性，需要应用层处理数据一致性
2. **错误处理**：GetMailAward函数已加强错误处理和日志记录
3. **监控**：建议监控邮件状态更新失败的情况，及时手动处理数据不一致问题