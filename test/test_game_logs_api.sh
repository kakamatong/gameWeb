#!/bin/bash

# 对局结果日志API测试脚本
# 用于测试管理后台的游戏对局日志查询相关接口

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

# 测试获取用户对局日志
test_get_game_logs() {
    print_separator
    log_info "测试1: 获取用户对局日志"
    
    # 测试1.1: 获取所有用户的对局日志（默认分页）
    log_info "1.1 获取所有用户的对局日志（默认分页）"
    local response=$(curl -s -X GET \
        "$BASE_URL/api/admin/logs/game" \
        -H "Authorization: Bearer $TOKEN")
    
    check_response "$response" "获取所有用户对局日志"
    
    if echo "$response" | jq -e '.data.data | length > 0' > /dev/null 2>&1; then
        local total=$(echo "$response" | jq -r '.data.total')
        local count=$(echo "$response" | jq -r '.data.data | length')
        log_success "返回 $count 条记录，总计 $total 条"
        
        # 显示第一条记录的详细信息
        local first_log=$(echo "$response" | jq -r '.data.data[0]')
        log_info "第一条对局日志:"
        echo "$first_log" | jq '.'
        
        # 分析对局结果分布
        local win_count=$(echo "$response" | jq '[.data.data[] | select(.result == 1)] | length')
        local lose_count=$(echo "$response" | jq '[.data.data[] | select(.result == 2)] | length')
        local draw_count=$(echo "$response" | jq '[.data.data[] | select(.result == 3)] | length')
        local escape_count=$(echo "$response" | jq '[.data.data[] | select(.result == 4)] | length')
        
        log_info "对局结果分布 - 胜利:$win_count, 失败:$lose_count, 平局:$draw_count, 逃跑:$escape_count"
    else
        log_warning "没有找到对局日志数据"
    fi
    
    # 测试1.2: 指定用户ID查询
    log_info "1.2 指定用户ID查询对局日志"
    local userid="12345"
    local response2=$(curl -s -X GET \
        "$BASE_URL/api/admin/logs/game?userid=$userid&page=1&pageSize=5" \
        -H "Authorization: Bearer $TOKEN")
    
    check_response "$response2" "指定用户ID查询对局日志"
    
    # 测试1.3: 时间范围查询
    log_info "1.3 时间范围查询对局日志"
    local start_time="2024-01-01T00:00:00Z"
    local end_time="2024-12-31T23:59:59Z"
    local response3=$(curl -s -X GET \
        "$BASE_URL/api/admin/logs/game?startTime=$start_time&endTime=$end_time&page=1&pageSize=3" \
        -H "Authorization: Bearer $TOKEN")
    
    check_response "$response3" "时间范围查询对局日志"
    
    # 测试1.4: 复合条件查询
    log_info "1.4 复合条件查询（用户ID + 时间范围）"
    local response4=$(curl -s -X GET \
        "$BASE_URL/api/admin/logs/game?userid=$userid&startTime=$start_time&endTime=$end_time&page=1&pageSize=10" \
        -H "Authorization: Bearer $TOKEN")
    
    check_response "$response4" "复合条件查询"
    
    # 测试1.5: 大分页查询
    log_info "1.5 大分页查询测试"
    local response5=$(curl -s -X GET \
        "$BASE_URL/api/admin/logs/game?page=1&pageSize=50" \
        -H "Authorization: Bearer $TOKEN")
    
    check_response "$response5" "大分页查询"
}

# 测试获取用户对局统计
test_get_game_stats() {
    print_separator
    log_info "测试2: 获取用户对局统计"
    
    # 测试2.1: 正常获取对局统计
    log_info "2.1 获取指定用户对局统计"
    local userid="12345"
    local response=$(curl -s -X GET \
        "$BASE_URL/api/admin/logs/game-stats?userid=$userid" \
        -H "Authorization: Bearer $TOKEN")
    
    if check_response "$response" "获取用户对局统计"; then
        log_info "对局统计详情:"
        local stats=$(echo "$response" | jq '.data')
        echo "$stats" | jq '.'
        
        # 解析统计数据
        local total_games=$(echo "$stats" | jq -r '.totalGames')
        local win_games=$(echo "$stats" | jq -r '.winGames')
        local win_rate=$(echo "$stats" | jq -r '.winRate')
        local net_profit=$(echo "$stats" | jq -r '.netProfit')
        local max_score=$(echo "$stats" | jq -r '.maxScore')
        
        log_info "统计摘要："
        log_info "  - 总对局: $total_games 局"
        log_info "  - 胜利: $win_games 局"
        log_info "  - 胜率: $win_rate"
        log_info "  - 净盈利: $net_profit"
        log_info "  - 最高得分: $max_score"
    fi
    
    # 测试2.2: 缺少用户ID参数
    log_info "2.2 缺少用户ID参数测试"
    local response2=$(curl -s -X GET \
        "$BASE_URL/api/admin/logs/game-stats" \
        -H "Authorization: Bearer $TOKEN")
    
    if echo "$response2" | jq -e '.code == 400' > /dev/null 2>&1; then
        log_success "正确返回400错误：缺少用户ID参数"
    else
        log_error "应该返回400错误，但返回了: $response2"
    fi
    
    # 测试2.3: 无效用户ID
    log_info "2.3 无效用户ID测试"
    local response3=$(curl -s -X GET \
        "$BASE_URL/api/admin/logs/game-stats?userid=invalid" \
        -H "Authorization: Bearer $TOKEN")
    
    if echo "$response3" | jq -e '.code == 400' > /dev/null 2>&1; then
        log_success "正确返回400错误：无效用户ID"
    else
        log_error "应该返回400错误，但返回了: $response3"
    fi
    
    # 测试2.4: 不存在的用户ID
    log_info "2.4 不存在的用户ID测试"
    local response4=$(curl -s -X GET \
        "$BASE_URL/api/admin/logs/game-stats?userid=999999999" \
        -H "Authorization: Bearer $TOKEN")
    
    if check_response "$response4" "不存在用户的对局统计"; then
        local stats=$(echo "$response4" | jq '.data')
        local total_games=$(echo "$stats" | jq -r '.totalGames')
        
        if [ "$total_games" = "0" ]; then
            log_success "正确返回空统计数据：总对局次数为0"
        else
            log_warning "不存在的用户返回了非零统计数据"
        fi
    fi
}

# 测试数据一致性
test_data_consistency() {
    print_separator
    log_info "测试3: 数据一致性验证"
    
    local userid="12345"
    
    # 获取对局日志
    log_info "3.1 获取对局日志数据"
    local logs_response=$(curl -s -X GET \
        "$BASE_URL/api/admin/logs/game?userid=$userid&page=1&pageSize=100" \
        -H "Authorization: Bearer $TOKEN")
    
    # 获取对局统计
    log_info "3.2 获取对局统计数据"
    local stats_response=$(curl -s -X GET \
        "$BASE_URL/api/admin/logs/game-stats?userid=$userid" \
        -H "Authorization: Bearer $TOKEN")
    
    if check_response "$logs_response" "获取对局日志" && check_response "$stats_response" "获取对局统计"; then
        local logs_total=$(echo "$logs_response" | jq -r '.data.total')
        local stats_total=$(echo "$stats_response" | jq -r '.data.totalGames')
        
        log_info "数据一致性检查："
        log_info "  - 日志总数: $logs_total"
        log_info "  - 统计总数: $stats_total"
        
        if [ "$logs_total" = "$stats_total" ]; then
            log_success "数据一致性验证通过：日志总数与统计总数一致"
        else
            log_warning "数据一致性问题：日志总数($logs_total) != 统计总数($stats_total)"
        fi
        
        # 验证胜利次数
        if echo "$logs_response" | jq -e '.data.data | length > 0' > /dev/null 2>&1; then
            local logs_win_count=$(echo "$logs_response" | jq '[.data.data[] | select(.result == 1)] | length')
            local stats_win_count=$(echo "$stats_response" | jq -r '.data.winGames')
            
            log_info "  - 日志胜利次数: $logs_win_count"
            log_info "  - 统计胜利次数: $stats_win_count"
            
            if [ "$logs_win_count" = "$stats_win_count" ]; then
                log_success "胜利次数一致性验证通过"
            else
                log_warning "胜利次数不一致：日志($logs_win_count) != 统计($stats_win_count)"
            fi
        fi
    fi
}

# 测试未认证访问
test_unauthorized_access() {
    print_separator
    log_info "测试4: 未认证访问测试"
    
    log_info "4.1 无Token访问对局日志接口"
    local response=$(curl -s -X GET \
        "$BASE_URL/api/admin/logs/game")
    
    if echo "$response" | jq -e '.code == 401' > /dev/null 2>&1; then
        log_success "正确返回401错误：未认证"
    else
        log_error "应该返回401错误，但返回了: $response"
    fi
    
    log_info "4.2 无效Token访问对局统计接口"
    local response2=$(curl -s -X GET \
        "$BASE_URL/api/admin/logs/game-stats?userid=12345" \
        -H "Authorization: Bearer invalid-token")
    
    if echo "$response2" | jq -e '.code == 401' > /dev/null 2>&1; then
        log_success "正确返回401错误：无效Token"
    else
        log_error "应该返回401错误，但返回了: $response2"
    fi
    
    log_info "4.3 过期Token测试"
    # 这里使用一个明显过期的JWT token
    local expired_token="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
    local response3=$(curl -s -X GET \
        "$BASE_URL/api/admin/logs/game-stats?userid=12345" \
        -H "Authorization: Bearer $expired_token")
    
    if echo "$response3" | jq -e '.code == 401' > /dev/null 2>&1; then
        log_success "正确返回401错误：过期Token"
    else
        log_error "应该返回401错误，但返回了: $response3"
    fi
}

# 性能测试
test_performance() {
    print_separator
    log_info "测试5: 性能测试"
    
    # 测试5.1: 大分页查询性能
    log_info "5.1 大分页查询性能测试"
    local start_time=$(date +%s.%3N)
    local response=$(curl -s -X GET \
        "$BASE_URL/api/admin/logs/game?page=1&pageSize=100" \
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
    
    # 测试5.2: 统计查询性能
    log_info "5.2 统计查询性能测试"
    local start_time=$(date +%s.%3N)
    local response2=$(curl -s -X GET \
        "$BASE_URL/api/admin/logs/game-stats?userid=12345" \
        -H "Authorization: Bearer $TOKEN")
    local end_time=$(date +%s.%3N)
    
    if check_response "$response2" "统计查询"; then
        local duration=$(echo "$end_time - $start_time" | bc)
        log_success "统计查询耗时: ${duration}s"
        
        if (( $(echo "$duration < 1.0" | bc -l) )); then
            log_success "统计查询性能良好：响应时间小于1秒"
        else
            log_warning "统计查询性能较慢：响应时间超过1秒"
        fi
    fi
    
    # 测试5.3: 并发查询测试
    log_info "5.3 并发查询测试（5个并发请求）"
    local start_time=$(date +%s.%3N)
    
    # 启动5个并发请求
    for i in {1..5}; do
        curl -s -X GET \
            "$BASE_URL/api/admin/logs/game?page=$i&pageSize=20" \
            -H "Authorization: Bearer $TOKEN" > /tmp/concurrent_test_$i.json &
    done
    
    # 等待所有请求完成
    wait
    
    local end_time=$(date +%s.%3N)
    local duration=$(echo "$end_time - $start_time" | bc)
    log_success "5个并发请求耗时: ${duration}s"
    
    # 检查并发请求结果
    local success_count=0
    for i in {1..5}; do
        if [ -f /tmp/concurrent_test_$i.json ]; then
            if jq -e '.code == 200' /tmp/concurrent_test_$i.json > /dev/null 2>&1; then
                ((success_count++))
            fi
            rm -f /tmp/concurrent_test_$i.json
        fi
    done
    
    log_info "并发测试结果：$success_count/5 个请求成功"
    
    if [ $success_count -eq 5 ]; then
        log_success "并发测试通过：所有请求都成功"
    else
        log_warning "并发测试部分失败：只有 $success_count 个请求成功"
    fi
}

# 数据验证测试
test_data_validation() {
    print_separator
    log_info "测试6: 数据格式验证"
    
    log_info "6.1 对局日志响应数据结构验证"
    local response=$(curl -s -X GET \
        "$BASE_URL/api/admin/logs/game?page=1&pageSize=1" \
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
        
        # 验证对局日志数据字段
        if echo "$response" | jq -e '.data.data | length > 0' > /dev/null 2>&1; then
            local first_log=$(echo "$response" | jq '.data.data[0]')
            local required_fields=("id" "userid" "gameid" "roomid" "gameMode" "result" "score" "winRiches" "loseRiches" "startTime" "endTime" "createTime")
            
            for field in "${required_fields[@]}"; do
                if echo "$first_log" | jq -e ".$field" > /dev/null 2>&1; then
                    log_success "字段 $field 存在"
                else
                    log_error "字段 $field 缺失"
                fi
            done
            
            # 验证数据类型
            local result=$(echo "$first_log" | jq -r '.result')
            if [[ "$result" =~ ^[0-4]$ ]]; then
                log_success "result字段值有效：$result"
            else
                log_warning "result字段值异常：$result"
            fi
        fi
    fi
    
    log_info "6.2 对局统计响应数据结构验证"
    local stats_response=$(curl -s -X GET \
        "$BASE_URL/api/admin/logs/game-stats?userid=12345" \
        -H "Authorization: Bearer $TOKEN")
    
    if check_response "$stats_response" "统计数据结构验证"; then
        local stats=$(echo "$stats_response" | jq '.data')
        local required_stats=("totalGames" "winGames" "winRate" "totalWinRiches" "totalLoseRiches" "netProfit" "maxScore" "todayGames")
        
        for field in "${required_stats[@]}"; do
            if echo "$stats" | jq -e ".$field" > /dev/null 2>&1; then
                log_success "统计字段 $field 存在"
            else
                log_error "统计字段 $field 缺失"
            fi
        done
        
        # 验证胜率格式
        local win_rate=$(echo "$stats" | jq -r '.winRate')
        if [[ "$win_rate" =~ ^[0-9]+\.[0-9]{2}%$ ]]; then
            log_success "胜率格式正确：$win_rate"
        else
            log_warning "胜率格式异常：$win_rate"
        fi
    fi
}

# 主函数
main() {
    print_separator
    log_info "开始对局结果日志API测试"
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
    
    if ! command -v bc &> /dev/null; then
        log_error "bc 命令未找到，请先安装 bc"
        exit 1
    fi
    
    # 执行测试
    admin_login
    test_get_game_logs
    test_get_game_stats
    test_data_consistency
    test_unauthorized_access
    test_performance
    test_data_validation
    
    print_separator
    log_success "对局结果日志API测试完成！"
    print_separator
}

# 运行测试
main "$@"