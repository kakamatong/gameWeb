# 日志API快速参考

## 概述

本文档提供游戏管理后台日志查询API的快速参考，包括登入认证日志和对局结果日志的查询接口。

## 基础信息

- **认证方式**: Bearer Token (管理员JWT)
- **基础URL**: `http://your-domain:8080/api/admin/logs`
- **数据格式**: JSON

## API接口列表

### 1. 登入认证日志

| 接口 | 方法 | 路径 | 描述 |
|------|------|------|------|
| 获取登录日志 | GET | `/auth` | 分页查询用户登录日志 |
| 登录统计 | GET | `/login-stats` | 获取用户登录统计信息 |

### 2. 对局结果日志

| 接口 | 方法 | 路径 | 描述 |
|------|------|------|------|
| 获取对局日志 | GET | `/game` | 分页查询用户对局日志 |
| 对局统计 | GET | `/game-stats` | 获取用户对局统计信息 |

## 快速示例

### 获取登录日志
```bash
curl -X GET "http://localhost:8080/api/admin/logs/auth?userid=12345&page=1&pageSize=10" \
  -H "Authorization: Bearer your-jwt-token"
```

### 获取对局日志
```bash
curl -X GET "http://localhost:8080/api/admin/logs/game?userid=12345&page=1&pageSize=10" \
  -H "Authorization: Bearer your-jwt-token"
```

### 获取登录统计
```bash
curl -X GET "http://localhost:8080/api/admin/logs/login-stats?userid=12345" \
  -H "Authorization: Bearer your-jwt-token"
```

### 获取对局统计
```bash
curl -X GET "http://localhost:8080/api/admin/logs/game-stats?userid=12345" \
  -H "Authorization: Bearer your-jwt-token"
```

## 通用查询参数

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| userid | integer | 否 | - | 用户ID |
| startTime | string | 否 | - | 开始时间 (ISO 8601) |
| endTime | string | 否 | - | 结束时间 (ISO 8601) |
| page | integer | 否 | 1 | 页码 |
| pageSize | integer | 否 | 20 | 每页大小 (1-100) |

## 响应格式

### 成功响应
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "total": 100,
    "page": 1,
    "pageSize": 20,
    "data": [...]
  }
}
```

### 错误响应
```json
{
  "code": 400,
  "message": "参数错误"
}
```

## 错误码

| 错误码 | 说明 |
|--------|------|
| 400 | 参数错误 |
| 401 | 未认证 |
| 403 | 权限不足 |
| 500 | 系统错误 |

## 测试脚本

- [`test_auth_logs_api.sh`](../test/test_auth_logs_api.sh) - 登入认证日志API测试
- [`test_game_logs_api.sh`](../test/test_game_logs_api.sh) - 对局结果日志API测试  
- [`test_all_logs_api.sh`](../test/test_all_logs_api.sh) - 综合测试套件

## 详细文档

参考 [完整日志API文档](./log_api_documentation.md) 获取详细信息。