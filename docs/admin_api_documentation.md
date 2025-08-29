# 管理员API接口文档

## 接口概述

本文档描述了gameWeb项目中管理员相关的所有API接口，包括登录、登出、信息管理等功能。

## 基础信息

- **Base URL**: `http://localhost:8080/api/admin`
- **Content-Type**: `application/json`
- **认证方式**: JWT Bearer Token

---

## 1. 管理员登录

### 接口信息
- **接口地址**: `POST /api/admin/login`
- **认证要求**: 无
- **接口描述**: 管理员账户登录，获取JWT访问令牌

### 请求参数
| 参数名 | 类型 | 必填 | 长度限制 | 说明 |
|--------|------|------|----------|------|
| username | string | 是 | 3-50字符 | 管理员用户名 |
| password | string | 是 | 6-50字符 | 管理员密码 |

### 请求示例
```json
{
    "username": "admin",
    "password": "123456"
}
```

### 响应参数

#### 成功响应 (200)
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
            "mobile": "13800138000",
            "realName": "系统管理员",
            "avatar": "https://example.com/avatar.jpg",
            "departmentId": 1,
            "note": "超级管理员",
            "status": 1,
            "isSuperAdmin": true,
            "lastLoginIp": "127.0.0.1",
            "lastLoginTime": "2025-08-29T10:30:00Z",
            "createdTime": "2025-01-01T00:00:00Z",
            "updatedTime": "2025-08-29T10:30:00Z"
        }
    }
}
```

#### 错误响应
```json
// 参数错误 (400)
{
    "code": 400,
    "message": "参数错误: username不能为空"
}

// 认证失败 (401)
{
    "code": 401,
    "message": "用户名或密码错误"
}

// 账户被禁用 (401)
{
    "code": 401,
    "message": "账户已被禁用"
}

// 系统错误 (500)
{
    "code": 500,
    "message": "系统错误"
}
```

---

## 2. 管理员登出

### 接口信息
- **接口地址**: `POST /api/admin/logout`
- **认证要求**: 需要JWT认证
- **接口描述**: 管理员登出，清除服务端会话

### 请求参数
无需请求体参数

### 请求头
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| Authorization | string | 是 | Bearer {token} |

### 响应参数

#### 成功响应 (200)
```json
{
    "code": 200,
    "message": "登出成功"
}
```

#### 错误响应
```json
// 未登录 (401)
{
    "code": 401,
    "message": "未登录"
}
```

---

## 3. 获取当前管理员信息

### 接口信息
- **接口地址**: `GET /api/admin/info`
- **认证要求**: 需要JWT认证
- **接口描述**: 获取当前登录管理员的详细信息

### 请求参数
无需请求参数

### 请求头
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| Authorization | string | 是 | Bearer {token} |

### 响应参数

#### 成功响应 (200)
```json
{
    "code": 200,
    "message": "获取成功",
    "data": {
        "id": 1,
        "username": "admin",
        "email": "admin@example.com",
        "mobile": "13800138000",
        "realName": "系统管理员",
        "avatar": "https://example.com/avatar.jpg",
        "departmentId": 1,
        "note": "超级管理员",
        "status": 1,
        "isSuperAdmin": true,
        "lastLoginIp": "127.0.0.1",
        "lastLoginTime": "2025-08-29T10:30:00Z",
        "createdTime": "2025-01-01T00:00:00Z",
        "updatedTime": "2025-08-29T10:30:00Z"
    }
}
```

---

## 4. 创建管理员账户

### 接口信息
- **接口地址**: `POST /api/admin/create-admin`
- **认证要求**: 需要超级管理员JWT认证
- **接口描述**: 创建新的管理员账户（仅超级管理员可用）

### 请求参数
| 参数名 | 类型 | 必填 | 长度限制 | 说明 |
|--------|------|------|----------|------|
| username | string | 是 | 3-50字符 | 管理员用户名，必须唯一 |
| password | string | 是 | 6-50字符 | 管理员密码 |
| email | string | 是 | 有效邮箱格式 | 邮箱地址，必须唯一 |
| realName | string | 是 | 1-50字符 | 真实姓名 |
| mobile | string | 否 | 最大20字符 | 手机号码 |
| isSuperAdmin | integer | 否 | 0或1 | 是否为超级管理员，默认0 |
| departmentId | integer | 否 | - | 所属部门ID |
| note | string | 否 | 最大500字符 | 备注信息 |

### 请求示例
```json
{
    "username": "newadmin",
    "password": "password123",
    "email": "newadmin@example.com",
    "realName": "新管理员",
    "mobile": "13900139000",
    "isSuperAdmin": 0,
    "departmentId": 2,
    "note": "负责用户管理"
}
```

### 响应参数

#### 成功响应 (200)
```json
{
    "code": 200,
    "message": "创建成功",
    "data": {
        "id": 5,
        "username": "newadmin"
    }
}
```

#### 错误响应
```json
// 用户名已存在 (400)
{
    "code": 400,
    "message": "用户名已存在"
}

// 邮箱已存在 (400)
{
    "code": 400,
    "message": "邮箱已存在"
}

// 权限不足 (403)
{
    "code": 403,
    "message": "权限不足"
}
```

---

## 5. 更新管理员信息

### 接口信息
- **接口地址**: `PUT /api/admin/update/{id}`
- **认证要求**: 需要JWT认证
- **接口描述**: 更新管理员的基本信息，包括邮箱、手机号、真实姓名、头像、部门ID和备注信息

### 权限控制
- 普通管理员只能修改自己的信息
- 超级管理员可以修改任何人的信息

### 请求参数

#### 路径参数
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | integer | 是 | 要更新的管理员ID |

#### 请求头
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| Authorization | string | 是 | Bearer {token} |
| Content-Type | string | 是 | application/json |

#### 请求体
| 参数名 | 类型 | 必填 | 长度限制 | 说明 |
|--------|------|------|----------|------|
| email | string | 否 | 最大100字符 | 邮箱地址，必须是有效的邮箱格式 |
| mobile | string | 否 | 最大20字符 | 手机号码 |
| realName | string | 否 | 1-50字符 | 用户真实姓名 |
| avatar | string | 否 | 最大255字符 | 头像URL地址 |
| departmentId | integer | 否 | - | 所属部门ID |
| note | string | 否 | 最大500字符 | 备注信息 |

**注意：** 所有字段都是可选的，只需传入要更新的字段即可。

### 请求示例
```json
{
    "email": "newemail@example.com",
    "mobile": "13800138000",
    "realName": "张三",
    "avatar": "https://example.com/avatar.jpg"
}
```

### 响应参数

#### 成功响应 (200)
```json
{
    "code": 200,
    "message": "更新成功"
}
```

#### 错误响应
```json
// 参数错误 (400)
{
    "code": 400,
    "message": "参数错误: email格式不正确"
}

// 权限不足 (403)
{
    "code": 403,
    "message": "没有权限修改该管理员信息"
}

// 管理员不存在 (404)
{
    "code": 404,
    "message": "管理员不存在"
}

// 邮箱冲突 (400)
{
    "code": 400,
    "message": "邮箱已被其他管理员使用"
}
```

---

## 6. 删除管理员账户

### 接口信息
- **接口地址**: `DELETE /api/admin/delete/{id}`
- **认证要求**: 需要超级管理员JWT认证
- **接口描述**: 删除指定的管理员账户（仅超级管理员可用）

### 安全限制
- 只有超级管理员可以执行删除操作
- 不能删除自己的账户
- 不能删除最后一个超级管理员账户

### 请求参数

#### 路径参数
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | integer | 是 | 要删除的管理员ID |

#### 请求头
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| Authorization | string | 是 | Bearer {token} |

### 响应参数

#### 成功响应 (200)
```json
{
    "code": 200,
    "message": "删除成功"
}
```

#### 错误响应
```json
// 权限不足 (403)
{
    "code": 403,
    "message": "仅超级管理员可执行此操作"
}

// 禁止删除自己 (400)
{
    "code": 400,
    "message": "不能删除自己的账户"
}

// 禁止删除最后超级管理员 (400)
{
    "code": 400,
    "message": "不能删除最后一个超级管理员账户"
}

// 管理员不存在 (404)
{
    "code": 404,
    "message": "管理员不存在"
}
```

---

## 7. 获取管理员列表

### 接口信息
- **接口地址**: `GET /api/admin/admins`
- **认证要求**: 需要超级管理员JWT认证
- **接口描述**: 获取系统中所有管理员的列表，支持分页和筛选（仅超级管理员可用）

### 请求参数

#### 查询参数
| 参数名 | 类型 | 必填 | 默认值 | 说明 |
|--------|------|------|--------|------|
| page | integer | 否 | 1 | 页码，从1开始 |
| pageSize | integer | 否 | 20 | 每页数量，最大100 |
| keyword | string | 否 | - | 关键词搜索（用户名、邮箱、真实姓名） |
| status | integer | 否 | - | 管理员状态（0-禁用，1-正常） |
| isSuperAdmin | integer | 否 | - | 是否为超级管理员（0-普通，1-超级） |

#### 请求头
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| Authorization | string | 是 | Bearer {token} |

### 请求示例
```bash
# 基本查询
GET /api/admin/admins?page=1&pageSize=20

# 关键词搜索
GET /api/admin/admins?keyword=admin&page=1&pageSize=10

# 筛选超级管理员
GET /api/admin/admins?isSuperAdmin=1

# 筛选正常状态的管理员
GET /api/admin/admins?status=1&page=1&pageSize=50
```

### 响应参数

#### 成功响应 (200)
```json
{
    "code": 200,
    "message": "查询成功",
    "data": {
        "list": [
            {
                "id": 1,
                "username": "admin",
                "email": "admin@example.com",
                "mobile": "13800138000",
                "realName": "系统管理员",
                "avatar": "https://example.com/avatar1.jpg",
                "departmentId": 1,
                "note": "超级管理员",
                "status": 1,
                "isSuperAdmin": true,
                "lastLoginIp": "192.168.1.100",
                "lastLoginTime": "2025-08-29T10:30:00Z",
                "createdBy": null,
                "updatedBy": 1,
                "createdTime": "2025-01-01T00:00:00Z",
                "updatedTime": "2025-08-29T10:30:00Z"
            },
            {
                "id": 2,
                "username": "manager",
                "email": "manager@example.com",
                "mobile": "13900139000",
                "realName": "部门经理",
                "avatar": "https://example.com/avatar2.jpg",
                "departmentId": 2,
                "note": "负责用户管理",
                "status": 1,
                "isSuperAdmin": false,
                "lastLoginIp": "192.168.1.101",
                "lastLoginTime": "2025-08-28T14:20:00Z",
                "createdBy": 1,
                "updatedBy": 1,
                "createdTime": "2025-02-15T09:00:00Z",
                "updatedTime": "2025-08-28T14:20:00Z"
            }
        ],
        "total": 15,
        "page": 1,
        "pageSize": 20
    }
}
```

#### 错误响应
```json
// 未登录 (401)
{
    "code": 401,
    "message": "未登录"
}

// 权限不足 (403)
{
    "code": 403,
    "message": "仅超级管理员可查看管理员列表"
}

// 系统错误 (500)
{
    "code": 500,
    "message": "查询失败"
}
```

### 功能特性

1. **分页支持**：支持灵活的分页查询，默认每页20条
2. **多维搜索**：关键词同时搜索用户名、邮箱和真实姓名
3. **精确筛选**：支持按状态和管理员类型进行筛选
4. **安全保障**：不返回密码哈希等敏感信息
5. **排序规则**：按创建时间倒序排列，新的管理员在前

---

---

## 使用示例

### curl 示例

```bash
# 1. 管理员登录
curl -X POST "http://localhost:8080/api/admin/login" \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "123456"}'

# 2. 获取管理员信息
curl -X GET "http://localhost:8080/api/admin/info" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# 3. 获取管理员列表
curl -X GET "http://localhost:8080/api/admin/admins?page=1&pageSize=20" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# 4. 搜索管理员
curl -X GET "http://localhost:8080/api/admin/admins?keyword=admin&status=1" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# 5. 更新管理员信息
curl -X PUT "http://localhost:8080/api/admin/update/1" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{"email": "newemail@example.com", "realName": "新姓名"}'

# 6. 删除管理员
curl -X DELETE "http://localhost:8080/api/admin/delete/5" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# 7. 管理员登出
curl -X POST "http://localhost:8080/api/admin/logout" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### JavaScript 示例

```javascript
// 管理员登录
async function adminLogin(username, password) {
    const response = await fetch('/api/admin/login', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ username, password })
    });
    
    const result = await response.json();
    if (result.code === 200) {
        localStorage.setItem('adminToken', result.data.token);
        return result.data;
    }
    throw new Error(result.message);
}

// 更新管理员信息
async function updateAdminInfo(adminId, updateData) {
    const token = localStorage.getItem('adminToken');
    const response = await fetch(`/api/admin/update/${adminId}`, {
        method: 'PUT',
        headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(updateData)
    });
    
    const result = await response.json();
    if (result.code !== 200) {
        throw new Error(result.message);
    }
    return result;
}

// 获取管理员列表
async function getAdminList(params = {}) {
    const token = localStorage.getItem('adminToken');
    const queryString = new URLSearchParams(params).toString();
    const url = `/api/admin/admins${queryString ? '?' + queryString : ''}`;
    
    const response = await fetch(url, {
        method: 'GET',
        headers: {
            'Authorization': `Bearer ${token}`
        }
    });
    
    const result = await response.json();
    if (result.code !== 200) {
        throw new Error(result.message);
    }
    return result.data;
}

// 搜索管理员示例
async function searchAdmins(keyword, page = 1, pageSize = 20) {
    return await getAdminList({
        keyword: keyword,
        page: page,
        pageSize: pageSize,
        status: 1  // 只查询正常状态的管理员
    });
}

// 筛选超级管理员
async function getSuperAdmins() {
    return await getAdminList({
        isSuperAdmin: 1
    });
}

// 删除管理员
async function deleteAdmin(adminId) {
    const token = localStorage.getItem('adminToken');
    const response = await fetch(`/api/admin/delete/${adminId}`, {
        method: 'DELETE',
        headers: {
            'Authorization': `Bearer ${token}`
        }
    });
    
    const result = await response.json();
    if (result.code !== 200) {
        throw new Error(result.message);
    }
    return result;
}
```

## 注意事项

1. **安全性**：
   - 所有需要认证的接口都必须携带有效的JWT token
   - 系统会记录所有管理员操作的详细日志
   - 密码使用bcrypt进行安全哈希存储

2. **权限控制**：
   - 超级管理员具有所有权限
   - 普通管理员只能管理自己的信息
   - 删除操作仅限超级管理员

3. **数据一致性**：
   - 用户名和邮箱必须在系统中保持唯一
   - 不能删除最后一个超级管理员账户
   - 所有更新操作都会自动记录操作时间和操作者

4. **会话管理**：
   - JWT token具有过期时间
   - 登出操作会清除服务端Redis会话
   - 支持并发会话管理