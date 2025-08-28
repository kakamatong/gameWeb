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
POST /api/admin/mails/send
Content-Type: application/json

{
  "type": 0,
  "title": "邮件标题",
  "content": "邮件内容",
  "awards": "{\"props\":[{\"id\":1,\"cnt\":100}]}",
  "startTime": "2024-01-01T00:00:00Z",
  "endTime": "2024-01-07T23:59:59Z"
}
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