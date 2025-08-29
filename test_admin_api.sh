#!/bin/bash

# 管理员API接口测试脚本

echo "=== 管理员API接口测试 ==="

# 服务器地址
SERVER="http://localhost:8080"

# 测试用的管理员账户信息
SUPER_ADMIN_USERNAME="admin"
SUPER_ADMIN_PASSWORD="123456"
ADMIN_TOKEN=""

# 1. 管理员登录
echo -e "\n1. 测试管理员登录:"
LOGIN_RESPONSE=$(curl -s -X POST "${SERVER}/api/admin/login" \
  -H "Content-Type: application/json" \
  -d "{\"username\": \"${SUPER_ADMIN_USERNAME}\", \"password\": \"${SUPER_ADMIN_PASSWORD}\"}")

echo "$LOGIN_RESPONSE" | jq '.' || echo "登录失败: $LOGIN_RESPONSE"

# 提取token（需要安装jq工具）
if command -v jq >/dev/null 2>&1; then
    ADMIN_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.token // empty')
    if [[ -n "$ADMIN_TOKEN" && "$ADMIN_TOKEN" != "null" ]]; then
        echo "✓ 获取到管理员Token"
    else
        echo "✗ 未能获取到Token，后续测试可能失败"
    fi
else
    echo "注意: 未安装jq工具，需要手动设置ADMIN_TOKEN"
fi

# 2. 获取当前管理员信息
echo -e "\n2. 测试获取管理员信息:"
curl -s -X GET "${SERVER}/api/admin/info" \
  -H "Authorization: Bearer ${ADMIN_TOKEN}" | jq '.' || echo "请求失败"

# 3. 获取管理员列表
echo -e "\n3. 测试获取管理员列表:"
curl -s -X GET "${SERVER}/api/admin/admins?page=1&pageSize=10" \
  -H "Authorization: Bearer ${ADMIN_TOKEN}" | jq '.' || echo "请求失败"

# 4. 测试搜索管理员
echo -e "\n4. 测试搜索管理员(关键词: admin):"
curl -s -X GET "${SERVER}/api/admin/admins?keyword=admin&status=1" \
  -H "Authorization: Bearer ${ADMIN_TOKEN}" | jq '.' || echo "请求失败"

# 5. 测试筛选超级管理员
echo -e "\n5. 测试筛选超级管理员:"
curl -s -X GET "${SERVER}/api/admin/admins?isSuperAdmin=1" \
  -H "Authorization: Bearer ${ADMIN_TOKEN}" | jq '.' || echo "请求失败"

# 6. 创建新管理员
echo -e "\n6. 测试创建新管理员:"
CREATE_RESPONSE=$(curl -s -X POST "${SERVER}/api/admin/create-admin" \
  -H "Authorization: Bearer ${ADMIN_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testadmin",
    "password": "password123",
    "email": "testadmin@example.com", 
    "realName": "测试管理员",
    "mobile": "13900139000",
    "isSuperAdmin": 0,
    "departmentId": 2,
    "note": "用于测试的管理员账户"
  }')

echo "$CREATE_RESPONSE" | jq '.' || echo "创建失败: $CREATE_RESPONSE"

# 提取新创建管理员的ID
if command -v jq >/dev/null 2>&1; then
    NEW_ADMIN_ID=$(echo "$CREATE_RESPONSE" | jq -r '.data.id // empty')
    if [[ -n "$NEW_ADMIN_ID" && "$NEW_ADMIN_ID" != "null" ]]; then
        echo "✓ 新管理员ID: $NEW_ADMIN_ID"
    fi
fi

# 7. 更新管理员信息
if [[ -n "$NEW_ADMIN_ID" && "$NEW_ADMIN_ID" != "null" ]]; then
    echo -e "\n7. 测试更新管理员信息:"
    curl -s -X PUT "${SERVER}/api/admin/update/${NEW_ADMIN_ID}" \
      -H "Authorization: Bearer ${ADMIN_TOKEN}" \
      -H "Content-Type: application/json" \
      -d '{
        "email": "updated@example.com",
        "mobile": "13800138000",
        "realName": "更新后的姓名",
        "note": "更新后的备注信息"
      }' | jq '.' || echo "更新失败"
else
    echo -e "\n7. 跳过更新测试（未获取到新管理员ID）"
fi

# 8. 测试删除管理员
if [[ -n "$NEW_ADMIN_ID" && "$NEW_ADMIN_ID" != "null" ]]; then
    echo -e "\n8. 测试删除管理员:"
    curl -s -X DELETE "${SERVER}/api/admin/delete/${NEW_ADMIN_ID}" \
      -H "Authorization: Bearer ${ADMIN_TOKEN}" | jq '.' || echo "删除失败"
else
    echo -e "\n8. 跳过删除测试（未获取到新管理员ID）"
fi

# 9. 测试错误情况
echo -e "\n9. 测试错误情况:"

echo -e "\n9.1 测试删除不存在的管理员:"
curl -s -X DELETE "${SERVER}/api/admin/delete/99999" \
  -H "Authorization: Bearer ${ADMIN_TOKEN}" | jq '.' || echo "请求失败"

echo -e "\n9.2 测试非超级管理员查看管理员列表（应该失败）:"
curl -s -X GET "${SERVER}/api/admin/admins" \
  -H "Authorization: Bearer invalid_token" | jq '.' || echo "预期失败"

echo -e "\n9.3 测试非超级管理员创建账户（应该失败）:"
curl -s -X POST "${SERVER}/api/admin/create-admin" \
  -H "Authorization: Bearer invalid_token" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "shouldfail",
    "password": "password123",
    "email": "shouldfail@example.com",
    "realName": "应该失败"
  }' | jq '.' || echo "预期失败"

echo -e "\n9.4 测试删除自己的账户（应该失败）:"
# 假设当前管理员ID为1
curl -s -X DELETE "${SERVER}/api/admin/delete/1" \
  -H "Authorization: Bearer ${ADMIN_TOKEN}" | jq '.' || echo "预期失败"

# 10. 管理员登出
echo -e "\n10. 测试管理员登出:"
curl -s -X POST "${SERVER}/api/admin/logout" \
  -H "Authorization: Bearer ${ADMIN_TOKEN}" | jq '.' || echo "登出失败"

echo -e "\n=== 测试完成 ==="

# 提供手动测试建议
echo -e "\n=== 手动测试建议 ==="
echo "1. 使用有效的管理员账户替换测试脚本中的用户名和密码"
echo "2. 确保数据库中有adminAccount表并包含测试数据"
echo "3. 验证JWT中间件和超级管理员权限检查是否正常工作"
echo "4. 测试邮箱唯一性验证"
echo "5. 测试最后一个超级管理员保护机制"