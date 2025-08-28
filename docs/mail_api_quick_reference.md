# 邮件API快速参考

## 客户端API

### 获取邮件列表
```bash
GET /api/mail/list?userid={userid}
```

### 获取邮件详情
```bash
GET /api/mail/detail/{id}?userid={userid}
```

### 标记已读
```bash
POST /api/mail/read/{id}?userid={userid}
```

### 领取奖励
```bash
POST /api/mail/claim/{id}?userid={userid}
```

## 管理后台API

### 发送系统邮件
```bash
# 全服邮件
POST /api/admin/mails/send
Content-Type: application/json

{
  "type": 0,
  "title": "系统维护补偿",
  "content": "感谢您的耐心等待，这是系统维护的补偿奖励。",
  "awards": "{\"props\":[{\"id\":1,\"cnt\":1000},{\"id\":2,\"cnt\":50}]}",
  "startTime": "2024-08-28T00:00:00Z",
  "endTime": "2024-09-04T23:59:59Z"
}

# 个人邮件
POST /api/admin/mails/send
Content-Type: application/json

{
  "type": 1,
  "title": "特殊奖励",
  "content": "恭喜您获得特殊成就！",
  "awards": "{\"props\":[{\"id\":3,\"cnt\":500}]}",
  "startTime": "2024-08-28T10:00:00Z",
  "endTime": "2024-09-28T23:59:59Z",
  "targetUsers": [12345, 67890, 11111]
}
```

### 获取用户邮件列表（仅生效）
```bash
# 基本查询
GET /api/admin/mails/?page=1&pageSize=10

# 按标题搜索
GET /api/admin/mails/?title=系统维护

# 按用户ID筛选
GET /api/admin/mails/?userid=12345
```

### 获取邮件详情（管理员视图）
```bash
GET /api/admin/mails/{id}
```

### 更新邮件状态
```bash
PUT /api/admin/mails/{id}/status
Content-Type: application/json

{
  "action": "extend",
  "endTime": "2024-02-01T23:59:59Z"
}
```

### 获取邮件统计
```bash
GET /api/admin/mails/stats
```

## 奖励格式
```json
{
  "props": [
    {"id": 1, "cnt": 100},
    {"id": 2, "cnt": 50}
  ]
}
```

## 邮件状态
- 0: 未读
- 1: 已读
- 2: 已领取
- 3: 已删除

详细文档请参考 [mail_api_documentation.md](./mail_api_documentation.md)