#!/bin/bash

# 管理员信息更新接口测试脚本
# 使用方法: ./test_admin_update.sh

# 配置
BASE_URL="http://localhost:8080"
API_ENDPOINT="/api/admin"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== 管理员信息更新接口测试 ===${NC}"

# 检查jq是否安装
if ! command -v jq &> /dev/null; then
    echo -e "${RED}错误: 需要安装 jq 工具来解析JSON响应${NC}"
    echo "请运行: sudo apt-get install jq (Ubuntu/Debian) 或 brew install jq (macOS)"
    exit 1
fi

# 测试1: 管理员登录获取token
echo -e "\n${YELLOW}测试1: 管理员登录获取token${NC}"
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL$API_ENDPOINT/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "password123"
  }')

echo "登录响应: $LOGIN_RESPONSE"

# 提取token
TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.data.token // empty')
ADMIN_ID=$(echo $LOGIN_RESPONSE | jq -r '.data.adminInfo.id // empty')

if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
    echo -e "${RED}登录失败，无法获取token，请检查用户名和密码${NC}"
    echo "请确保数据库中存在测试管理员账户"
    exit 1
fi

echo -e "${GREEN}登录成功，获取到token: ${TOKEN:0:20}...${NC}"
echo -e "${GREEN}管理员ID: $ADMIN_ID${NC}"

# 测试2: 更新自己的基本信息
echo -e "\n${YELLOW}测试2: 更新自己的基本信息${NC}"
UPDATE_RESPONSE=$(curl -s -X PUT "$BASE_URL$API_ENDPOINT/update/$ADMIN_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "updated_admin@example.com",
    "mobile": "13800138000",
    "realName": "测试管理员",
    "note": "这是通过API更新的备注信息"
  }')

echo "更新响应: $UPDATE_RESPONSE"

UPDATE_CODE=$(echo $UPDATE_RESPONSE | jq -r '.code // empty')
if [ "$UPDATE_CODE" = "200" ]; then
    echo -e "${GREEN}✓ 更新成功${NC}"
else
    echo -e "${RED}✗ 更新失败${NC}"
fi

# 测试3: 只更新部分字段
echo -e "\n${YELLOW}测试3: 只更新部分字段（部门和头像）${NC}"
PARTIAL_UPDATE_RESPONSE=$(curl -s -X PUT "$BASE_URL$API_ENDPOINT/update/$ADMIN_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "departmentId": 1,
    "avatar": "https://example.com/avatar.jpg"
  }')

echo "部分更新响应: $PARTIAL_UPDATE_RESPONSE"

PARTIAL_CODE=$(echo $PARTIAL_UPDATE_RESPONSE | jq -r '.code // empty')
if [ "$PARTIAL_CODE" = "200" ]; then
    echo -e "${GREEN}✓ 部分更新成功${NC}"
else
    echo -e "${RED}✗ 部分更新失败${NC}"
fi

# 测试4: 测试邮箱格式验证
echo -e "\n${YELLOW}测试4: 测试邮箱格式验证（应该失败）${NC}"
INVALID_EMAIL_RESPONSE=$(curl -s -X PUT "$BASE_URL$API_ENDPOINT/update/$ADMIN_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "invalid-email-format"
  }')

echo "无效邮箱响应: $INVALID_EMAIL_RESPONSE"

INVALID_CODE=$(echo $INVALID_EMAIL_RESPONSE | jq -r '.code // empty')
if [ "$INVALID_CODE" = "400" ]; then
    echo -e "${GREEN}✓ 邮箱格式验证正常工作${NC}"
else
    echo -e "${RED}✗ 邮箱格式验证未生效${NC}"
fi

# 测试5: 测试权限控制（尝试修改不存在的管理员）
echo -e "\n${YELLOW}测试5: 测试权限控制（修改不存在的管理员）${NC}"
PERMISSION_RESPONSE=$(curl -s -X PUT "$BASE_URL$API_ENDPOINT/update/99999" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "realName": "尝试修改不存在的管理员"
  }')

echo "权限测试响应: $PERMISSION_RESPONSE"

PERMISSION_CODE=$(echo $PERMISSION_RESPONSE | jq -r '.code // empty')
if [ "$PERMISSION_CODE" = "404" ]; then
    echo -e "${GREEN}✓ 权限控制正常工作${NC}"
else
    echo -e "${RED}✗ 权限控制可能存在问题${NC}"
fi

# 测试6: 测试无token访问（应该失败）
echo -e "\n${YELLOW}测试6: 测试无token访问（应该失败）${NC}"
NO_TOKEN_RESPONSE=$(curl -s -X PUT "$BASE_URL$API_ENDPOINT/update/$ADMIN_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "realName": "无权限修改"
  }')

echo "无token响应: $NO_TOKEN_RESPONSE"

NO_TOKEN_CODE=$(echo $NO_TOKEN_RESPONSE | jq -r '.code // empty')
if [ "$NO_TOKEN_CODE" = "401" ]; then
    echo -e "${GREEN}✓ 身份验证正常工作${NC}"
else
    echo -e "${RED}✗ 身份验证可能存在问题${NC}"
fi

# 测试7: 获取管理员信息验证更新结果
echo -e "\n${YELLOW}测试7: 获取管理员信息验证更新结果${NC}"
INFO_RESPONSE=$(curl -s -X GET "$BASE_URL$API_ENDPOINT/info" \
  -H "Authorization: Bearer $TOKEN")

echo "管理员信息: $INFO_RESPONSE"

INFO_CODE=$(echo $INFO_RESPONSE | jq -r '.code // empty')
if [ "$INFO_CODE" = "200" ]; then
    CURRENT_EMAIL=$(echo $INFO_RESPONSE | jq -r '.data.email // empty')
    CURRENT_REALNAME=$(echo $INFO_RESPONSE | jq -r '.data.realName // empty')
    echo -e "${GREEN}✓ 当前邮箱: $CURRENT_EMAIL${NC}"
    echo -e "${GREEN}✓ 当前姓名: $CURRENT_REALNAME${NC}"
else
    echo -e "${RED}✗ 获取管理员信息失败${NC}"
fi

echo -e "\n${YELLOW}=== 测试完成 ===${NC}"
echo -e "${GREEN}所有测试已完成，请检查以上结果${NC}"
echo -e "${YELLOW}注意: 如果某些测试失败，请检查：${NC}"
echo "1. 服务器是否正在运行在 $BASE_URL"
echo "2. 数据库是否正确配置"
echo "3. 测试用户是否存在（用户名: admin, 密码: password123）"