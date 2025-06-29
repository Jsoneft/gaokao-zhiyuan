#!/bin/bash

# 高考志愿填报系统测试脚本
# 功能：1. 启动本地服务（如果未启动）2. 测试三个接口并打印curl命令日志

set -e  # 遇到错误立即退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
SERVER_PORT=${PORT:-8031}
SERVER_HOST="localhost"
BASE_URL="http://${SERVER_HOST}:${SERVER_PORT}"
SERVER_BINARY="./bin/gaokao-server"
LOG_DIR="./logs"
LOG_FILE="$LOG_DIR/test_$(date +%Y%m%d_%H%M%S).log"
PID_FILE="$LOG_DIR/server.pid"

# 确保日志目录存在
mkdir -p "$LOG_DIR"

# 日志函数
log() {
    echo -e "${BLUE}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $1" | tee -a "$LOG_FILE"
}

log_success() {
    echo -e "${GREEN}[$(date '+%Y-%m-%d %H:%M:%S')] ✅ $1${NC}" | tee -a "$LOG_FILE"
}

log_error() {
    echo -e "${RED}[$(date '+%Y-%m-%d %H:%M:%S')] ❌ $1${NC}" | tee -a "$LOG_FILE"
}

log_warning() {
    echo -e "${YELLOW}[$(date '+%Y-%m-%d %H:%M:%S')] ⚠️  $1${NC}" | tee -a "$LOG_FILE"
}

# 检查服务是否运行
check_server_running() {
    curl -s "$BASE_URL/api/health" > /dev/null 2>&1
    return $?
}

# 停止现有服务
stop_existing_server() {
    if [ -f "$PID_FILE" ]; then
        local old_pid=$(cat "$PID_FILE")
        if ps -p "$old_pid" > /dev/null 2>&1; then
            log "停止现有服务 (PID: $old_pid)..."
            kill "$old_pid"
            sleep 2
            # 强制杀死如果还在运行
            if ps -p "$old_pid" > /dev/null 2>&1; then
                log_warning "强制停止服务..."
                kill -9 "$old_pid"
            fi
            rm -f "$PID_FILE"
            log_success "现有服务已停止"
        else
            rm -f "$PID_FILE"
        fi
    fi
    
    # 检查端口是否被占用
    local port_pid=$(lsof -ti:$SERVER_PORT 2>/dev/null)
    if [ -n "$port_pid" ]; then
        log "端口 $SERVER_PORT 被进程 $port_pid 占用，正在停止..."
        kill "$port_pid" 2>/dev/null || kill -9 "$port_pid" 2>/dev/null
        sleep 1
    fi
}

# 启动服务
start_server() {
    log "停止现有服务并重新编译..."
    
    # 停止现有服务
    stop_existing_server
    
    # 重新编译
    log "重新编译服务器..."
    make clean && make build
    if [ $? -ne 0 ]; then
        log_error "编译失败"
        exit 1
    fi
    
    # 启动服务器
    log "启动服务器..."
    nohup "$SERVER_BINARY" > "$LOG_DIR/server.log" 2>&1 &
    SERVER_PID=$!
    
    # 等待服务器启动
    log "等待服务器启动..."
    for i in {1..30}; do
        if check_server_running; then
            log_success "服务器启动成功，PID: $SERVER_PID"
            echo "$SERVER_PID" > "$PID_FILE"
            return 0
        fi
        sleep 1
    done
    
    log_error "服务器启动失败"
    exit 1
}

# 测试接口函数
test_api() {
    local name="$1"
    local curl_cmd="$2"
    local expected_code="$3"
    
    echo "" | tee -a "$LOG_FILE"
    log "==================== 测试 $name ===================="
    
    # 打印可复制的curl命令
    echo -e "${YELLOW}📋 可复制的curl命令:${NC}" | tee -a "$LOG_FILE"
    echo "$curl_cmd" | tee -a "$LOG_FILE"
    echo "" | tee -a "$LOG_FILE"
    
    # 执行curl命令
    log "执行API请求..."
    
    # 获取响应和HTTP状态码
    response=$(eval "$curl_cmd" 2>/dev/null)
    http_code=$(eval "$curl_cmd -w '%{http_code}' -o /dev/null -s" 2>/dev/null)
    
    # 检查HTTP状态码
    if [ "$http_code" = "$expected_code" ]; then
        log_success "HTTP状态码: $http_code ✅"
    else
        log_error "HTTP状态码: $http_code (期望: $expected_code) ❌"
    fi
    
    # 打印响应内容
    log "响应内容:"
    # 使用jq格式化JSON并保持中文字符可读
    if command -v jq >/dev/null 2>&1; then
        echo "$response" | jq -r '.' 2>/dev/null | tee -a "$LOG_FILE" || echo "$response" | tee -a "$LOG_FILE"
    else
        # 如果没有jq，使用python但确保中文显示正确
        echo "$response" | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    print(json.dumps(data, ensure_ascii=False, indent=2))
except:
    print(sys.stdin.read())
" 2>/dev/null | tee -a "$LOG_FILE" || echo "$response" | tee -a "$LOG_FILE"
    fi
    
    # 检查响应是否包含预期的JSON结构
    if echo "$response" | grep -q '"code"'; then
        local response_code=$(echo "$response" | python3 -c "import sys, json; print(json.load(sys.stdin).get('code', -1))" 2>/dev/null || echo "-1")
        if [ "$response_code" = "0" ]; then
            log_success "API响应成功 (code: $response_code) ✅"
        else
            log_error "API响应错误 (code: $response_code) ❌"
        fi
    else
        log_warning "响应格式可能不是标准JSON"
    fi
}

# 主测试函数
run_tests() {
    log "开始API接口测试..."
    
    # 测试1: 健康检查接口
    test_api "健康检查接口" \
        "curl -s -X GET '$BASE_URL/api/health'" \
        "200"
    
    # 测试2: 获取排名接口
    test_api "获取排名接口" \
        "curl -s -X GET '$BASE_URL/api/rank/get?score=555'" \
        "200"
    
    # 测试3: 获取报表接口
    test_api "获取报表接口" \
        "curl -s -X GET '$BASE_URL/api/report/get?rank=50000&class_first_choise=物理&province=湖北&page=1&page_size=5'" \
        "200"
    
    # 额外测试: 高级位次查询接口 (POST)
    test_api "高级位次查询接口" \
        "curl -s -X POST '$BASE_URL/api/v1/query_rank' -H 'Content-Type: application/json' -d '{\"province\":\"湖北\",\"year\":2024,\"score\":555,\"subject_type\":\"物理\",\"class_demand\":[\"物\",\"化\",\"生\"]}'" \
        "200"
    
    # 测试major_min_rank_2024字段: 物理类专业
    test_api "测试物理类专业major_min_rank_2024字段" \
        "curl -s -X GET '$BASE_URL/api/report/get?rank=30000&class_first_choise=物理&province=湖北&page=1&page_size=3'" \
        "200"
    
    # 测试major_min_rank_2024字段: 历史类专业
    test_api "测试历史类专业major_min_rank_2024字段" \
        "curl -s -X GET '$BASE_URL/api/report/get?rank=15000&class_first_choise=历史&province=湖北&page=1&page_size=3'" \
        "200"
    
    # 验证物理类专业排名计算准确性
    test_api "验证物理类494分对应排名94438" \
        "curl -s -X GET '$BASE_URL/api/report/get?rank=50000&class_first_choise=物理&province=湖北&page=1&page_size=1'" \
        "200"
    
    # 验证历史类专业排名计算准确性  
    test_api "验证历史类488分对应排名25516" \
        "curl -s -X GET '$BASE_URL/api/report/get?rank=15000&class_first_choise=历史&province=湖北&page=1&page_size=1'" \
        "200"
    
    # 测试新增的fuzzy_subject_category参数 - 模糊查询包含"临床"的专业名称
    test_api "测试fuzzy_subject_category模糊查询临床类专业" \
        "curl -s -X GET '$BASE_URL/api/report/get?rank=18888&class_first_choise=物理&strategy=0&page=1&page_size=3&fuzzy_subject_category=临床'" \
        "200"
    
    # 测试fuzzy_subject_category参数 - 模糊查询包含"计算机"的专业名称
    test_api "测试fuzzy_subject_category模糊查询计算机类专业" \
        "curl -s -X GET '$BASE_URL/api/report/get?rank=18888&class_first_choise=物理&strategy=0&page=1&page_size=3&fuzzy_subject_category=计算机'" \
        "200"
    
    # 测试fuzzy_subject_category参数 - 模糊查询包含"电气"的专业名称
    test_api "测试fuzzy_subject_category模糊查询电气工程类专业" \
        "curl -s -X GET '$BASE_URL/api/report/get?rank=18888&class_first_choise=物理&strategy=0&page=1&page_size=3&fuzzy_subject_category=电气'" \
        "200"
    
    # 测试fuzzy_subject_category参数 - 模糊查询包含"工程"的专业名称
    test_api "测试fuzzy_subject_category模糊查询工程类专业" \
        "curl -s -X GET '$BASE_URL/api/report/get?rank=18888&class_first_choise=物理&strategy=0&page=1&page_size=3&fuzzy_subject_category=工程'" \
        "200"
    

}

# 生成测试报告
generate_report() {
    echo "" | tee -a "$LOG_FILE"
    log "==================== 测试报告 ===================="
    log "测试时间: $(date)"
    log "服务地址: $BASE_URL"
    log "日志文件: $LOG_FILE"
    log "服务日志: $LOG_DIR/server.log"
    
    if [ -f "$PID_FILE" ]; then
        local pid=$(cat "$PID_FILE")
        if ps -p "$pid" > /dev/null 2>&1; then
            log "服务状态: 运行中 (PID: $pid)"
        else
            log "服务状态: 已停止"
        fi
    else
        log "服务状态: 未记录"
    fi
    
    log_success "测试完成！详细日志请查看: $LOG_FILE"
}

# 清理函数
cleanup() {
    if [ -f "$PID_FILE" ]; then
        local pid=$(cat "$PID_FILE")
        if ps -p "$pid" > /dev/null 2>&1; then
            log "停止测试服务 (PID: $pid)..."
            kill "$pid" 2>/dev/null
            sleep 2
            # 强制杀死如果还在运行
            if ps -p "$pid" > /dev/null 2>&1; then
                kill -9 "$pid" 2>/dev/null
            fi
            rm -f "$PID_FILE"
            log_success "服务已停止，测试完成"
        else
            rm -f "$PID_FILE"
        fi
    fi
}

# 主函数
main() {
    log "==================== 高考志愿填报系统测试脚本 ===================="
    log "开始执行测试..."
    
    # 启动服务
    start_server
    
    # 等待服务完全启动
    sleep 2
    
    # 运行测试
    run_tests
    
    # 生成报告
    generate_report
    
    # 清理
    cleanup
}

# 捕获退出信号
trap cleanup EXIT

# 执行主函数
main "$@" 