#!/bin/bash

# 测试个人邮件和全服邮件的发送逻辑

echo "=== 测试邮件发送逻辑 ==="

# 服务器地址
SERVER="http://localhost:8080"

# 测试1: 发送全服邮件（会插入mailSystem表）
echo -e "\n1. 测试发送全服邮件（会插入mailSystem表）:"
curl -X POST "${SERVER}/api/admin/mails/send" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-admin-jwt-token" \
  -d '{
    "type": 0,
    "title": "全服邮件测试",
    "content": "这是一个全服邮件，会插入到mailSystem表",
    "awards": "{\"props\":[{\"id\":1,\"cnt\":100}]}",
    "startTime": "2025-08-29T10:00:00Z",
    "endTime": "2025-09-05T23:59:59Z"
  }' | jq '.' || echo "请求失败"

echo -e "\n"

# 测试2: 发送个人邮件（不会插入mailSystem表，直接发送给玩家）
echo -e "\n2. 测试发送个人邮件（不会插入mailSystem表）:"
curl -X POST "${SERVER}/api/admin/mails/send" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-admin-jwt-token" \
  -d '{
    "type": 1,
    "title": "个人邮件测试",
    "content": "这是一个个人邮件，不会插入到mailSystem表，直接发送给指定玩家",
    "awards": "{\"props\":[{\"id\":2,\"cnt\":200}]}",
    "startTime": "2025-08-29T10:00:00Z",
    "endTime": "2025-09-05T23:59:59Z",
    "targetUsers": [100001, 100002, 100003]
  }' | jq '.' || echo "请求失败"

echo -e "\n=== 测试完成 ==="

echo -e "\n=== 数据库验证SQL ==="
echo "-- 查看mails表中的邮件："
echo "SELECT id, type, title, content FROM mails ORDER BY id DESC LIMIT 5;"
echo ""
echo "-- 查看mailSystem表（应该只有全服邮件）："
echo "SELECT id, type, mailid, startTime, endTime FROM mailSystem ORDER BY id DESC LIMIT 5;"
echo ""
echo "-- 查看mailUsers表（应该只有个人邮件的用户记录）："
echo "SELECT id, userid, mailid, status FROM mailUsers ORDER BY id DESC LIMIT 10;"