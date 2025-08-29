#!/bin/bash

# 测试SendSystemMail接口参数绑定问题

echo "=== 测试SendSystemMail接口参数绑定 ==="

# 服务器地址
SERVER="http://localhost:8080"

# 测试1: 正确格式的请求
echo -e "\n1. 测试正确格式的请求:"
curl -X POST "${SERVER}/api/admin/mails/send" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-admin-jwt-token" \
  -d '{
    "type": 0,
    "title": "测试邮件标题",
    "content": "这是一个测试邮件内容",
    "awards": "{\"props\":[{\"id\":1,\"cnt\":100}]}",
    "startTime": "2025-08-29T10:00:00Z",
    "endTime": "2025-09-05T23:59:59Z"
  }' | jq '.' || echo "请求失败"

echo -e "\n"

# 测试2: 缺少type字段
echo -e "\n2. 测试缺少type字段:"
curl -X POST "${SERVER}/api/admin/mails/send" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-admin-jwt-token" \
  -d '{
    "title": "测试邮件标题",
    "content": "这是一个测试邮件内容",
    "awards": "",
    "startTime": "2025-08-29T10:00:00Z",
    "endTime": "2025-09-05T23:59:59Z"
  }' | jq '.' || echo "请求失败"

echo -e "\n"

# 测试3: type为null
echo -e "\n3. 测试type为null:"
curl -X POST "${SERVER}/api/admin/mails/send" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-admin-jwt-token" \
  -d '{
    "type": null,
    "title": "测试邮件标题", 
    "content": "这是一个测试邮件内容",
    "awards": "",
    "startTime": "2025-08-29T10:00:00Z",
    "endTime": "2025-09-05T23:59:59Z"
  }' | jq '.' || echo "请求失败"

echo -e "\n"

# 测试4: 空JSON
echo -e "\n4. 测试空JSON:"
curl -X POST "${SERVER}/api/admin/mails/send" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-admin-jwt-token" \
  -d '{}' | jq '.' || echo "请求失败"

echo -e "\n"

# 测试5: 错误的Content-Type
echo -e "\n5. 测试错误的Content-Type:"
curl -X POST "${SERVER}/api/admin/mails/send" \
  -H "Content-Type: text/plain" \
  -H "Authorization: Bearer your-admin-jwt-token" \
  -d '{
    "type": 0,
    "title": "测试邮件标题",
    "content": "这是一个测试邮件内容"
  }' | jq '.' || echo "请求失败"

echo -e "\n=== 测试完成 ==="