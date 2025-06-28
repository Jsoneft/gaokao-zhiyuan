#!/bin/bash

# 高考志愿填报系统部署脚本
# 功能：1. 交叉编译项目 2. 上传到远程服务器 3. 停止旧服务 4. 测试 5. 重启服务

# 注意：不使用 set -e，因为某些命令失败是正常的（如停止不存在的服务）

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log() {
    echo -e "${BLUE}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[$(date '+%Y-%m-%d %H:%M:%S')] ✅ $1${NC}"
}

log_error() {
    echo -e "${RED}[$(date '+%Y-%m-%d %H:%M:%S')] ❌ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}[$(date '+%Y-%m-%d %H:%M:%S')] ⚠️  $1${NC}"
}

# 读取 .env 文件
if [ ! -f ".env" ]; then
    log_error ".env 文件不存在"
    exit 1
fi

log "读取 .env 配置文件..."
export $(grep -v '^#' .env | xargs)

# 验证必要的环境变量
if [ -z "$REMOTE_SERVER_IP" ] || [ -z "$REMOTE_SERVER_PORT" ] || [ -z "$REMOTE_SERVER_USER" ] || [ -z "$REMOTE_SERVER_USER_PASSWORD" ] || [ -z "$REMOTE_SERVER_PROJECT_PATH" ]; then
    log_error "缺少必要的远程服务器配置变量"
    exit 1
fi

# 配置变量
BINARY_NAME="gaokao-zhiyuan"
LOCAL_BINARY="./bin/${BINARY_NAME}"
REMOTE_BINARY="${REMOTE_SERVER_PROJECT_PATH}/${BINARY_NAME}"
REMOTE_TEST_SCRIPT="${REMOTE_SERVER_PROJECT_PATH}/test.sh"
REMOTE_ENV_FILE="${REMOTE_SERVER_PROJECT_PATH}/.env"

log "部署配置:"
log "  远程服务器: ${REMOTE_SERVER_IP}:${REMOTE_SERVER_PORT}"
log "  远程用户: ${REMOTE_SERVER_USER}"
log "  远程项目路径: ${REMOTE_SERVER_PROJECT_PATH}"
log "  二进制文件名: ${BINARY_NAME}"

# 1. 清理并交叉编译
log "开始交叉编译..."
make clean

log "编译 Linux 版本..."
GOOS=linux GOARCH=amd64 go build -o "${LOCAL_BINARY}" .

if [ ! -f "${LOCAL_BINARY}" ]; then
    log_error "编译失败，二进制文件不存在"
    exit 1
fi

log_success "编译完成: ${LOCAL_BINARY}"

# 2. 创建远程连接函数
ssh_exec() {
    sshpass -p "${REMOTE_SERVER_USER_PASSWORD}" ssh -o StrictHostKeyChecking=no -p "${REMOTE_SERVER_PORT}" "${REMOTE_SERVER_USER}@${REMOTE_SERVER_IP}" "$1"
}

scp_upload() {
    sshpass -p "${REMOTE_SERVER_USER_PASSWORD}" scp -o StrictHostKeyChecking=no -P "${REMOTE_SERVER_PORT}" "$1" "${REMOTE_SERVER_USER}@${REMOTE_SERVER_IP}:$2"
}

# 检查 sshpass 是否安装
if ! command -v sshpass &> /dev/null; then
    log_error "sshpass 未安装，请先安装: brew install sshpass (macOS) 或 apt-get install sshpass (Ubuntu)"
    exit 1
fi

# 3. 测试远程连接
log "测试远程服务器连接..."
if ! ssh_exec "echo 'Connection test successful'"; then
    log_error "无法连接到远程服务器"
    exit 1
fi
log_success "远程服务器连接成功"

# 4. 创建远程目录
log "创建远程项目目录..."
ssh_exec "mkdir -p ${REMOTE_SERVER_PROJECT_PATH}"

# 5. 停止现有服务
log "停止远程服务器上的现有服务..."

# 创建远程目录
ssh_exec "mkdir -p ${REMOTE_SERVER_PROJECT_PATH}" || {
    log_warning "创建目录时出现警告，继续部署流程..."
}

# 检查是否有运行中的服务
running_pids=$(ssh_exec "pgrep -f '${BINARY_NAME}' 2>/dev/null || true")

if [ -n "$running_pids" ]; then
    log "发现运行中的服务，PID: $running_pids"
    
    # 优雅停止
    ssh_exec "pkill -f '${BINARY_NAME}' 2>/dev/null || true"
    sleep 3
    
    # 检查是否还在运行
    still_running=$(ssh_exec "pgrep -f '${BINARY_NAME}' 2>/dev/null || true")
    if [ -n "$still_running" ]; then
        log_warning "服务未能优雅停止，强制停止..."
        ssh_exec "pkill -9 -f '${BINARY_NAME}' 2>/dev/null || true"
        sleep 2
    fi
    
    log_success "现有服务已停止"
else
    log "未发现运行中的服务"
fi

# 6. 上传文件
log "上传二进制文件到远程服务器..."
scp_upload "${LOCAL_BINARY}" "${REMOTE_BINARY}"

log "上传 .env 配置文件..."
scp_upload ".env" "${REMOTE_ENV_FILE}"

log "上传 test.sh 测试脚本..."
scp_upload "test.sh" "${REMOTE_TEST_SCRIPT}"

# 7. 设置文件权限
log "设置远程文件权限..."
ssh_exec "
    chmod +x ${REMOTE_BINARY}
    chmod +x ${REMOTE_TEST_SCRIPT}
"

log_success "文件上传完成"

# 8. 运行测试脚本
log "在远程服务器上运行测试脚本..."
test_result=$(ssh_exec "
    cd ${REMOTE_SERVER_PROJECT_PATH}
    export \$(grep -v '^#' .env | xargs) 2>/dev/null || true
    
    # 启动服务进行测试
    echo '启动服务进行测试...'
    nohup ./${BINARY_NAME} > server_test.log 2>&1 &
    SERVER_PID=\$!
    echo \"服务已启动，PID: \$SERVER_PID\"
    
    # 等待服务启动
    echo '等待服务启动...'
    for i in {1..30}; do
        if curl -s http://localhost:\${PORT:-8031}/api/health > /dev/null 2>&1; then
            echo '服务启动成功，开始测试...'
            break
        fi
        echo \"等待中... \$i/30\"
        sleep 1
    done
    
    # 运行简单的健康检查测试
    echo '执行健康检查...'
    health_response=\$(curl -s http://localhost:\${PORT:-8031}/api/health 2>/dev/null || echo 'CURL_FAILED')
    echo \"健康检查响应: \$health_response\"
    
    if echo \"\$health_response\" | grep -q 'ok\\|success\\|healthy\\|正常\\|运行'; then
        echo 'TEST_PASSED'
        # 停止测试服务
        kill \$SERVER_PID 2>/dev/null || true
        sleep 2
        exit 0
    else
        echo 'TEST_FAILED'
        echo \"服务日志:\"
        tail -10 server_test.log 2>/dev/null || echo '无法读取日志'
        # 停止测试服务
        kill \$SERVER_PID 2>/dev/null || true
        sleep 2
        exit 1
    fi
" 2>/dev/null)

if echo "$test_result" | grep -q "TEST_PASSED"; then
    log_success "远程测试通过"
else
    log_error "远程测试失败"
    log "测试输出："
    echo "$test_result"
    exit 1
fi

# 9. 正式启动服务
log "正式启动远程服务..."
start_result=$(ssh_exec "
    cd ${REMOTE_SERVER_PROJECT_PATH}
    export \$(grep -v '^#' .env | xargs) 2>/dev/null || true
    
    # 启动服务
    echo '启动正式服务...'
    nohup ./${BINARY_NAME} > server.log 2>&1 &
    SERVER_PID=\$!
    echo \$SERVER_PID > server.pid
    
    echo \"服务已启动，PID: \$SERVER_PID\"
    
    # 验证服务是否正常运行
    sleep 3
    if ps -p \$SERVER_PID > /dev/null 2>&1; then
        echo 'SERVICE_STARTED'
        echo \"服务运行正常，PID: \$SERVER_PID\"
    else
        echo 'SERVICE_FAILED'
        echo '服务启动失败，查看日志:'
        tail -10 server.log 2>/dev/null || echo '无法读取日志'
        exit 1
    fi
" 2>/dev/null)

if echo "$start_result" | grep -q "SERVICE_STARTED"; then
    log_success "服务启动成功"
else
    log_error "服务启动失败"
    echo "$start_result"
    exit 1
fi

# 10. 最终验证
log "进行最终服务验证..."
final_check=$(ssh_exec "
    cd ${REMOTE_SERVER_PROJECT_PATH}
    export \$(grep -v '^#' .env | xargs)
    
    # 等待服务完全启动
    sleep 5
    
    # 检查健康状态
    if curl -s http://localhost:\${PORT:-8031}/api/health > /dev/null 2>&1; then
        echo 'SERVICE_HEALTHY'
    else
        echo 'SERVICE_UNHEALTHY'
    fi
")

if echo "$final_check" | grep -q "SERVICE_HEALTHY"; then
    log_success "部署成功！服务正在远程服务器上正常运行"
    log "服务地址: http://${REMOTE_SERVER_IP}:${PORT:-8031}"
    log "可以通过以下命令查看服务状态:"
    log "  ssh -p ${REMOTE_SERVER_PORT} ${REMOTE_SERVER_USER}@${REMOTE_SERVER_IP} 'cd ${REMOTE_SERVER_PROJECT_PATH} && tail -f server.log'"
else
    log_error "部署失败，服务未能正常启动"
    exit 1
fi

log_success "部署流程完成！" 