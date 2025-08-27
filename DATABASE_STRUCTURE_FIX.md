# 数据库表结构不匹配问题分析

## 问题概述

通过分析错误日志和代码，发现了数据库表结构与代码模型定义不匹配的问题：

### 错误日志
```
Error 1054 (42S22): Unknown column 'ms.isGlobal' in 'where clause'
Error 1054 (42S22): Unknown column 'm.startTime' in 'field list'
```

## 表结构对比

### 1. mails 表

**代码模型定义** (models/models.go):
```go
type Mails struct {
    ID        int64     `json:"id" db:"id"`
    Type      int8      `json:"type" db:"type"`
    Title     string    `json:"title" db:"title"`
    Content   string    `json:"content" db:"content"`
    Awards    string    `json:"awards" db:"awards"`
    StartTime time.Time `json:"startTime" db:"startTime"`  // ❌ 实际表中不存在
    EndTime   time.Time `json:"endTime" db:"endTime"`      // ❌ 实际表中不存在
    Status    int8      `json:"status" db:"status"`        // ❌ 实际表中不存在
}
```

**实际数据库表** (sql/mails.sql):
```sql
CREATE TABLE mails (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    type INT NOT NULL,
    senderid BIGINT NOT NULL DEFAULT 0,           -- ✅ 代码中缺少
    title VARCHAR(100) NOT NULL,
    content TEXT,
    awards VARCHAR(512),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP  -- ✅ 代码中缺少
);
```

### 2. mailSystem 表

**代码模型定义** (models/models.go):
```go
type MailSystem struct {
    ID       int64 `json:"id" db:"id"`
    MailID   int64 `json:"mailId" db:"mailid"`
    UserID   int64 `json:"userid" db:"userid"`
    IsGlobal int8  `json:"isGlobal" db:"isGlobal"`  // ❌ 实际表中不存在
}
```

**实际数据库表** (sql/mailSystem.sql):
```sql
CREATE TABLE mailSystem (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    type INT NOT NULL,                    -- ✅ 代码中缺少
    mailid BIGINT NOT NULL,
    startTime DATETIME NOT NULL,         -- ✅ 代码中缺少
    endTime DATETIME NOT NULL,           -- ✅ 代码中缺少
    -- 注意：没有 userid 和 isGlobal 字段
);
```

### 3. mailUsers 表

**代码模型定义** (models/models.go):
```go
type MailUsers struct {
    ID          int64     `json:"id" db:"id"`
    MailID      int64     `json:"mailId" db:"mailid"`
    UserID      int64     `json:"userid" db:"userid"`
    IsRead      int8      `json:"isRead" db:"isRead"`      // ❌ 实际表中不存在
    IsReceived  int8      `json:"isReceived" db:"isReceived"` // ❌ 实际表中不存在
    ReadTime    time.Time `json:"readTime" db:"readTime"`     // ❌ 实际表中不存在
    ReceiveTime time.Time `json:"receiveTime" db:"receiveTime"` // ❌ 实际表中不存在
    CreateTime  time.Time `json:"createTime" db:"create_time"`  // ❌ 实际表中不存在
}
```

**实际数据库表** (sql/mailUsers.sql):
```sql
CREATE TABLE mailUsers (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    userid BIGINT NOT NULL,
    mailid BIGINT NOT NULL,
    status TINYINT NOT NULL DEFAULT 0,    -- ✅ 代码中缺少，0-未读,1-已读,2-已领取,3-已删除
    startTime DATETIME NOT NULL,          -- ✅ 代码中缺少
    endTime DATETIME NOT NULL,            -- ✅ 代码中缺少
    update_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

## 修复方案

基于实际的数据库表结构，我已经修改了客户端邮件接口的SQL查询：

### 已修复的函数
1. **syncSystemMails()** - 使用 `ms.type = 0` 替代 `ms.isGlobal = 1`
2. **getClientMailList()** - 根据实际表结构重写查询
3. **getClientMailDetail()** - 根据实际表结构重写查询
4. **MarkMailAsRead()** - 使用 `status` 字段替代 `isRead`
5. **GetMailAward()** - 使用 `status` 字段判断领取状态

### 待修复的管理后台函数
以下函数仍需要根据实际表结构进行调整：
1. **getMailList()** - 查询不存在的 `startTime`、`endTime`、`status` 字段
2. **getMailByID()** - 查询不存在的 `startTime`、`endTime`、`status` 字段
3. **getMailSystemByMailID()** - 查询不存在的 `userid`、`isGlobal` 字段
4. **updateMailStatus()** - 更新不存在的 `status` 字段
5. **getMailStats()** - 查询不存在的 `status` 字段

### 字段映射关系

| 功能 | 原始设计 | 实际实现 |
|------|----------|----------|
| 邮件类型判断 | `mailSystem.isGlobal` | `mailSystem.type` (0=全服, 1=个人) |
| 邮件时间控制 | `mails.startTime/endTime` | `mailSystem.startTime/endTime` |
| 邮件状态 | `mails.status` | 不存在，可能需要业务逻辑判断 |
| 用户邮件状态 | `mailUsers.isRead/isReceived` | `mailUsers.status` (0=未读,1=已读,2=已领取) |

## 建议的完整修复步骤

1. **立即修复** - 修改管理后台邮件函数的SQL查询以匹配实际表结构
2. **长期规划** - 考虑是否需要更新数据库表结构或模型定义以保持一致性
3. **文档更新** - 更新API文档和数据库设计文档

## 注意事项

- 当前修复保持了API接口的向后兼容性
- `status` 字段的值映射：0=未读，1=已读，2=已领取，3=已删除
- 邮件的有效性通过 `mailSystem` 表的 `startTime` 和 `endTime` 控制
- 全服邮件通过 `mailSystem.type = 0` 标识