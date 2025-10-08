#!/bin/bash

################################################################################
# 测试框架核心 - HTTP 封装、断言、报告
################################################################################

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m' # No Color

# 全局变量
CURRENT_SUITE=""
CURRENT_TEST_CASE=""
TEST_CASE_NUMBER=0
SUITE_TESTS_TOTAL=0
SUITE_TESTS_PASSED=0
SUITE_TESTS_FAILED=0
SUITE_START_TIME=0

GLOBAL_TESTS_TOTAL=0
GLOBAL_TESTS_PASSED=0
GLOBAL_TESTS_FAILED=0
GLOBAL_START_TIME=0

# 重试配置
MAX_RETRIES=3
RETRY_DELAY=1

# 日志文件路径（由主脚本设置）
LOG_DETAILED="${LOG_DETAILED:-}"
LOG_API_CALLS="${LOG_API_CALLS:-}"
LOG_FAILURES="${LOG_FAILURES:-}"

################################################################################
# 时间工具
################################################################################

# 获取当前时间戳（毫秒）
get_timestamp_ms() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        python3 -c 'import time; print(int(time.time() * 1000))'
    else
        # Linux
        date +%s%3N
    fi
}

# 计算时间差（毫秒）
time_diff_ms() {
    local start=$1
    local end=$2
    echo $((end - start))
}

################################################################################
# 日志记录
################################################################################

log_info() {
    local msg="$1"
    echo -e "${BLUE}[INFO]${NC} $msg" | tee -a "$LOG_DETAILED"
}

log_success() {
    local msg="$1"
    echo -e "${GREEN}[✓]${NC} $msg" | tee -a "$LOG_DETAILED"
}

log_error() {
    local msg="$1"
    echo -e "${RED}[✗]${NC} $msg" | tee -a "$LOG_DETAILED"
    [ -n "$LOG_FAILURES" ] && echo -e "[$(date '+%Y-%m-%d %H:%M:%S')] $msg" >> "$LOG_FAILURES"
}

log_warning() {
    local msg="$1"
    echo -e "${YELLOW}[!]${NC} $msg" | tee -a "$LOG_DETAILED"
}

log_debug() {
    local msg="$1"
    if [ "$VERBOSE" = "true" ]; then
        echo -e "${CYAN}[DEBUG]${NC} $msg" | tee -a "$LOG_DETAILED"
    else
        echo "[DEBUG] $msg" >> "$LOG_DETAILED"
    fi
}

################################################################################
# 测试套件管理
################################################################################

# 开始测试套件
start_test_suite() {
    local suite_name="$1"
    CURRENT_SUITE="$suite_name"
    TEST_CASE_NUMBER=0
    SUITE_TESTS_TOTAL=0
    SUITE_TESTS_PASSED=0
    SUITE_TESTS_FAILED=0
    SUITE_START_TIME=$(get_timestamp_ms)
    
    echo "" | tee -a "$LOG_DETAILED"
    echo -e "${BLUE}========================================${NC}" | tee -a "$LOG_DETAILED"
    echo -e "${BLUE}测试套件: $suite_name${NC}" | tee -a "$LOG_DETAILED"
    echo -e "${BLUE}========================================${NC}" | tee -a "$LOG_DETAILED"
    echo "" | tee -a "$LOG_DETAILED"
}

# 结束测试套件
end_test_suite() {
    local suite_end_time=$(get_timestamp_ms)
    local duration=$(time_diff_ms $SUITE_START_TIME $suite_end_time)
    
    echo "" | tee -a "$LOG_DETAILED"
    echo -e "${CYAN}----------------------------------------${NC}" | tee -a "$LOG_DETAILED"
    echo -e "套件: ${CYAN}$CURRENT_SUITE${NC}" | tee -a "$LOG_DETAILED"
    echo -e "总计: $SUITE_TESTS_TOTAL 个测试" | tee -a "$LOG_DETAILED"
    echo -e "${GREEN}通过: $SUITE_TESTS_PASSED${NC}" | tee -a "$LOG_DETAILED"
    echo -e "${RED}失败: $SUITE_TESTS_FAILED${NC}" | tee -a "$LOG_DETAILED"
    echo -e "用时: ${duration}ms" | tee -a "$LOG_DETAILED"
    echo -e "${CYAN}----------------------------------------${NC}" | tee -a "$LOG_DETAILED"
    echo "" | tee -a "$LOG_DETAILED"
    
    # 更新全局统计
    GLOBAL_TESTS_TOTAL=$((GLOBAL_TESTS_TOTAL + SUITE_TESTS_TOTAL))
    GLOBAL_TESTS_PASSED=$((GLOBAL_TESTS_PASSED + SUITE_TESTS_PASSED))
    GLOBAL_TESTS_FAILED=$((GLOBAL_TESTS_FAILED + SUITE_TESTS_FAILED))
}

################################################################################
# HTTP 请求封装
################################################################################

# HTTP 请求（带重试）
# 用法: http_request METHOD URL [DATA] [AUTH_REQUIRED]
http_request() {
    local method="$1"
    local url="$2"
    local data="${3:-}"
    local auth_required="${4:-true}"
    
    local full_url="${BASE_URL}${url}"
    local retry_count=0
    local response=""
    local http_code=""
    local start_time=$(get_timestamp_ms)
    
    while [ $retry_count -lt $MAX_RETRIES ]; do
        log_debug "Request: $method $full_url (attempt $((retry_count + 1)))"
        
        # 构建 curl 命令
        local curl_cmd="curl -s -w \"\\n%{http_code}\" -X $method \"$full_url\""
        curl_cmd="$curl_cmd -H \"Content-Type: application/json\""
        
        if [ "$auth_required" = "true" ] && [ -n "$AUTH_TOKEN" ]; then
            curl_cmd="$curl_cmd -H \"Authorization: Bearer $AUTH_TOKEN\""
        fi
        
        if [ -n "$data" ]; then
            curl_cmd="$curl_cmd -d '$data'"
        fi
        
        # 执行请求
        response=$(eval $curl_cmd 2>&1)
        
        if [ $? -eq 0 ]; then
            http_code=$(echo "$response" | tail -n1)
            local body=$(echo "$response" | sed '$d')
            
            local end_time=$(get_timestamp_ms)
            local duration=$(time_diff_ms $start_time $end_time)
            
            # 记录 API 调用
            if [ -n "$LOG_API_CALLS" ]; then
                echo "---" >> "$LOG_API_CALLS"
                echo "Time: $(date '+%Y-%m-%d %H:%M:%S')" >> "$LOG_API_CALLS"
                echo "Method: $method" >> "$LOG_API_CALLS"
                echo "URL: $full_url" >> "$LOG_API_CALLS"
                [ -n "$data" ] && echo "Data: $data" >> "$LOG_API_CALLS"
                echo "Status: $http_code" >> "$LOG_API_CALLS"
                echo "Duration: ${duration}ms" >> "$LOG_API_CALLS"
                echo "Response: $body" >> "$LOG_API_CALLS"
                echo "" >> "$LOG_API_CALLS"
            fi
            
            log_debug "Response: HTTP $http_code (${duration}ms)"
            
            # 设置全局变量供断言使用
            LAST_HTTP_CODE="$http_code"
            LAST_RESPONSE_BODY="$body"
            LAST_REQUEST_DURATION="$duration"
            
            return 0
        else
            retry_count=$((retry_count + 1))
            if [ $retry_count -lt $MAX_RETRIES ]; then
                log_warning "Request failed, retrying in ${RETRY_DELAY}s..."
                sleep $RETRY_DELAY
            fi
        fi
    done
    
    log_error "Request failed after $MAX_RETRIES attempts"
    LAST_HTTP_CODE="000"
    LAST_RESPONSE_BODY="Connection failed"
    LAST_REQUEST_DURATION="0"
    return 1
}

################################################################################
# 断言函数
################################################################################

# 开始测试用例
test_case() {
    local description="$1"
    TEST_CASE_NUMBER=$((TEST_CASE_NUMBER + 1))
    CURRENT_TEST_CASE="$description"
    SUITE_TESTS_TOTAL=$((SUITE_TESTS_TOTAL + 1))
    
    log_debug "Test case [$TEST_CASE_NUMBER]: $description"
}

# 断言 HTTP 状态码
assert_status() {
    local expected="$1"
    local description="${2:-Test case $TEST_CASE_NUMBER}"
    
    if [ "$LAST_HTTP_CODE" = "$expected" ]; then
        SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
        log_success "[$TEST_CASE_NUMBER] $description - HTTP $LAST_HTTP_CODE (${LAST_REQUEST_DURATION}ms)"
        return 0
    else
        SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
        log_error "[$TEST_CASE_NUMBER] $description - HTTP $LAST_HTTP_CODE (expected $expected)"
        log_error "  Response: $LAST_RESPONSE_BODY"
        
        if [ "$CONTINUE_ON_FAILURE" != "true" ]; then
            return 1
        fi
        return 1
    fi
}

# 断言成功响应 (2xx)
assert_success() {
    local description="${1:-Test case $TEST_CASE_NUMBER}"
    
    if [[ "$LAST_HTTP_CODE" =~ ^2[0-9]{2}$ ]]; then
        SUITE_TESTS_PASSED=$((SUITE_TESTS_PASSED + 1))
        log_success "[$TEST_CASE_NUMBER] $description - HTTP $LAST_HTTP_CODE (${LAST_REQUEST_DURATION}ms)" >&2
        return 0
    else
        SUITE_TESTS_FAILED=$((SUITE_TESTS_FAILED + 1))
        log_error "[$TEST_CASE_NUMBER] $description - HTTP $LAST_HTTP_CODE (expected 2xx)" >&2
        log_error "  Response: $LAST_RESPONSE_BODY" >&2
        
        if [ "$CONTINUE_ON_FAILURE" != "true" ]; then
            return 1
        fi
        return 1
    fi
}

# 断言响应字段存在
assert_field_exists() {
    local field_path="$1"
    local description="${2:-Field $field_path should exist}"
    local silent="${3:-false}"
    
    local value=$(echo "$LAST_RESPONSE_BODY" | jq -r "$field_path" 2>/dev/null)
    
    if [ -n "$value" ] && [ "$value" != "null" ]; then
        log_debug "Field $field_path exists: $value"
        return 0
    else
        if [ "$silent" != "true" ]; then
            log_error "Field $field_path does not exist or is null"
            log_error "  Response: $LAST_RESPONSE_BODY"
        fi
        return 1
    fi
}

# 断言响应字段值
assert_field_equals() {
    local field_path="$1"
    local expected="$2"
    local description="${3:-Field $field_path should equal $expected}"
    
    local actual=$(echo "$LAST_RESPONSE_BODY" | jq -r "$field_path" 2>/dev/null)
    
    if [ "$actual" = "$expected" ]; then
        log_debug "Field $field_path = $expected"
        return 0
    else
        log_error "Field $field_path mismatch"
        log_error "  Expected: $expected"
        log_error "  Actual: $actual"
        return 1
    fi
}

# 提取响应字段
extract_field() {
    local field_path="$1"
    echo "$LAST_RESPONSE_BODY" | jq -r "$field_path" 2>/dev/null
}

################################################################################
# 全局测试报告
################################################################################

# 开始全局测试
start_global_test() {
    GLOBAL_START_TIME=$(get_timestamp_ms)
    GLOBAL_TESTS_TOTAL=0
    GLOBAL_TESTS_PASSED=0
    GLOBAL_TESTS_FAILED=0
    
    echo -e "${MAGENTA}╔════════════════════════════════════════╗${NC}" | tee -a "$LOG_DETAILED"
    echo -e "${MAGENTA}║     TSU Admin API 全面测试框架     ║${NC}" | tee -a "$LOG_DETAILED"
    echo -e "${MAGENTA}╚════════════════════════════════════════╝${NC}" | tee -a "$LOG_DETAILED"
    echo "" | tee -a "$LOG_DETAILED"
    log_info "测试开始时间: $(date '+%Y-%m-%d %H:%M:%S')"
    log_info "API 地址: $BASE_URL"
    log_info "测试账号: $USERNAME"
    echo "" | tee -a "$LOG_DETAILED"
}

# 结束全局测试并生成报告
end_global_test() {
    local global_end_time=$(get_timestamp_ms)
    local total_duration=$(time_diff_ms $GLOBAL_START_TIME $global_end_time)
    local pass_rate=0
    
    if [ $GLOBAL_TESTS_TOTAL -gt 0 ]; then
        pass_rate=$((GLOBAL_TESTS_PASSED * 100 / GLOBAL_TESTS_TOTAL))
    fi
    
    echo "" | tee -a "$LOG_DETAILED"
    echo -e "${MAGENTA}╔════════════════════════════════════════╗${NC}" | tee -a "$LOG_DETAILED"
    echo -e "${MAGENTA}║           测试总结报告             ║${NC}" | tee -a "$LOG_DETAILED"
    echo -e "${MAGENTA}╚════════════════════════════════════════╝${NC}" | tee -a "$LOG_DETAILED"
    echo "" | tee -a "$LOG_DETAILED"
    echo "测试完成时间: $(date '+%Y-%m-%d %H:%M:%S')" | tee -a "$LOG_DETAILED"
    echo "总用时: ${total_duration}ms ($((total_duration / 1000))s)" | tee -a "$LOG_DETAILED"
    echo "" | tee -a "$LOG_DETAILED"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" | tee -a "$LOG_DETAILED"
    echo "  测试统计" | tee -a "$LOG_DETAILED"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" | tee -a "$LOG_DETAILED"
    echo "  总测试数:   $GLOBAL_TESTS_TOTAL" | tee -a "$LOG_DETAILED"
    echo -e "  ${GREEN}通过:       $GLOBAL_TESTS_PASSED${NC}" | tee -a "$LOG_DETAILED"
    echo -e "  ${RED}失败:       $GLOBAL_TESTS_FAILED${NC}" | tee -a "$LOG_DETAILED"
    echo "  通过率:     ${pass_rate}%" | tee -a "$LOG_DETAILED"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" | tee -a "$LOG_DETAILED"
    echo "" | tee -a "$LOG_DETAILED"
    
    if [ $GLOBAL_TESTS_FAILED -gt 0 ]; then
        echo -e "${RED}⚠ 存在失败的测试用例，请检查日志${NC}" | tee -a "$LOG_DETAILED"
        echo "失败日志: $LOG_FAILURES" | tee -a "$LOG_DETAILED"
        return 1
    else
        echo -e "${GREEN}✓ 所有测试通过！${NC}" | tee -a "$LOG_DETAILED"
        return 0
    fi
}
