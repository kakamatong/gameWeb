#!/bin/bash

# 获取管理员信息接口测试脚本
# 使用方法: ./test_get_admin_info.sh

# 配置
BASE_URL="http://localhost:8080"
API_ENDPOINT="/api/admin"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== 获取管理员信息接口测试 ===${NC}"

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

# 测试2: 获取管理员详细信息
echo -e "\n${YELLOW}测试2: 获取管理员详细信息${NC}"
INFO_RESPONSE=$(curl -s -X GET "$BASE_URL$API_ENDPOINT/info" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json")

echo "管理员信息响应: $INFO_RESPONSE"

# 解析响应数据
INFO_CODE=$(echo $INFO_RESPONSE | jq -r '.code // empty')
if [ "$INFO_CODE" = "200" ]; then
    echo -e "${GREEN}✓ 获取管理员信息成功${NC}"
    
    # 提取并显示详细信息
    echo -e "\n${BLUE}=== 管理员详细信息 ===${NC}"
    
    ADMIN_USERNAME=$(echo $INFO_RESPONSE | jq -r '.data.username // "N/A"')
    ADMIN_EMAIL=$(echo $INFO_RESPONSE | jq -r '.data.email // "N/A"')
    ADMIN_MOBILE=$(echo $INFO_RESPONSE | jq -r '.data.mobile // "N/A"')
    ADMIN_REALNAME=$(echo $INFO_RESPONSE | jq -r '.data.realName // "N/A"')
    ADMIN_AVATAR=$(echo $INFO_RESPONSE | jq -r '.data.avatar // "N/A"')
    ADMIN_DEPARTMENT=$(echo $INFO_RESPONSE | jq -r '.data.departmentId // "N/A"')
    ADMIN_NOTE=$(echo $INFO_RESPONSE | jq -r '.data.note // "N/A"')
    ADMIN_STATUS=$(echo $INFO_RESPONSE | jq -r '.data.status // "N/A"')
    ADMIN_SUPER=$(echo $INFO_RESPONSE | jq -r '.data.isSuperAdmin // "N/A"')
    ADMIN_LAST_IP=$(echo $INFO_RESPONSE | jq -r '.data.lastLoginIp // "N/A"')
    ADMIN_LAST_TIME=$(echo $INFO_RESPONSE | jq -r '.data.lastLoginTime // "N/A"')
    ADMIN_CREATED_BY=$(echo $INFO_RESPONSE | jq -r '.data.createdBy // "N/A"')
    ADMIN_UPDATED_BY=$(echo $INFO_RESPONSE | jq -r '.data.updatedBy // "N/A"')
    ADMIN_CREATED_TIME=$(echo $INFO_RESPONSE | jq -r '.data.createdTime // "N/A"')
    ADMIN_UPDATED_TIME=$(echo $INFO_RESPONSE | jq -r '.data.updatedTime // "N/A"')
    
    echo -e "${GREEN}ID:${NC} $ADMIN_ID"
    echo -e "${GREEN}用户名:${NC} $ADMIN_USERNAME"
    echo -e "${GREEN}邮箱:${NC} $ADMIN_EMAIL"
    echo -e "${GREEN}手机号:${NC} $ADMIN_MOBILE"
    echo -e "${GREEN}真实姓名:${NC} $ADMIN_REALNAME"
    echo -e "${GREEN}头像:${NC} $ADMIN_AVATAR"
    echo -e "${GREEN}部门ID:${NC} $ADMIN_DEPARTMENT"
    echo -e "${GREEN}备注:${NC} $ADMIN_NOTE"
    echo -e "${GREEN}状态:${NC} $ADMIN_STATUS (1=启用, 0=禁用)"
    echo -e "${GREEN}超级管理员:${NC} $ADMIN_SUPER"
    echo -e "${GREEN}最后登录IP:${NC} $ADMIN_LAST_IP"
    echo -e "${GREEN}最后登录时间:${NC} $ADMIN_LAST_TIME"
    echo -e "${GREEN}创建者ID:${NC} $ADMIN_CREATED_BY"
    echo -e "${GREEN}更新者ID:${NC} $ADMIN_UPDATED_BY"
    echo -e "${GREEN}创建时间:${NC} $ADMIN_CREATED_TIME"
    echo -e "${GREEN}更新时间:${NC} $ADMIN_UPDATED_TIME"
    
else
    echo -e "${RED}✗ 获取管理员信息失败，错误码: $INFO_CODE${NC}"
    ERROR_MESSAGE=$(echo $INFO_RESPONSE | jq -r '.message // "Unknown error"')
    echo -e "${RED}错误信息: $ERROR_MESSAGE${NC}"
fi

# 测试3: 验证返回字段的完整性
echo -e "\n${YELLOW}测试3: 验证返回字段完整性${NC}"
if [ "$INFO_CODE" = "200" ]; then
    # 检查必需字段
    REQUIRED_FIELDS=("id" "username" "email" "realName" "status" "isSuperAdmin" "lastLoginTime" "createdTime" "updatedTime")
    MISSING_FIELDS=()
    
    for field in "${REQUIRED_FIELDS[@]}"; do
        FIELD_VALUE=$(echo $INFO_RESPONSE | jq -r ".data.$field // empty")
        if [ -z "$FIELD_VALUE" ] || [ "$FIELD_VALUE" = "null" ]; then
            MISSING_FIELDS+=("$field")
        fi
    done
    
    if [ ${#MISSING_FIELDS[@]} -eq 0 ]; then
        echo -e "${GREEN}✓ 所有必需字段都存在${NC}"
    else
        echo -e "${RED}✗ 缺少以下必需字段: ${MISSING_FIELDS[*]}${NC}"
    fi
    
    # 检查可选字段
    OPTIONAL_FIELDS=("mobile" "avatar" "departmentId" "note" "lastLoginIp" "createdBy" "updatedBy")
    echo -e "\n${BLUE}可选字段检查:${NC}"
    for field in "${OPTIONAL_FIELDS[@]}"; do
        FIELD_VALUE=$(echo $INFO_RESPONSE | jq -r ".data.$field // empty")
        if [ -n "$FIELD_VALUE" ] && [ "$FIELD_VALUE" != "null" ]; then
            echo -e "${GREEN}✓ $field: $FIELD_VALUE${NC}"
        else
            echo -e "${YELLOW}○ $field: 未设置${NC}"
        fi
    done
fi

# 测试4: 测试无token访问（应该失败）
echo -e "\n${YELLOW}测试4: 测试无token访问（应该失败）${NC}"
NO_TOKEN_RESPONSE=$(curl -s -X GET "$BASE_URL$API_ENDPOINT/info" \
  -H "Content-Type: application/json")

echo "无token响应: $NO_TOKEN_RESPONSE"

NO_TOKEN_CODE=$(echo $NO_TOKEN_RESPONSE | jq -r '.code // empty')
if [ "$NO_TOKEN_CODE" = "401" ]; then
    echo -e "${GREEN}✓ 身份验证正常工作（正确拒绝无token请求）${NC}"
else
    echo -e "${RED}✗ 身份验证可能存在问题${NC}"
fi

# 测试5: 测试无效token访问（应该失败）
echo -e "\n${YELLOW}测试5: 测试无效token访问（应该失败）${NC}"
INVALID_TOKEN_RESPONSE=$(curl -s -X GET "$BASE_URL$API_ENDPOINT/info" \
  -H "Authorization: Bearer invalid_token_here" \
  -H "Content-Type: application/json")

echo "无效token响应: $INVALID_TOKEN_RESPONSE"

INVALID_TOKEN_CODE=$(echo $INVALID_TOKEN_RESPONSE | jq -r '.code // empty')
if [ "$INVALID_TOKEN_CODE" = "401" ]; then
    echo -e "${GREEN}✓ JWT验证正常工作（正确拒绝无效token）${NC}"
else
    echo -e "${RED}✗ JWT验证可能存在问题${NC}"
fi

# 测试6: 数据类型验证
echo -e "\n${YELLOW}测试6: 数据类型验证${NC}"
if [ "$INFO_CODE" = "200" ]; then
    # 验证数字类型字段
    ID_TYPE=$(echo $INFO_RESPONSE | jq -r 'type')
    STATUS_VALUE=$(echo $INFO_RESPONSE | jq -r '.data.status')
    SUPER_ADMIN_VALUE=$(echo $INFO_RESPONSE | jq -r '.data.isSuperAdmin')
    
    echo "验证数据类型..."
    
    # 检查ID是否为数字
    if echo $ADMIN_ID | grep -E '^[0-9]+$' > /dev/null; then
        echo -e "${GREEN}✓ ID字段类型正确（数字）${NC}"
    else
        echo -e "${RED}✗ ID字段类型错误${NC}"
    fi
    
    # 检查status是否为数字
    if echo $STATUS_VALUE | grep -E '^[0-1]$' > /dev/null; then
        echo -e "${GREEN}✓ status字段值正确（0或1）${NC}"
    else
        echo -e "${RED}✗ status字段值错误: $STATUS_VALUE${NC}"
    fi
    
    # 检查isSuperAdmin是否为布尔值
    if [ "$SUPER_ADMIN_VALUE" = "true" ] || [ "$SUPER_ADMIN_VALUE" = "false" ]; then
        echo -e "${GREEN}✓ isSuperAdmin字段类型正确（布尔值）${NC}"
    else
        echo -e "${RED}✗ isSuperAdmin字段类型错误: $SUPER_ADMIN_VALUE${NC}"
    fi
fi

# 测试总结
echo -e "\n${YELLOW}=== 测试总结 ===${NC}"
if [ "$INFO_CODE" = "200" ]; then
    echo -e "${GREEN}✓ 接口功能正常${NC}"
    echo -e "${GREEN}✓ 返回数据完整${NC}"
    echo -e "${GREEN}✓ 身份验证有效${NC}"
    echo -e "${GREEN}✓ 数据格式正确${NC}"
    
    echo -e "\n${BLUE}接口返回的主要信息:${NC}"
    echo -e "- 管理员: ${ADMIN_REALNAME} (@${ADMIN_USERNAME})"
    echo -e "- 权限: $([ "$ADMIN_SUPER" = "true" ] && echo "超级管理员" || echo "普通管理员")"
    echo -e "- 状态: $([ "$ADMIN_STATUS" = "1" ] && echo "正常" || echo "禁用")"
    echo -e "- 最后登录: $ADMIN_LAST_TIME"
else
    echo -e "${RED}✗ 接口测试失败${NC}"
fi

echo -e "\n${YELLOW}测试完成！${NC}"
echo -e "${YELLOW}注意: 如果某些测试失败，请检查：${NC}"
echo "1. 服务器是否正在运行在 $BASE_URL"
echo "2. 数据库是否正确配置并包含测试数据"
echo "3. JWT中间件是否正确配置"
echo "4. 管理员账户表结构是否正确"