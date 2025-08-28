# 日志查询API接入文档

## 概述

本文档详细描述了游戏管理后台的日志查询相关API接口，包括用户登入认证日志查询、对局结果日志查询和相关统计功能。

## 基础信息

- **基础URL**: `http://your-domain:8080/api/admin/logs`
- **认证方式**: JWT Bearer Token（管理员专用）
- **数据格式**: JSON
- **字符编码**: UTF-8

## 数据库表结构

### 1. 登入认证日志表 (logAuth)

该表存储在 `gamelog` 数据库中，记录用户的登录认证行为。

**字段说明**:
- `id`: BIGINT UNSIGNED，主键，自增，日志ID
- `userid`: BIGINT，用户ID
- `nickname`: VARCHAR(64)，用户昵称
- `ip`: VARCHAR(50)，登录IP地址
- `loginType`: VARCHAR(32)，认证类型（渠道）
- `status`: TINYINT(1)，认证状态：0-失败，1-成功
- `ext`: VARCHAR(256)，扩展数据
- `create_time`: DATETIME，创建时间（默认当前时间）

### 2. 对局结果日志表 (logResult10001)

该表存储在 `gamelog` 数据库中，记录用户的游戏对局结果。

**字段说明**:
- `id`: BIGINT，主键，自增，日志ID
- `type`: TINYINT，计分类型（默认0）
- `userid`: BIGINT，用户ID
- `gameid`: BIGINT，游戏ID
- `roomid`: BIGINT，房间ID
- `result`: TINYINT，对局结果：0-无，1-赢，2-输，3-平，4-逃跑
- `score1`: BIGINT，财富1（默认0）
- `score2`: BIGINT，财富2（默认0）
- `score3`: BIGINT，财富3（默认0）
- `score4`: BIGINT，财富4（默认0）
- `score5`: BIGINT，财富5（默认0）
- `time`: TIMESTAMP，发生时间
- `ext`: TEXT，扩展数据

## API接口详情

### 1. 获取用户登入认证日志

#### 接口信息
- **URL**: `/api/admin/logs/auth`
- **方法**: GET
- **认证**: 需要管理员JWT认证

#### 请求参数

**查询参数**:
| 参数名 | 类型 | 必填 | 默认值 | 说明 |
|--------|------|------|--------|------|
| userid | integer | 否 | - | 用户ID，不传则查询所有用户 |
| startTime | string | 否 | - | 开始时间，ISO 8601格式 |
| endTime | string | 否 | - | 结束时间，ISO 8601格式 |
| page | integer | 否 | 1 | 页码，最小值为1 |
| pageSize | integer | 否 | 20 | 每页大小，范围1-100 |

#### 请求示例

```bash
# 查询特定用户的登录日志
curl -X GET "http://localhost:8080/api/admin/logs/auth?userid=12345&startTime=2024-01-01T00:00:00Z&endTime=2024-01-31T23:59:59Z&page=1&pageSize=10" \
  -H "Authorization: Bearer your-jwt-token"

# 查询所有用户的最近登录日志
curl -X GET "http://localhost:8080/api/admin/logs/auth?page=1&pageSize=20" \
  -H "Authorization: Bearer your-jwt-token"
```

#### 响应示例

**成功响应 (200)**:
```json
{
    "code": 200,
    "message": "获取成功",
    "data": {
        "total": 158,
        "page": 1,
        "pageSize": 10,
        "data": [
            {
                "id": 1001,
                "userid": 12345,
                "nickname": "测试用户1",
                "ip": "192.168.1.100",
                "loginType": "android",
                "status": 1,
                "ext": "",
                "createTime": "2024-01-15T10:30:00Z"
            },
            {
                "id": 1002,
                "userid": 12345,
                "nickname": "测试用户1",
                "ip": "192.168.1.101",
                "loginType": "ios",
                "status": 1,
                "ext": "",
                "createTime": "2024-01-15T14:20:00Z"
            }
        ]
    }
}
```

### 2. 获取用户对局结果日志

#### 接口信息
- **URL**: `/api/admin/logs/game`
- **方法**: GET
- **认证**: 需要管理员JWT认证

#### 请求参数

**查询参数**:
| 参数名 | 类型 | 必填 | 默认值 | 说明 |
|--------|------|------|--------|------|
| userid | integer | 否 | - | 用户ID，不传则查询所有用户 |
| startTime | string | 否 | - | 开始时间，ISO 8601格式 |
| endTime | string | 否 | - | 结束时间，ISO 8601格式 |
| page | integer | 否 | 1 | 页码，最小值为1 |
| pageSize | integer | 否 | 20 | 每页大小，范围1-100 |

#### 请求示例

```bash
# 查询特定用户的对局日志
curl -X GET "http://localhost:8080/api/admin/logs/game?userid=12345&startTime=2024-01-01T00:00:00Z&endTime=2024-01-31T23:59:59Z&page=1&pageSize=10" \
  -H "Authorization: Bearer your-jwt-token"

# 查询最近的对局日志
curl -X GET "http://localhost:8080/api/admin/logs/game?page=1&pageSize=20" \
  -H "Authorization: Bearer your-jwt-token"
```

#### 响应示例

**成功响应 (200)**:
```json
{
    "code": 200,
    "message": "获取成功",
    "data": {
        "total": 89,
        "page": 1,
        "pageSize": 10,
        "data": [
            {
                "id": 2001,
                "type": 0,
                "userid": 12345,
                "gameid": 10001,
                "roomid": 500123,
                "result": 1,
                "score1": 5000,
                "score2": 0,
                "score3": 0,
                "score4": 0,
                "score5": 0,
                "time": "2024-01-15T10:30:00Z",
                "ext": ""
            },
            {
                "id": 2002,
                "type": 0,
                "userid": 12345,
                "gameid": 10001,
                "roomid": 500124,
                "result": 2,
                "score1": -3000,
                "score2": 0,
                "score3": 0,
                "score4": 0,
                "score5": 0,
                "time": "2024-01-15T11:00:00Z",
                "ext": ""
            }
        ]
    }
}
```

### 3. 获取用户登录统计信息

#### 接口信息
- **URL**: `/api/admin/logs/login-stats`
- **方法**: GET
- **认证**: 需要管理员JWT认证

#### 请求参数

**查询参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| userid | integer | 是 | 用户ID |

#### 请求示例

```bash
curl -X GET "http://localhost:8080/api/admin/logs/login-stats?userid=12345" \
  -H "Authorization: Bearer your-jwt-token"
```

#### 响应示例

**成功响应 (200)**:
```json
{
    "code": 200,
    "message": "获取成功",
    "data": {
        "totalLogins": 245,
        "lastLoginTime": "2024-01-15T14:20:00Z",
        "todayLogins": 3,
        "weekLogins": 15,
        "successLogins": 230
    }
}
```

**字段说明**:
- `totalLogins`: 总登录次数
- `lastLoginTime`: 最后登录时间
- `todayLogins`: 今日登录次数
- `weekLogins`: 本周登录次数
- `successLogins`: 成功登录次数

### 4. 获取用户对局统计信息

#### 接口信息
- **URL**: `/api/admin/logs/game-stats`
- **方法**: GET
- **认证**: 需要管理员JWT认证

#### 请求参数

**查询参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| userid | integer | 是 | 用户ID |

#### 请求示例

```bash
curl -X GET "http://localhost:8080/api/admin/logs/game-stats?userid=12345" \
  -H "Authorization: Bearer your-jwt-token"
```

#### 响应示例

**成功响应 (200)**:
```json
{
    "code": 200,
    "message": "获取成功",
    "data": {
        "totalGames": 128,
        "winGames": 76,
        "winRate": "59.38%",
        "totalScore1": 150000,
        "totalScore2": 89000,
        "totalScore3": 12000,
        "totalScore4": 5000,
        "totalScore5": 2000,
        "totalScore": 258000,
        "lastGameTime": "2024-01-15T11:20:00Z",
        "todayGames": 5
    }
}
```

**字段说明**:
- `totalGames`: 总对局次数
- `winGames`: 胜利次数
- `winRate`: 胜率百分比
- `totalScore1`: 财富1总计
- `totalScore2`: 财富2总计
- `totalScore3`: 财富3总计
- `totalScore4`: 财富4总计
- `totalScore5`: 财富5总计
- `totalScore`: 所有财富总计
- `lastGameTime`: 最后对局时间
- `todayGames`: 今日对局次数

## 错误响应

### 通用错误响应

#### 参数错误 (400)
```json
{
    "code": 400,
    "message": "参数错误: 无效的用户ID"
}
```

#### 未认证 (401)
```json
{
    "code": 401,
    "message": "未登录"
}
```

#### 权限不足 (403)
```json
{
    "code": 403,
    "message": "权限不足"
}
```

#### 系统错误 (500)
```json
{
    "code": 500,
    "message": "系统错误"
}
```

## 业务规则

### 1. 权限控制
- 所有日志查询接口都需要管理员JWT认证
- 只有认证的管理员才能访问日志数据

### 2. 分页限制
- 页码最小值为1
- 每页大小范围为1-100，默认20
- 超出范围会自动调整到边界值

### 3. 时间过滤
- `startTime` 和 `endTime` 支持ISO 8601格式
- 时间范围过滤是可选的，不传则查询所有时间范围
- 登录日志使用 `create_time` 字段进行时间过滤
- 对局日志使用 `time` 字段进行时间过滤

### 4. 数据返回
- 登录日志按 `create_time` 降序排列（最新的在前）
- 对局日志按 `time` 降序排列（最新的在前）
- 统计数据实时计算，反映当前最新状态

## JavaScript SDK 示例

### 基础配置
```javascript
class LogAPI {
    constructor(baseURL, token) {
        this.baseURL = baseURL;
        this.token = token;
    }

    async request(url, options = {}) {
        const response = await fetch(`${this.baseURL}${url}`, {
            ...options,
            headers: {
                'Authorization': `Bearer ${this.token}`,
                'Content-Type': 'application/json',
                ...options.headers
            }
        });
        
        const result = await response.json();
        if (result.code !== 200) {
            throw new Error(result.message);
        }
        return result.data;
    }
}
```

### 获取登录日志
```javascript
// 初始化API客户端
const logAPI = new LogAPI('http://localhost:8080/api/admin/logs', 'your-jwt-token');

// 获取用户登录日志
async function getUserAuthLogs(userid, startTime, endTime, page = 1, pageSize = 20) {
    const params = new URLSearchParams({
        page: page.toString(),
        pageSize: pageSize.toString()
    });
    
    if (userid) params.append('userid', userid.toString());
    if (startTime) params.append('startTime', startTime);
    if (endTime) params.append('endTime', endTime);
    
    try {
        const data = await logAPI.request(`/auth?${params}`);
        console.log('登录日志数据:', data);
        return data;
    } catch (error) {
        console.error('获取登录日志失败:', error);
        throw error;
    }
}

// 使用示例
getUserAuthLogs(12345, '2024-01-01T00:00:00Z', '2024-01-31T23:59:59Z', 1, 10);
```

### 获取对局日志
```javascript
// 获取用户对局日志
async function getUserGameLogs(userid, startTime, endTime, page = 1, pageSize = 20) {
    const params = new URLSearchParams({
        page: page.toString(),
        pageSize: pageSize.toString()
    });
    
    if (userid) params.append('userid', userid.toString());
    if (startTime) params.append('startTime', startTime);
    if (endTime) params.append('endTime', endTime);
    
    try {
        const data = await logAPI.request(`/game?${params}`);
        console.log('对局日志数据:', data);
        return data;
    } catch (error) {
        console.error('获取对局日志失败:', error);
        throw error;
    }
}

// 使用示例
getUserGameLogs(12345, '2024-01-01T00:00:00Z', '2024-01-31T23:59:59Z', 1, 10);
```

### 获取统计信息
```javascript
// 获取登录统计
async function getUserLoginStats(userid) {
    try {
        const data = await logAPI.request(`/login-stats?userid=${userid}`);
        console.log('登录统计:', data);
        return data;
    } catch (error) {
        console.error('获取登录统计失败:', error);
        throw error;
    }
}

// 获取对局统计
async function getUserGameStats(userid) {
    try {
        const data = await logAPI.request(`/game-stats?userid=${userid}`);
        console.log('对局统计:', data);
        return data;
    } catch (error) {
        console.error('获取对局统计失败:', error);
        throw error;
    }
}

// 使用示例
getUserLoginStats(12345);
getUserGameStats(12345);
```

## 注意事项

1. **性能考虑**：日志数据量可能很大，建议合理使用分页和时间过滤
2. **时区处理**：所有时间都是UTC时间，前端需要根据用户时区进行转换
3. **数据安全**：日志包含敏感信息，请确保在安全环境下使用
4. **缓存策略**：统计数据可以考虑适当缓存以提升性能
5. **监控报警**：建议对日志查询接口进行监控，防止异常访问
