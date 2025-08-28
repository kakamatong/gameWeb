#!/bin/bash

# 登入认证日志API测试脚本
# 用于测试管理后台的登录日志查询相关接口

# 配置信息
BASE_URL="http://localhost:8080"
ADMIN_USERNAME="admin"
ADMIN_PASSWORD="123456"

# 颜色输出配置
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 分隔线函数
print_separator() {
    echo "============================================================"
}

# 检查响应是否成功
check_response() {
    local response="$1"
    local description="$2"
    
    if echo "$response" | jq -e '.code == 200' > /dev/null 2>&1; then
        log_success "$description 成功"
        return 0
    else
        log_error "$description 失败"
        echo "响应内容: $response"
        return 1
    fi
}

# 管理员登录获取Token
admin_login() {
    log_info "正在进行管理员登录..."
    
    local login_response=$(curl -s -X POST \
        "$BASE_URL/api/admin/login" \
        -H "Content-Type: application/json" \
        -d "{
            \"username\": \"$ADMIN_USERNAME\",
            \"password\": \"$ADMIN_PASSWORD\"
        }")
    
    if check_response "$login_response" "管理员登录"; then
        TOKEN=$(echo "$login_response" | jq -r '.data.token')
        log_success "Token获取成功: ${TOKEN:0:20}..."
        return 0
    else
        log_error "管理员登录失败，无法继续测试"
        exit 1
    fi
}

# 测试获取用户登录日志
test_get_auth_logs() {
    print_separator
    log_info "测试1: 获取用户登录日志"
    
    # 测试1.1: 获取所有用户的登录日志（默认分页）
    log_info "1.1 获取所有用户的登录日志（默认分页）"
    local response=$(curl -s -X GET \
        "$BASE_URL/api/admin/logs/auth" \
        -H "Authorization: Bearer $TOKEN")
    
    check_response "$response" "获取所有用户登录日志"
    
    if echo "$response" | jq -e '.data.data | length > 0' > /dev/null 2>&1; then
        local total=$(echo "$response" | jq -r '.data.total')
        local count=$(echo "$response" | jq -r '.data.data | length')
        log_success "返回 $count 条记录，总计 $total 条"
        
        # 显示第一条记录的详细信息
        local first_log=$(echo "$response" | jq -r '.data.data[0]')
        log_info "第一条登录日志:"
        echo "$first_log" | jq '.'
    else
        log_warning "没有找到登录日志数据"
    fi
    
    # 测试1.2: 指定用户ID查询
    log_info "1.2 指定用户ID查询登录日志"
    local userid="12345"
    local response2=$(curl -s -X GET \
        "$BASE_URL/api/admin/logs/auth?userid=$userid&page=1&pageSize=5" \
        -H "Authorization: Bearer $TOKEN")
    
    check_response "$response2" "指定用户ID查询登录日志"
    
    # 测试1.3: 时间范围查询
    log_info "1.3 时间范围查询登录日志"
    local start_time="2024-01-01T00:00:00Z"
    local end_time="2024-12-31T23:59:59Z"
    local response3=$(curl -s -X GET \
        "$BASE_URL/api/admin/logs/auth?startTime=$start_time&endTime=$end_time&page=1&pageSize=3" \
        -H "Authorization: Bearer $TOKEN")
    
    check_response "$response3" "时间范围查询登录日志"
    
    # 测试1.4: 分页参数测试
    log_info "1.4 分页参数测试"
    local response4=$(curl -s -X GET \
        "$BASE_URL/api/admin/logs/auth?page=2&pageSize=10" \
        -H "Authorization: Bearer $TOKEN")
    
    check_response "$response4" "分页参数测试"
    
    # 测试1.5: 无效参数测试
    log_info "1.5 无效参数测试"
    local response5=$(curl -s -X GET \
        "$BASE_URL/api/admin/logs/auth?page=0&pageSize=200" \
        -H "Authorization: Bearer $TOKEN")
    
    # 这里应该会自动调整参数，依然返回成功
    check_response "$response5" "无效参数自动调整测试"
}

# 测试获取用户登录统计
test_get_login_stats() {
    print_separator
    log_info "测试2: 获取用户登录统计"
    
    # 测试2.1: 正常获取登录统计
    log_info "2.1 获取指定用户登录统计"
    local userid="12345"
    local response=$(curl -s -X GET \
        "$BASE_URL/api/admin/logs/login-stats?userid=$userid" \
        -H "Authorization: Bearer $TOKEN")
    
    if check_response "$response" "获取用户登录统计"; then
        log_info "登录统计详情:"
        echo "$response" | jq '.data'
    fi
    
    # 测试2.2: 缺少用户ID参数
    log_info "2.2 缺少用户ID参数测试"
    local response2=$(curl -s -X GET \
        "$BASE_URL/api/admin/logs/login-stats" \
        -H "Authorization: Bearer $TOKEN")
    
    if echo "$response2" | jq -e '.code == 400' > /dev/null 2>&1; then
        log_success "正确返回400错误：缺少用户ID参数"
    else
        log_error "应该返回400错误，但返回了: $response2"
    fi
    
    # 测试2.3: 无效用户ID
    log_info "2.3 无效用户ID测试"
    local response3=$(curl -s -X GET \
        "$BASE_URL/api/admin/logs/login-stats?userid=invalid" \
        -H "Authorization: Bearer $TOKEN")
    
    if echo "$response3" | jq -e '.code == 400' > /dev/null 2>&1; then
        log_success "正确返回400错误：无效用户ID"
    else
        log_error "应该返回400错误，但返回了: $response3"
    fi
}

# 测试未认证访问
test_unauthorized_access() {
    print_separator
    log_info "测试3: 未认证访问测试"
    
    log_info "3.1 无Token访问登录日志接口"
    local response=$(curl -s -X GET \
        "$BASE_URL/api/admin/logs/auth")
    
    if echo "$response" | jq -e '.code == 401' > /dev/null 2>&1; then
        log_success "正确返回401错误：未认证"
    else
        log_error "应该返回401错误，但返回了: $response"
    fi
    
    log_info "3.2 无效Token访问登录统计接口"
    local response2=$(curl -s -X GET \
        "$BASE_URL/api/admin/logs/login-stats?userid=12345" \
        -H "Authorization: Bearer invalid-token")
    
    if echo "$response2" | jq -e '.code == 401' > /dev/null 2>&1; then
        log_success "正确返回401错误：无效Token"
    else
        log_error "应该返回401错误，但返回了: $response2"
    fi
}

# 性能测试
test_performance() {
    print_separator
    log_info "测试4: 性能测试"
    
    log_info "4.1 大分页查询测试"
    local start_time=$(date +%s.%3N)
    local response=$(curl -s -X GET \
        "$BASE_URL/api/admin/logs/auth?page=1&pageSize=100" \
        -H "Authorization: Bearer $TOKEN")
    local end_time=$(date +%s.%3N)
    
    if check_response "$response" "大分页查询"; then
        local duration=$(echo "$end_time - $start_time" | bc)
        log_success "查询100条记录耗时: ${duration}s"
        
        if (( $(echo "$duration < 2.0" | bc -l) )); then
            log_success "性能良好：响应时间小于2秒"
        else
            log_warning "性能较慢：响应时间超过2秒"
        fi
    fi
}

# 数据验证测试
test_data_validation() {
    print_separator
    log_info "测试5: 数据格式验证"
    
    log_info "5.1 响应数据结构验证"
    local response=$(curl -s -X GET \
        "$BASE_URL/api/admin/logs/auth?page=1&pageSize=1" \
        -H "Authorization: Bearer $TOKEN")
    
    if check_response "$response" "数据结构验证"; then
        # 验证分页信息
        if echo "$response" | jq -e '.data.total' > /dev/null 2>&1 && \
           echo "$response" | jq -e '.data.page' > /dev/null 2>&1 && \
           echo "$response" | jq -e '.data.pageSize' > /dev/null 2>&1 && \
           echo "$response" | jq -e '.data.data' > /dev/null 2>&1; then
            log_success "分页数据结构正确"
        else
            log_error "分页数据结构不正确"
        fi
        
        # 验证日志数据字段
        if echo "$response" | jq -e '.data.data | length > 0' > /dev/null 2>&1; then
            local first_log=$(echo "$response" | jq '.data.data[0]')
            local required_fields=("id" "userid" "channel" "ip" "deviceId" "loginTime" "duration" "status")
            
            for field in "${required_fields[@]}"; do
                if echo "$first_log" | jq -e ".$field" > /dev/null 2>&1; then
                    log_success "字段 $field 存在"
                else
                    log_error "字段 $field 缺失"
                fi
            done
        fi
    fi
}

# 主函数
main() {
    print_separator
    log_info "开始登入认证日志API测试"
    log_info "测试目标: $BASE_URL"
    print_separator
    
    # 检查依赖
    if ! command -v jq &> /dev/null; then
        log_error "jq 命令未找到，请先安装 jq"
        exit 1
    fi
    
    if ! command -v curl &> /dev/null; then
        log_error "curl 命令未找到，请先安装 curl"
        exit 1
    fi
    
    # 执行测试
    admin_login
    test_get_auth_logs
    test_get_login_stats
    test_unauthorized_access
    test_performance
    test_data_validation
    
    print_separator
    log_success "登入认证日志API测试完成！"
    print_separator
}

# 运行测试
main "$@"