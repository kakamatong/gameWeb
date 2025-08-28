# 邮件模块API接入文档

## 概述

邮件模块为游戏提供完整的邮件系统功能，包括系统邮件发送、用户邮件获取、邮件阅读和奖励领取等功能。

## 数据库设计

### 三表结构设计

#### 1. mails表 - 邮件基本信息
存放每封邮件的基本内容，不包含生效时间。

```sql
CREATE TABLE mails (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '邮件唯一ID',
    type INT NOT NULL COMMENT '邮件类型: 0-全服邮件, 1-个人邮件',
    senderid BIGINT NOT NULL DEFAULT 0 COMMENT '发送者ID, 0表示系统',
    title VARCHAR(100) NOT NULL COMMENT '邮件标题',
    content TEXT COMMENT '邮件内容',
    awards VARCHAR(512) COMMENT '奖励内容，JSON格式',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间'
);
```

**awards字段JSON结构**:
```json
{
  "props": [
    {"id": 2, "cnt": 20000},
    {"id": 1, "cnt": 100}
  ]
}
```

#### 2. mailSystem表 - 系统邮件配置
包含每封系统邮件的生效时间和类型配置。

```sql
CREATE TABLE mailSystem (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    type INT NOT NULL COMMENT '邮件类型: 0-全服邮件, 1-个人邮件',
    mailid BIGINT NOT NULL COMMENT '邮件ID，关联mails表id',
    startTime DATETIME NOT NULL COMMENT '邮件生效开始时间',
    endTime DATETIME NOT NULL COMMENT '邮件生效结束时间'
);
```

#### 3. mailUsers表 - 用户邮件状态
表示用户目前收到的邮件状态。

```sql
CREATE TABLE mailUsers (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    userid BIGINT NOT NULL COMMENT '用户ID',
    mailid BIGINT NOT NULL COMMENT '邮件ID，关联mails表id',
    status TINYINT NOT NULL DEFAULT 0 COMMENT '状态: 0-未读, 1-已读, 2-已领取, 3-已删除',
    startTime DATETIME NOT NULL COMMENT '邮件生效开始时间',
    endTime DATETIME NOT NULL COMMENT '邮件生效结束时间',
    update_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

## 业务逻辑

### 邮件同步机制
1. 用户请求邮件列表时，系统自动同步当前生效的系统邮件
2. 对比用户已有邮件，将新邮件插入到用户邮件表
3. 只返回当前生效的邮件，过期邮件不显示

### 邮件状态流转
- **0-未读**: 邮件刚分发给用户，未查看
- **1-已读**: 用户查看过邮件内容
- **2-已领取**: 用户已领取邮件奖励
- **3-已删除**: 邮件被用户删除（不再显示）

## API接口

### 基础信息
- **客户端API基础URL**: `http://your-domain:8080/api/mail`
- **管理后台API基础URL**: `http://your-domain:8080/api/admin/mails`
- **认证方式**: JWT Bearer Token
- **数据格式**: JSON

## 客户端API

### 1. 获取邮件列表

#### 接口信息
- **URL**: `/api/mail/list`
- **方法**: GET
- **认证**: 需要JWT认证

#### 请求参数
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| userid | integer | 是 | 用户ID |

#### 请求示例
```bash
curl -X GET "http://localhost:8080/api/mail/list?userid=12345" \
  -H "Authorization: Bearer your-jwt-token"
```

#### 响应示例
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "total": 3,
    "mails": [
      {
        "id": 1001,
        "type": 0,
        "title": "系统维护补偿",
        "content": "感谢您的耐心等待，这是系统维护的补偿奖励。",
        "awards": "{\"props\":[{\"id\":1,\"cnt\":1000},{\"id\":2,\"cnt\":50}]}",
        "status": 0,
        "startTime": "2024-01-01T00:00:00Z",
        "endTime": "2024-01-07T23:59:59Z",
        "createdAt": "2024-01-01T10:00:00Z"
      },
      {
        "id": 1002,
        "type": 0,
        "title": "每日登录奖励",
        "content": "连续登录7天的奖励，请查收！",
        "awards": "{\"props\":[{\"id\":3,\"cnt\":500}]}",
        "status": 1,
        "startTime": "2024-01-02T00:00:00Z",
        "endTime": "2024-01-09T23:59:59Z",
        "createdAt": "2024-01-02T10:00:00Z"
      }
    ]
  }
}
```

### 2. 获取邮件详情

#### 接口信息
- **URL**: `/api/mail/detail/{id}`
- **方法**: GET
- **认证**: 需要JWT认证

#### 请求参数
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | integer | 是 | 邮件ID（路径参数）|
| userid | integer | 是 | 用户ID（查询参数）|

#### 请求示例
```bash
curl -X GET "http://localhost:8080/api/mail/detail/1001?userid=12345" \
  -H "Authorization: Bearer your-jwt-token"
```

#### 响应示例
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "id": 1001,
    "type": 0,
    "title": "系统维护补偿",
    "content": "感谢您的耐心等待，系统于2024年1月1日进行了重要更新，为了补偿给您带来的不便，特发放以下奖励。",
    "awards": "{\"props\":[{\"id\":1,\"cnt\":1000},{\"id\":2,\"cnt\":50}]}",
    "status": 0,
    "startTime": "2024-01-01T00:00:00Z",
    "endTime": "2024-01-07T23:59:59Z",
    "createdAt": "2024-01-01T10:00:00Z"
  }
}
```

### 3. 标记邮件已读

#### 接口信息
- **URL**: `/api/mail/read/{id}`
- **方法**: POST
- **认证**: 需要JWT认证

#### 请求参数
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | integer | 是 | 邮件ID（路径参数）|
| userid | integer | 是 | 用户ID（查询参数）|

#### 请求示例
```bash
curl -X POST "http://localhost:8080/api/mail/read/1001?userid=12345" \
  -H "Authorization: Bearer your-jwt-token"
```

#### 响应示例
```json
{
  "code": 200,
  "message": "标记成功"
}
```

### 4. 领取邮件奖励

#### 接口信息
- **URL**: `/api/mail/claim/{id}`
- **方法**: POST
- **认证**: 需要JWT认证

#### 请求参数
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | integer | 是 | 邮件ID（路径参数）|
| userid | integer | 是 | 用户ID（查询参数）|

#### 请求示例
```bash
curl -X POST "http://localhost:8080/api/mail/claim/1001?userid=12345" \
  -H "Authorization: Bearer your-jwt-token"
```

#### 响应示例
```json
{
  "code": 200,
  "message": "领取成功",
  "data": {
    "awards": "{\"props\":[{\"id\":1,\"cnt\":1000},{\"id\":2,\"cnt\":50}]}"
  }
}
```

## 管理后台API

### 1. 发送系统邮件

#### 接口信息
- **URL**: `/api/admin/mails/send`
- **方法**: POST
- **认证**: 需要管理员JWT认证

#### 请求参数
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| type | integer | 是 | 邮件类型：0-全服邮件，1-个人邮件 |
| title | string | 是 | 邮件标题（1-100字符）|
| content | string | 是 | 邮件内容（1-1000字符）|
| awards | string | 否 | 奖励JSON字符串 |
| startTime | string | 是 | 生效开始时间（ISO 8601格式）|
| endTime | string | 是 | 生效结束时间（ISO 8601格式）|
| targetUsers | array | 否 | 目标用户ID数组（个人邮件时使用）|

#### 请求示例
```bash
curl -X POST "http://localhost:8080/api/admin/mails/send" \
  -H "Authorization: Bearer admin-jwt-token" \
  -H "Content-Type: application/json" \
  -d '{
    "type": 0,
    "title": "新年活动奖励",
    "content": "祝您新年快乐！特发放新年活动奖励。",
    "awards": "{\"props\":[{\"id\":1,\"cnt\":2000},{\"id\":2,\"cnt\":100}]}",
    "startTime": "2024-02-01T00:00:00Z",
    "endTime": "2024-02-07T23:59:59Z"
  }'
```

#### 响应示例
```json
{
  "code": 200,
  "message": "发送成功",
  "data": {
    "mailId": 1003
  }
}
```

## 错误响应

### 通用错误码
| 错误码 | 说明 |
|--------|------|
| 400 | 参数错误 |
| 401 | 未认证 |
| 403 | 权限不足 |
| 404 | 资源不存在 |
| 500 | 系统错误 |

### 错误响应示例
```json
{
  "code": 400,
  "message": "缺少用户ID参数"
}
```

```json
{
  "code": 404,
  "message": "邮件不存在或已过期"
}
```

```json
{
  "code": 500,
  "message": "奖励已经领取过了"
}
```

## 管理后台API

### 1. 获取用户邮件列表（仅生效邮件）

#### 接口信息
- **URL**: `/api/admin/mails/`
- **方法**: GET
- **认证**: 需要管理员JWT认证
- **功能**: 查询所有用户的邮件记录，仅返回当前生效的邮件

#### 请求参数
| 参数名 | 类型 | 必填 | 默认值 | 说明 |
|--------|------|------|--------|------|
| page | integer | 否 | 1 | 页码，从1开始 |
| pageSize | integer | 否 | 10 | 每页数量，1-100 |
| title | string | 否 | - | 邮件标题模糊搜索 |
| userid | integer | 否 | - | 用户ID筛选 |

#### 请求示例
```bash
# 获取第1页，每页10条
curl -X GET "http://localhost:8080/api/admin/mails/?page=1&pageSize=10" \
  -H "Authorization: Bearer admin-jwt-token"

# 根据标题搜索
curl -X GET "http://localhost:8080/api/admin/mails/?title=系统维护" \
  -H "Authorization: Bearer admin-jwt-token"

# 根据用户ID筛选
curl -X GET "http://localhost:8080/api/admin/mails/?userid=12345" \
  -H "Authorization: Bearer admin-jwt-token"
```

#### 响应示例
```json
{
  "code": 200,
  "message": "查询成功",
  "data": {
    "list": [
      {
        "id": 1001,
        "type": 0,
        "title": "系统维护补偿",
        "content": "感谢您的耐心等待，这是系统维护的补偿奖励。",
        "awards": "{\"props\":[{\"id\":1,\"cnt\":1000},{\"id\":2,\"cnt\":50}]}",
        "createdAt": "2024-01-01T10:00:00Z",
        "userid": 12345,
        "status": 1,
        "startTime": "2024-01-01T00:00:00Z",
        "endTime": "2024-01-07T23:59:59Z",
        "updateAt": "2024-01-02T08:30:00Z"
      },
      {
        "id": 1002,
        "type": 0,
        "title": "每日登录奖励",
        "content": "连续登录7天的奖励，请查收！",
        "awards": "{\"props\":[{\"id\":3,\"cnt\":500}]}",
        "createdAt": "2024-01-02T10:00:00Z",
        "userid": 12345,
        "status": 2,
        "startTime": "2024-01-02T00:00:00Z",
        "endTime": "2024-01-09T23:59:59Z",
        "updateAt": "2024-01-03T09:15:00Z"
      }
    ],
    "total": 25,
    "page": 1,
    "pageSize": 10,
    "summary": {
      "description": "当前生效的用户邮件",
      "filterTime": "2024-08-28 15:30:00"
    }
  }
}
```

#### 响应字段说明

**用户邮件记录字段**:
- `id`: 邮件ID
- `type`: 邮件类型 (0-全服邮件, 1-个人邮件)
- `title`: 邮件标题
- `content`: 邮件内容
- `awards`: 奖励内容（JSON格式）
- `createdAt`: 邮件创建时间
- `userid`: 用户ID
- `status`: 邮件状态 (0-未读, 1-已读, 2-已领取)
- `startTime`: 邮件生效开始时间
- `endTime`: 邮件生效结束时间
- `updateAt`: 最后更新时间

**统计信息字段**:
- `total`: 符合条件的总记录数
- `page`: 当前页码
- `pageSize`: 每页数量
- `summary.description`: 查询结果描述
- `summary.filterTime`: 查询时间点

#### 错误响应
```json
// 参数错误
{
  "code": 400,
  "message": "参数错误"
}

// 系统错误
{
  "code": 500,
  "message": "系统错误"
}
```

### 2. 获取邮件详情（管理员视图）

#### 接口信息
- **URL**: `/api/admin/mails/{id}`
- **方法**: GET
- **认证**: 需要管理员JWT认证

### 3. 更新邮件状态

#### 接口信息
- **URL**: `/api/admin/mails/{id}/status`
- **方法**: PUT
- **认证**: 需要管理员JWT认证

### 5. 发送系统邮件

#### 接口信息
- **URL**: `/api/admin/mails/send`
- **方法**: POST
- **认证**: 需要管理员JWT认证
- **功能**: 发送全服邮件或个人邮件

#### 请求参数
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| type | integer | 是 | 邮件类型: 0-全服邮件, 1-个人邮件 |
| title | string | 是 | 邮件标题（1-100字符） |
| content | string | 是 | 邮件内容（1-1000字符） |
| awards | string | 否 | 奖励JSON格式，最多10种道具 |
| startTime | datetime | 是 | 邮件生效开始时间 |
| endTime | datetime | 是 | 邮件生效结束时间 |
| targetUsers | array | 条件 | 个人邮件的目标用户ID列表（最多1000个） |

#### 参数验证规则
- **type**: 只能为0或1
- **个人邮件** (type=1): 必须提供targetUsers，且不能为空
- **全服邮件** (type=0): 忽略targetUsers参数
- **时间范围**: endTime必须晚于startTime，startTime不能早于昨天
- **奖励格式**: 必须是有效的JSON，道具ID和数量必须>0

#### 请求示例

**全服邮件示例**:
```bash
curl -X POST "http://localhost:8080/api/admin/mails/send" \
  -H "Authorization: Bearer admin-jwt-token" \
  -H "Content-Type: application/json" \
  -d '{
    "type": 0,
    "title": "系统维护补偿",
    "content": "感谢您的耐心等待，这是系统维护的补偿奖励。",
    "awards": "{\"props\":[{\"id\":1,\"cnt\":1000},{\"id\":2,\"cnt\":50}]}",
    "startTime": "2024-08-28T00:00:00Z",
    "endTime": "2024-09-04T23:59:59Z"
  }'
```

**个人邮件示例**:
```bash
curl -X POST "http://localhost:8080/api/admin/mails/send" \
  -H "Authorization: Bearer admin-jwt-token" \
  -H "Content-Type: application/json" \
  -d '{
    "type": 1,
    "title": "特殊奖励",
    "content": "恭喜您获得特殊成就，请查收奖励！",
    "awards": "{\"props\":[{\"id\":3,\"cnt\":500}]}",
    "startTime": "2024-08-28T10:00:00Z",
    "endTime": "2024-09-28T23:59:59Z",
    "targetUsers": [12345, 67890, 11111]
  }'
```

#### 响应示例

**全服邮件成功响应**:
```json
{
  "code": 200,
  "message": "发送成功",
  "data": {
    "mailId": 1001,
    "type": 0,
    "message": "全服邮件发送成功，用户登录时会自动收到"
  }
}
```

**个人邮件成功响应**:
```json
{
  "code": 200,
  "message": "发送成功",
  "data": {
    "mailId": 1002,
    "type": 1,
    "affectedUsers": 3,
    "message": "个人邮件发送成功，影响3个用户"
  }
}
```

#### 错误响应
```json
// 参数错误
{
  "code": 400,
  "message": "邮件类型错误，只支持0-全服邮件, 1-个人邮件"
}

// 个人邮件缺少目标用户
{
  "code": 400,
  "message": "个人邮件必须指定目标用户"
}

// 奖励格式错误
{
  "code": 400,
  "message": "奖励格式错误，正确格式: {\"props\":[{\"id\":1,\"cnt\":100}]}"
}

// 系统错误
{
  "code": 500,
  "message": "发送失败"
}
```

### 管理后台接口特点

1. **数据范围**: 仅返回用户邮件表中的记录，不返回邮件模板
2. **时间筛选**: 自动过滤已过期和已删除的邮件
3. **实时数据**: 显示用户对邮件的实际操作状态
4. **分页支持**: 支持大数据量的分页查询
5. **多条件搜索**: 支持按标题、用户ID等条件筛选
6. **邮件发送**: 支持全服邮件和个人邮件发送，个人邮件立即发送给指定用户
7. **参数验证**: 完整的输入参数验证，包括邮件类型、时间范围、奖励格式等
8. **操作日志**: 详细记录管理员操作，包括发送邮件、查询等行为

## 前端集成指南

### 1. 认证配置
```javascript
// 设置JWT Token
const token = 'your-jwt-token';
const headers = {
  'Authorization': `Bearer ${token}`,
  'Content-Type': 'application/json'
};
```

### 2. 获取邮件列表
```javascript
async function getMailList(userId) {
  try {
    const response = await fetch(`/api/mail/list?userid=${userId}`, {
      method: 'GET',
      headers: headers
    });
    
    const result = await response.json();
    if (result.code === 200) {
      return result.data.mails;
    } else {
      throw new Error(result.message);
    }
  } catch (error) {
    console.error('获取邮件列表失败:', error);
    throw error;
  }
}
```

### 3. 邮件状态处理
```javascript
// 邮件状态枚举
const MailStatus = {
  UNREAD: 0,    // 未读
  READ: 1,      // 已读
  CLAIMED: 2,   // 已领取
  DELETED: 3    // 已删除
};

// 根据状态显示不同的UI
function renderMailItem(mail) {
  const statusText = {
    [MailStatus.UNREAD]: '未读',
    [MailStatus.READ]: '已读',
    [MailStatus.CLAIMED]: '已领取',
    [MailStatus.DELETED]: '已删除'
  };
  
  return {
    ...mail,
    statusText: statusText[mail.status],
    canClaim: mail.status === MailStatus.READ && mail.awards
  };
}
```

### 4. 奖励解析
```javascript
// 解析奖励数据
function parseAwards(awardsString) {
  try {
    const awards = JSON.parse(awardsString);
    return awards.props || [];
  } catch (error) {
    console.error('解析奖励数据失败:', error);
    return [];
  }
}

// 使用示例
const mail = {
  awards: '{"props":[{"id":1,"cnt":1000},{"id":2,"cnt":50}]}'
};

const props = parseAwards(mail.awards);
// props = [{"id":1,"cnt":1000},{"id":2,"cnt":50}]
```

### 5. 完整的邮件管理类
```javascript
class MailManager {
  constructor(apiBase, token) {
    this.apiBase = apiBase;
    this.headers = {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    };
  }

  async getMailList(userId) {
    const response = await fetch(`${this.apiBase}/mail/list?userid=${userId}`, {
      method: 'GET',
      headers: this.headers
    });
    return await this.handleResponse(response);
  }

  async getMailDetail(mailId, userId) {
    const response = await fetch(`${this.apiBase}/mail/detail/${mailId}?userid=${userId}`, {
      method: 'GET',
      headers: this.headers
    });
    return await this.handleResponse(response);
  }

  async markAsRead(mailId, userId) {
    const response = await fetch(`${this.apiBase}/mail/read/${mailId}?userid=${userId}`, {
      method: 'POST',
      headers: this.headers
    });
    return await this.handleResponse(response);
  }

  async claimAward(mailId, userId) {
    const response = await fetch(`${this.apiBase}/mail/claim/${mailId}?userid=${userId}`, {
      method: 'POST',
      headers: this.headers
    });
    return await this.handleResponse(response);
  }

  async handleResponse(response) {
    const result = await response.json();
    if (result.code === 200) {
      return result.data;
    } else {
      throw new Error(result.message);
    }
  }
}

// 使用示例
const mailManager = new MailManager('/api', 'your-jwt-token');

// 获取邮件列表
const mails = await mailManager.getMailList(12345);

// 领取奖励
const awards = await mailManager.claimAward(1001, 12345);
```

## 测试用例

### 1. 邮件列表测试
```bash
# 正常获取邮件列表
curl -X GET "http://localhost:8080/api/mail/list?userid=12345" \
  -H "Authorization: Bearer valid-token"

# 缺少用户ID
curl -X GET "http://localhost:8080/api/mail/list" \
  -H "Authorization: Bearer valid-token"

# 无效用户ID
curl -X GET "http://localhost:8080/api/mail/list?userid=invalid" \
  -H "Authorization: Bearer valid-token"
```

### 2. 奖励领取测试
```bash
# 正常领取奖励
curl -X POST "http://localhost:8080/api/mail/claim/1001?userid=12345" \
  -H "Authorization: Bearer valid-token"

# 重复领取奖励
curl -X POST "http://localhost:8080/api/mail/claim/1001?userid=12345" \
  -H "Authorization: Bearer valid-token"

# 邮件不存在
curl -X POST "http://localhost:8080/api/mail/claim/99999?userid=12345" \
  -H "Authorization: Bearer valid-token"
```

## 注意事项

1. **时区处理**: 所有时间都使用UTC时间，前端需要根据用户时区进行转换
2. **邮件过期**: 系统会自动过滤过期邮件，不会返回给客户端
3. **并发安全**: 邮件领取操作使用数据库事务保证并发安全
4. **性能优化**: 建议对邮件列表接口进行适当的缓存
5. **数据同步**: 用户每次拉取邮件时都会自动同步最新的系统邮件

---

**文档版本**: v1.0  
**最后更新**: 2024-08-28  
**维护者**: gameWeb开发团队