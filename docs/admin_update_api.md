# 管理员信息更新接口文档

## 接口概述

### 接口名称
更新管理员信息

### 接口描述
用于更新管理员的基本信息，包括邮箱、手机号、真实姓名、头像、部门ID和备注信息。

### 接口地址
```
PUT /api/admin/update/{id}
```

### 请求方式
PUT

### 认证要求
- 需要管理员JWT认证
- 只能修改自己的信息，或者超级管理员可以修改任何人的信息

## 请求参数

### 路径参数
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | integer | 是 | 要更新的管理员ID |

### 请求头
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| Authorization | string | 是 | Bearer {token} |
| Content-Type | string | 是 | application/json |

### 请求体
| 参数名 | 类型 | 必填 | 长度限制 | 说明 |
|--------|------|------|----------|------|
| email | string | 否 | 最大100字符 | 邮箱地址，必须是有效的邮箱格式 |
| mobile | string | 否 | 最大20字符 | 手机号码 |
| realName | string | 否 | 1-50字符 | 用户真实姓名 |
| avatar | string | 否 | 最大255字符 | 头像URL地址 |
| departmentId | integer | 否 | - | 所属部门ID |
| note | string | 否 | 最大500字符 | 备注信息 |

**注意：** 所有字段都是可选的，只需传入要更新的字段即可。

## 响应参数

### 成功响应
```json
{
    "code": 200,
    "message": "更新成功"
}
```

### 错误响应

#### 参数错误 (400)
```json
{
    "code": 400,
    "message": "参数错误: email格式不正确"
}
```

#### 未登录 (401)
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
    "message": "没有权限修改该管理员信息"
}
```

#### 管理员不存在 (404)
```json
{
    "code": 404,
    "message": "管理员不存在"
}
```

#### 邮箱冲突 (400)
```json
{
    "code": 400,
    "message": "邮箱已被其他管理员使用"
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

1. **权限控制**：
   - 普通管理员只能修改自己的信息（id必须与当前登录管理员一致）
   - 超级管理员可以修改任何管理员的信息

2. **邮箱唯一性**：
   - 如果要更新邮箱，系统会检查邮箱是否已被其他管理员使用
   - 同一个邮箱不能被多个管理员使用

3. **字段验证**：
   - 邮箱字段必须符合邮箱格式规范
   - 真实姓名不能为空字符串
   - 各字段长度必须符合限制要求

4. **部分更新**：
   - 支持部分字段更新，只需传入要修改的字段
   - 未传入的字段保持原值不变

## 请求示例

### 示例1：更新自己的基本信息
```bash
curl -X PUT "http://localhost:8080/api/admin/update/1" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "email": "newemail@example.com",
    "mobile": "13800138000",
    "realName": "张三",
    "avatar": "https://example.com/avatar.jpg"
  }'
```

### 示例2：只更新部门和备注
```bash
curl -X PUT "http://localhost:8080/api/admin/update/1" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "departmentId": 2,
    "note": "负责用户管理模块"
  }'
```

### 示例3：清空某个字段（传入空字符串或null）
```bash
curl -X PUT "http://localhost:8080/api/admin/update/1" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "mobile": "",
    "note": ""
  }'
```

## JavaScript 示例

### 使用 fetch API
```javascript
// 更新管理员信息
async function updateAdminInfo(adminId, updateData) {
    try {
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
        
        if (result.code === 200) {
            console.log('更新成功');
            return result;
        } else {
            console.error('更新失败:', result.message);
            throw new Error(result.message);
        }
    } catch (error) {
        console.error('请求失败:', error);
        throw error;
    }
}

// 使用示例
updateAdminInfo(1, {
    email: 'newemail@example.com',
    realName: '李四',
    departmentId: 3
}).then(result => {
    alert('信息更新成功');
}).catch(error => {
    alert('更新失败: ' + error.message);
});
```

### 使用 axios
```javascript
import axios from 'axios';

// 配置axios拦截器添加token
axios.interceptors.request.use(config => {
    const token = localStorage.getItem('adminToken');
    if (token) {
        config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
});

// 更新管理员信息
const updateAdminInfo = async (adminId, updateData) => {
    try {
        const response = await axios.put(`/api/admin/update/${adminId}`, updateData);
        return response.data;
    } catch (error) {
        if (error.response) {
            throw new Error(error.response.data.message);
        }
        throw error;
    }
};

// 使用示例
updateAdminInfo(1, {
    mobile: '13900139000',
    avatar: 'https://newavatar.example.com/image.jpg'
}).then(() => {
    console.log('更新成功');
}).catch(error => {
    console.error('更新失败:', error.message);
});
```

## 注意事项

1. **安全性**：
   - 所有请求都需要有效的JWT token
   - 系统会记录操作日志，包括操作者和被操作者的信息

2. **幂等性**：
   - 该接口是幂等的，多次调用相同参数会得到相同结果
   - 如果没有任何字段需要更新，接口会直接返回成功

3. **字段限制**：
   - 不能通过此接口修改用户名、密码、状态等敏感字段
   - 这些字段需要使用专门的接口进行修改

4. **数据库事务**：
   - 更新操作会自动记录更新时间和更新者信息
   - 所有字段更新在同一个事务中完成，保证数据一致性

## 测试用例

### 正常情况测试
- 普通管理员更新自己的信息
- 超级管理员更新其他管理员的信息
- 部分字段更新
- 清空可选字段

### 异常情况测试
- 未登录访问
- 普通管理员尝试修改其他人信息
- 使用已存在的邮箱
- 提供无效的参数格式
- 访问不存在的管理员ID