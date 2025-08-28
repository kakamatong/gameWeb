#!/bin/bash

# 综合日志API测试脚本
BASE_URL="http://localhost:8080"
ADMIN_USERNAME="admin"
ADMIN_PASSWORD="123456"

# 颜色配置
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
print_header() { echo -e "\n${BLUE}=== $1 ===${NC}"; }

TOTAL_TESTS=0
PASSED_TESTS=0

check_response() {
    local response="$1"
    local description="$2"
    ((TOTAL_TESTS++))
    
    if echo "$response" | jq -e '.code == 200' > /dev/null 2>&1; then
        log_success "$description"
        ((PASSED_TESTS++))
        return 0
    else
        log_error "$description"
        return 1
    fi
}

# 管理员登录
admin_login() {
    print_header "管理员登录"
    local response=$(curl -s -X POST "$BASE_URL/api/admin/login" \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"$ADMIN_USERNAME\",\"password\":\"$ADMIN_PASSWORD\"}")
    
    if check_response "$response" "管理员登录"; then
        TOKEN=$(echo "$response" | jq -r '.data.token')
        log_success "Token获取成功"
    else
        log_error "登录失败，退出测试"
        exit 1
    fi
}

# 测试登录日志API
test_auth_logs() {
    print_header "登入认证日志测试"
    
    # 基础查询
    local response=$(curl -s -X GET "$BASE_URL/api/admin/logs/auth?page=1&pageSize=5" \
        -H "Authorization: Bearer $TOKEN")
    check_response "$response" "获取登入认证日志"
    
    # 用户过滤
    local response2=$(curl -s -X GET "$BASE_URL/api/admin/logs/auth?userid=12345" \
        -H "Authorization: Bearer $TOKEN")
    check_response "$response2" "按用户ID过滤"
    
    # 时间过滤
    local response3=$(curl -s -X GET "$BASE_URL/api/admin/logs/auth?startTime=2024-01-01T00:00:00Z&endTime=2024-12-31T23:59:59Z" \
        -H "Authorization: Bearer $TOKEN")
    check_response "$response3" "按时间范围过滤"
}

# 测试对局日志API
test_game_logs() {
    print_header "对局结果日志测试"
    
    # 基础查询
    local response=$(curl -s -X GET "$BASE_URL/api/admin/logs/game?page=1&pageSize=5" \
        -H "Authorization: Bearer $TOKEN")
    check_response "$response" "获取对局结果日志"
    
    # 用户过滤
    local response2=$(curl -s -X GET "$BASE_URL/api/admin/logs/game?userid=12345" \
        -H "Authorization: Bearer $TOKEN")
    check_response "$response2" "按用户ID过滤"
    
    # 复合条件
    local response3=$(curl -s -X GET "$BASE_URL/api/admin/logs/game?userid=12345&startTime=2024-01-01T00:00:00Z" \
        -H "Authorization: Bearer $TOKEN")
    check_response "$response3" "复合条件查询"
}

# 测试统计API
test_stats() {
    print_header "统计API测试"
    
    # 登录统计
    local response=$(curl -s -X GET "$BASE_URL/api/admin/logs/login-stats?userid=12345" \
        -H "Authorization: Bearer $TOKEN")
    check_response "$response" "获取登录统计"
    
    # 对局统计
    local response2=$(curl -s -X GET "$BASE_URL/api/admin/logs/game-stats?userid=12345" \
        -H "Authorization: Bearer $TOKEN")
    check_response "$response2" "获取对局统计"
    
    # 错误参数测试
    local response3=$(curl -s -X GET "$BASE_URL/api/admin/logs/login-stats")
    if echo "$response3" | jq -e '.code == 401' > /dev/null 2>&1; then
        log_success "无认证访问正确返回401"
        ((PASSED_TESTS++))
    else
        log_error "无认证访问应返回401"
    fi
    ((TOTAL_TESTS++))
}

# 性能测试
test_performance() {
    print_header "性能测试"
    
    local start_time=$(date +%s.%3N)
    local response=$(curl -s -X GET "$BASE_URL/api/admin/logs/auth?page=1&pageSize=50" \
        -H "Authorization: Bearer $TOKEN")
    local end_time=$(date +%s.%3N)
    
    if check_response "$response" "大分页查询"; then
        local duration=$(echo "$end_time - $start_time" | bc)
        if (( $(echo "$duration < 2.0" | bc -l) )); then
            log_success "性能测试通过，耗时: ${duration}s"
            ((PASSED_TESTS++))
        else
            log_error "性能测试失败，耗时过长: ${duration}s"
        fi
        ((TOTAL_TESTS++))
    fi
}

# 生成报告
generate_report() {
    print_header "测试报告"
    local success_rate=0
    if [ $TOTAL_TESTS -gt 0 ]; then
        success_rate=$(echo "scale=2; $PASSED_TESTS * 100 / $TOTAL_TESTS" | bc)
    fi
    
    echo "总测试数: $TOTAL_TESTS"
    echo "通过测试: $PASSED_TESTS"
    echo "成功率: ${success_rate}%"
    
    if [ $PASSED_TESTS -eq $TOTAL_TESTS ]; then
        log_success "🎉 所有测试通过！"
    else
        log_error "❌ 部分测试失败"
    fi
}

# 主函数
main() {
    echo "开始日志API综合测试"
    echo "测试目标: $BASE_URL"
    
    # 检查依赖
    for cmd in jq curl bc; do
        if ! command -v $cmd &> /dev/null; then
            log_error "$cmd 未安装"
            exit 1
        fi
    done
    
    # 执行测试
    admin_login
    test_auth_logs
    test_game_logs
    test_stats
    test_performance
    generate_report
    
    # 返回结果
    [ $PASSED_TESTS -eq $TOTAL_TESTS ] && exit 0 || exit 1
}

main "$@"