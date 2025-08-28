# 获取管理员信息接口文档

## 接口概述

### 接口名称
获取当前登录管理员的详细信息

### 接口描述
用于获取当前登录管理员的完整个人信息，包括基本信息、联系方式、权限信息、登录记录和创建/更新记录等。

### 接口地址
```
GET /api/admin/info
```

### 请求方式
GET

### 认证要求
- 需要管理员JWT认证
- 返回当前登录管理员的信息

## 请求参数

### 请求头
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| Authorization | string | 是 | Bearer {token} |

### 请求参数
无需额外参数，直接GET请求即可

## 响应参数

### 成功响应 (200)
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
        "note": "系统超级管理员账户",
        "status": 1,
        "isSuperAdmin": true,
        "lastLoginIp": "192.168.1.100",
        "lastLoginTime": "2024-08-28T10:30:00Z",
        "createdBy": null,
        "updatedBy": 1,
        "createdTime": "2024-01-01T00:00:00Z",
        "updatedTime": "2024-08-28T09:15:00Z"
    }
}
```

### 响应字段说明
| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | integer | 管理员唯一ID |
| username | string | 登录用户名 |
| email | string | 邮箱地址 |
| mobile | string | 手机号码 |
| realName | string | 真实姓名 |
| avatar | string | 头像URL地址 |
| departmentId | integer | 所属部门ID，可能为null |
| note | string | 备注信息 |
| status | integer | 账户状态：0-禁用，1-启用 |
| isSuperAdmin | boolean | 是否是超级管理员 |
| lastLoginIp | string | 最后一次登录IP地址 |
| lastLoginTime | string | 最后一次登录时间（ISO 8601格式） |
| createdBy | integer | 创建者ID，可能为null（系统初始账户） |
| updatedBy | integer | 最后修改者ID，可能为null |
| createdTime | string | 创建时间（ISO 8601格式） |
| updatedTime | string | 最后更新时间（ISO 8601格式） |

### 错误响应

#### 未登录 (401)
```json
{
    "code": 401,
    "message": "未登录"
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

1. **身份验证**：
   - 必须提供有效的JWT token
   - 只能获取当前登录管理员的信息

2. **数据完整性**：
   - 返回管理员的完整信息（除密码外）
   - 包含账户创建和修改的审计信息

3. **权限信息**：
   - 包含管理员权限级别信息
   - 可用于前端权限控制

4. **登录记录**：
   - 包含最后登录时间和IP地址
   - 便于安全审计

## 前端使用指南

### 数据绑定示例

#### Vue.js 示例
```vue
<template>
  <div class="admin-profile">
    <div class="profile-header">
      <img :src="adminInfo.avatar || defaultAvatar" 
           :alt="adminInfo.realName" 
           class="avatar">
      <div class="basic-info">
        <h2>{{ adminInfo.realName }}</h2>
        <p class="username">@{{ adminInfo.username }}</p>
        <span class="badge" :class="adminInfo.isSuperAdmin ? 'super-admin' : 'admin'">
          {{ adminInfo.isSuperAdmin ? '超级管理员' : '普通管理员' }}
        </span>
      </div>
    </div>
    
    <div class="profile-details">
      <div class="detail-section">
        <h3>联系信息</h3>
        <div class="detail-item">
          <label>邮箱：</label>
          <span>{{ adminInfo.email }}</span>
        </div>
        <div class="detail-item">
          <label>手机：</label>
          <span>{{ adminInfo.mobile || '未设置' }}</span>
        </div>
      </div>
      
      <div class="detail-section">
        <h3>账户信息</h3>
        <div class="detail-item">
          <label>状态：</label>
          <span :class="adminInfo.status === 1 ? 'status-active' : 'status-inactive'">
            {{ adminInfo.status === 1 ? '正常' : '禁用' }}
          </span>
        </div>
        <div class="detail-item">
          <label>部门ID：</label>
          <span>{{ adminInfo.departmentId || '未分配' }}</span>
        </div>
        <div class="detail-item" v-if="adminInfo.note">
          <label>备注：</label>
          <span>{{ adminInfo.note }}</span>
        </div>
      </div>
      
      <div class="detail-section">
        <h3>登录记录</h3>
        <div class="detail-item">
          <label>最后登录时间：</label>
          <span>{{ formatDate(adminInfo.lastLoginTime) }}</span>
        </div>
        <div class="detail-item">
          <label>最后登录IP：</label>
          <span>{{ adminInfo.lastLoginIp }}</span>
        </div>
      </div>
      
      <div class="detail-section">
        <h3>账户历史</h3>
        <div class="detail-item">
          <label>创建时间：</label>
          <span>{{ formatDate(adminInfo.createdTime) }}</span>
        </div>
        <div class="detail-item">
          <label>最后更新：</label>
          <span>{{ formatDate(adminInfo.updatedTime) }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'AdminProfile',
  data() {
    return {
      adminInfo: {},
      defaultAvatar: '/default-avatar.png'
    }
  },
  async mounted() {
    await this.fetchAdminInfo();
  },
  methods: {
    async fetchAdminInfo() {
      try {
        const response = await this.$http.get('/api/admin/info');
        if (response.data.code === 200) {
          this.adminInfo = response.data.data;
        } else {
          this.$message.error(response.data.message);
        }
      } catch (error) {
        console.error('获取管理员信息失败:', error);
        this.$message.error('获取管理员信息失败');
      }
    },
    formatDate(dateString) {
      if (!dateString) return '未知';
      return new Date(dateString).toLocaleString('zh-CN');
    }
  }
}
</script>
```

#### React 示例
```jsx
import React, { useState, useEffect } from 'react';
import axios from 'axios';

const AdminProfile = () => {
  const [adminInfo, setAdminInfo] = useState({});
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchAdminInfo();
  }, []);

  const fetchAdminInfo = async () => {
    try {
      setLoading(true);
      const response = await axios.get('/api/admin/info');
      if (response.data.code === 200) {
        setAdminInfo(response.data.data);
      } else {
        console.error('获取失败:', response.data.message);
      }
    } catch (error) {
      console.error('请求失败:', error);
    } finally {
      setLoading(false);
    }
  };

  const formatDate = (dateString) => {
    if (!dateString) return '未知';
    return new Date(dateString).toLocaleString('zh-CN');
  };

  if (loading) {
    return <div>加载中...</div>;
  }

  return (
    <div className="admin-profile">
      <div className="profile-header">
        <img 
          src={adminInfo.avatar || '/default-avatar.png'} 
          alt={adminInfo.realName}
          className="avatar"
        />
        <div className="basic-info">
          <h2>{adminInfo.realName}</h2>
          <p className="username">@{adminInfo.username}</p>
          <span className={`badge ${adminInfo.isSuperAdmin ? 'super-admin' : 'admin'}`}>
            {adminInfo.isSuperAdmin ? '超级管理员' : '普通管理员'}
          </span>
        </div>
      </div>
      
      <div className="profile-details">
        <div className="detail-section">
          <h3>联系信息</h3>
          <div className="detail-item">
            <label>邮箱：</label>
            <span>{adminInfo.email}</span>
          </div>
          <div className="detail-item">
            <label>手机：</label>
            <span>{adminInfo.mobile || '未设置'}</span>
          </div>
        </div>
        
        <div className="detail-section">
          <h3>账户信息</h3>
          <div className="detail-item">
            <label>状态：</label>
            <span className={adminInfo.status === 1 ? 'status-active' : 'status-inactive'}>
              {adminInfo.status === 1 ? '正常' : '禁用'}
            </span>
          </div>
          <div className="detail-item">
            <label>部门ID：</label>
            <span>{adminInfo.departmentId || '未分配'}</span>
          </div>
          {adminInfo.note && (
            <div className="detail-item">
              <label>备注：</label>
              <span>{adminInfo.note}</span>
            </div>
          )}
        </div>
        
        <div className="detail-section">
          <h3>登录记录</h3>
          <div className="detail-item">
            <label>最后登录时间：</label>
            <span>{formatDate(adminInfo.lastLoginTime)}</span>
          </div>
          <div className="detail-item">
            <label>最后登录IP：</label>
            <span>{adminInfo.lastLoginIp}</span>
          </div>
        </div>
      </div>
    </div>
  );
};

export default AdminProfile;
```

## JavaScript API 调用示例

### 使用 fetch API
```javascript
// 获取管理员信息
async function getAdminInfo() {
    try {
        const token = localStorage.getItem('adminToken');
        const response = await fetch('/api/admin/info', {
            method: 'GET',
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            }
        });
        
        const result = await response.json();
        
        if (result.code === 200) {
            console.log('管理员信息:', result.data);
            return result.data;
        } else {
            console.error('获取失败:', result.message);
            throw new Error(result.message);
        }
    } catch (error) {
        console.error('请求失败:', error);
        throw error;
    }
}

// 使用示例
getAdminInfo()
    .then(adminInfo => {
        // 更新页面显示
        document.getElementById('admin-name').textContent = adminInfo.realName;
        document.getElementById('admin-email').textContent = adminInfo.email;
        
        // 根据权限显示/隐藏功能
        if (adminInfo.isSuperAdmin) {
            document.getElementById('super-admin-panel').style.display = 'block';
        }
        
        // 显示头像
        if (adminInfo.avatar) {
            document.getElementById('admin-avatar').src = adminInfo.avatar;
        }
    })
    .catch(error => {
        alert('获取管理员信息失败: ' + error.message);
    });
```

### 使用 axios
```javascript
import axios from 'axios';

// 配置axios拦截器
axios.interceptors.request.use(config => {
    const token = localStorage.getItem('adminToken');
    if (token) {
        config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
});

// 获取管理员信息
const getAdminInfo = async () => {
    try {
        const response = await axios.get('/api/admin/info');
        return response.data.data;
    } catch (error) {
        if (error.response && error.response.data) {
            throw new Error(error.response.data.message);
        }
        throw error;
    }
};

// 使用示例
getAdminInfo()
    .then(adminInfo => {
        console.log('管理员信息:', adminInfo);
        
        // 权限检查示例
        const hasPermission = (requiredLevel) => {
            if (requiredLevel === 'super' && !adminInfo.isSuperAdmin) {
                return false;
            }
            return adminInfo.status === 1; // 账户必须是启用状态
        };
        
        // 根据权限显示功能
        if (hasPermission('super')) {
            enableSuperAdminFeatures();
        }
        
        // 显示用户信息
        displayUserInfo(adminInfo);
    })
    .catch(error => {
        console.error('获取管理员信息失败:', error.message);
    });

// 显示用户信息的辅助函数
function displayUserInfo(adminInfo) {
    const userCard = document.createElement('div');
    userCard.innerHTML = `
        <div class="user-card">
            <img src="${adminInfo.avatar || '/default-avatar.png'}" alt="头像">
            <h3>${adminInfo.realName}</h3>
            <p>用户名: ${adminInfo.username}</p>
            <p>邮箱: ${adminInfo.email}</p>
            <p>权限: ${adminInfo.isSuperAdmin ? '超级管理员' : '普通管理员'}</p>
            <p>最后登录: ${new Date(adminInfo.lastLoginTime).toLocaleString()}</p>
        </div>
    `;
    document.body.appendChild(userCard);
}
```

## 测试用例

### cURL 测试
```bash
# 获取管理员信息
curl -X GET "http://localhost:8080/api/admin/info" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json"

# 预期成功响应
# {
#   "code": 200,
#   "message": "获取成功",
#   "data": {
#     "id": 1,
#     "username": "admin",
#     "email": "admin@example.com",
#     ...
#   }
# }
```

### Postman 测试
1. **请求类型**: GET
2. **URL**: `{{baseUrl}}/api/admin/info`
3. **Headers**:
   - `Authorization`: `Bearer {{adminToken}}`
   - `Content-Type`: `application/json`
4. **测试脚本**:
```javascript
pm.test("Status code is 200", function () {
    pm.response.to.have.status(200);
});

pm.test("Response has correct structure", function () {
    const jsonData = pm.response.json();
    pm.expect(jsonData).to.have.property('code', 200);
    pm.expect(jsonData).to.have.property('message', '获取成功');
    pm.expect(jsonData).to.have.property('data');
    
    const adminInfo = jsonData.data;
    pm.expect(adminInfo).to.have.property('id');
    pm.expect(adminInfo).to.have.property('username');
    pm.expect(adminInfo).to.have.property('email');
    pm.expect(adminInfo).to.have.property('realName');
    pm.expect(adminInfo).to.have.property('isSuperAdmin');
});
```

## 注意事项

1. **安全性**：
   - 接口不返回密码字段
   - 只能获取当前登录用户的信息
   - 需要有效的JWT认证

2. **缓存策略**：
   - 建议前端缓存管理员信息
   - 在信息更新后及时刷新缓存

3. **权限控制**：
   - 可根据 `isSuperAdmin` 字段控制前端功能显示
   - 根据 `status` 字段判断账户状态

4. **数据处理**：
   - 时间字段为ISO 8601格式，需要转换为本地时间显示
   - 部分字段可能为null，需要做空值处理

5. **错误处理**：
   - 401错误表示token失效，需要重新登录
   - 500错误表示服务器异常，建议重试或联系管理员

## 常见问题

### Q: 为什么有些字段返回null？
A: 某些可选字段（如departmentId、createdBy等）在创建时可能未设置，因此返回null。

### Q: 如何判断管理员权限？
A: 通过 `isSuperAdmin` 字段判断是否为超级管理员，通过 `status` 字段判断账户是否启用。

### Q: 时间格式如何处理？
A: 返回的时间为UTC时间的ISO 8601格式，前端需要转换为本地时间显示。

### Q: 头像字段为空怎么处理？
A: 当avatar字段为空时，建议显示默认头像。