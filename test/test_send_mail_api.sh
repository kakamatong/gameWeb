#!/bin/bash

# 邮件发送接口测试脚本
# 测试 SendSystemMail 接口的各种场景

# 配置信息
BASE_URL="http://localhost:8080"
ADMIN_TOKEN="your-admin-jwt-token"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# 显示测试标题
print_test_header() {
    echo -e "\n${BLUE}==================== $1 ====================${NC}"
}

# 发送请求并检查响应
send_request() {
    local test_name="$1"
    local json_data="$2"
    local expected_code="$3"
    
    print_status "测试: $test_name"
    
    response=$(curl -s -w "\n%{http_code}" \
        -X POST "$BASE_URL/api/admin/mails/send" \
        -H "Authorization: Bearer $ADMIN_TOKEN" \
        -H "Content-Type: application/json" \
        -d "$json_data")
    
    # 分离响应体和状态码
    http_code=$(echo "$response" | tail -n1)
    response_body=$(echo "$response" | sed '$d')
    
    echo "HTTP状态码: $http_code"
    echo "响应内容: $response_body"
    
    # 检查状态码
    if [ "$http_code" == "$expected_code" ]; then
        print_success "状态码匹配 ($expected_code)"
    else
        print_error "状态码不匹配 (期望: $expected_code, 实际: $http_code)"
    fi
    
    # 解析JSON响应
    if command -v jq >/dev/null 2>&1; then
        echo "$response_body" | jq '.'
    else
        print_warning "未安装jq，无法格式化JSON输出"
    fi
    
    echo ""
    return 0
}

# 主测试函数
main() {
    print_test_header "邮件发送接口测试"
    
    print_status "开始测试 SendSystemMail 接口"
    print_status "Base URL: $BASE_URL"
    print_warning "请确保已设置正确的管理员Token: $ADMIN_TOKEN"
    
    # ==================== 正常场景测试 ====================
    
    print_test_header "1. 全服邮件发送测试"
    
    # 测试1: 发送全服邮件（带奖励）
    send_request "发送全服邮件（带奖励）" '{
        "type": 0,
        "title": "系统维护补偿邮件",
        "content": "感谢您的耐心等待，系统维护已完成，请查收补偿奖励。",
        "awards": "{\"props\":[{\"id\":1,\"cnt\":1000},{\"id\":2,\"cnt\":50}]}",
        "startTime": "2024-08-28T00:00:00Z",
        "endTime": "2024-09-04T23:59:59Z"
    }' "200"
    
    # 测试2: 发送全服邮件（无奖励）
    send_request "发送全服邮件（无奖励）" '{
        "type": 0,
        "title": "系统公告",
        "content": "游戏将于明天进行更新维护，预计维护时间2小时。",
        "awards": "",
        "startTime": "2024-08-28T10:00:00Z",
        "endTime": "2024-08-30T23:59:59Z"
    }' "200"
    
    print_test_header "2. 个人邮件发送测试"
    
    # 测试3: 发送个人邮件
    send_request "发送个人邮件" '{
        "type": 1,
        "title": "特殊奖励",
        "content": "恭喜您获得特殊成就，请查收奖励！",
        "awards": "{\"props\":[{\"id\":3,\"cnt\":500}]}",
        "startTime": "2024-08-28T12:00:00Z",
        "endTime": "2024-09-28T23:59:59Z",
        "targetUsers": [12345, 67890, 11111]
    }' "200"
    
    # ==================== 异常场景测试 ====================
    
    print_test_header "3. 参数验证测试"
    
    # 测试4: 邮件类型错误
    send_request "邮件类型错误" '{
        "type": 2,
        "title": "测试邮件",
        "content": "测试内容",
        "startTime": "2024-08-28T00:00:00Z",
        "endTime": "2024-09-04T23:59:59Z"
    }' "400"
    
    # 测试5: 个人邮件缺少目标用户
    send_request "个人邮件缺少目标用户" '{
        "type": 1,
        "title": "测试个人邮件",
        "content": "测试内容",
        "startTime": "2024-08-28T00:00:00Z",
        "endTime": "2024-09-04T23:59:59Z"
    }' "400"
    
    # 测试6: 时间范围错误
    send_request "时间范围错误" '{
        "type": 0,
        "title": "测试邮件",
        "content": "测试内容",
        "startTime": "2024-09-04T00:00:00Z",
        "endTime": "2024-08-28T23:59:59Z"
    }' "400"
    
    # 测试7: 标题为空
    send_request "标题为空" '{
        "type": 0,
        "title": "",
        "content": "测试内容",
        "startTime": "2024-08-28T00:00:00Z",
        "endTime": "2024-09-04T23:59:59Z"
    }' "400"
    
    # 测试8: 内容为空
    send_request "内容为空" '{
        "type": 0,
        "title": "测试邮件",
        "content": "",
        "startTime": "2024-08-28T00:00:00Z",
        "endTime": "2024-09-04T23:59:59Z"
    }' "400"
    
    # 测试9: 奖励格式错误
    send_request "奖励格式错误" '{
        "type": 0,
        "title": "测试邮件",
        "content": "测试内容",
        "awards": "invalid json",
        "startTime": "2024-08-28T00:00:00Z",
        "endTime": "2024-09-04T23:59:59Z"
    }' "400"
    
    # 测试10: 奖励道具ID无效
    send_request "奖励道具ID无效" '{
        "type": 0,
        "title": "测试邮件",
        "content": "测试内容",
        "awards": "{\"props\":[{\"id\":0,\"cnt\":100}]}",
        "startTime": "2024-08-28T00:00:00Z",
        "endTime": "2024-09-04T23:59:59Z"
    }' "400"
    
    # 测试11: 个人邮件目标用户过多
    send_request "个人邮件目标用户过多" '{
        "type": 1,
        "title": "测试邮件",
        "content": "测试内容",
        "startTime": "2024-08-28T00:00:00Z",
        "endTime": "2024-09-04T23:59:59Z",
        "targetUsers": ['$(for i in {1..1001}; do echo -n "$i,"; done | sed 's/,$//')']
    }' "400"
    
    print_test_header "4. 认证测试"
    
    # 测试12: 无Token
    print_status "测试: 无Token访问"
    response=$(curl -s -w "\n%{http_code}" \
        -X POST "$BASE_URL/api/admin/mails/send" \
        -H "Content-Type: application/json" \
        -d '{"type": 0, "title": "test", "content": "test", "startTime": "2024-08-28T00:00:00Z", "endTime": "2024-09-04T23:59:59Z"}')
    
    http_code=$(echo "$response" | tail -n1)
    echo "HTTP状态码: $http_code"
    if [ "$http_code" == "401" ]; then
        print_success "正确拒绝无Token访问"
    else
        print_error "应该返回401状态码"
    fi
    echo ""
    
    # 测试13: 无效Token
    print_status "测试: 无效Token"
    response=$(curl -s -w "\n%{http_code}" \
        -X POST "$BASE_URL/api/admin/mails/send" \
        -H "Authorization: Bearer invalid-token" \
        -H "Content-Type: application/json" \
        -d '{"type": 0, "title": "test", "content": "test", "startTime": "2024-08-28T00:00:00Z", "endTime": "2024-09-04T23:59:59Z"}')
    
    http_code=$(echo "$response" | tail -n1)
    echo "HTTP状态码: $http_code"
    if [ "$http_code" == "401" ]; then
        print_success "正确拒绝无效Token"
    else
        print_error "应该返回401状态码"
    fi
    echo ""
    
    print_test_header "测试完成"
    print_success "所有邮件发送接口测试已完成"
    print_warning "请检查以上测试结果，确保接口按预期工作"
    
    echo -e "\n${BLUE}测试说明:${NC}"
    echo "1. 成功的邮件发送应返回HTTP 200和包含mailId的响应"
    echo "2. 参数错误应返回HTTP 400和相应的错误信息"
    echo "3. 认证失败应返回HTTP 401"
    echo "4. 全服邮件会在用户登录时自动同步"
    echo "5. 个人邮件会立即发送给指定用户"
    echo ""
}

# 检查依赖
check_dependencies() {
    if ! command -v curl >/dev/null 2>&1; then
        print_error "curl 未安装，请先安装 curl"
        exit 1
    fi
    
    if ! command -v jq >/dev/null 2>&1; then
        print_warning "jq 未安装，JSON输出将不会格式化"
        print_warning "建议安装jq: sudo apt-get install jq (Ubuntu/Debian) 或 brew install jq (macOS)"
    fi
}

# 脚本入口
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    check_dependencies
    main "$@"
fi