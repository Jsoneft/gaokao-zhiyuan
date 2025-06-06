#!/bin/bash

# 高考志愿填报系统远程部署脚本
# 使用方法: ./deploy.sh <服务器IP> <用户名> [端口]

set -e

# 检查参数
if [ $# -lt 2 ]; then
    echo "使用方法: $0 <服务器IP> <用户名> [端口]"
    echo "例如: $0 192.168.1.100 root"
    echo "例如: $0 192.168.1.100 ubuntu 2222"
    exit 1
fi

SERVER_IP=$1
USERNAME=$2
PORT=${3:-22}
PROJECT_NAME="gaokao-zhiyuan"
REMOTE_PATH="/opt/$PROJECT_NAME"

echo "开始部署到服务器 $USERNAME@$SERVER_IP:$PORT"

# 检查SSH连接
echo "检查SSH连接..."
if ! ssh -p $PORT -o ConnectTimeout=10 $USERNAME@$SERVER_IP "echo 'SSH连接成功'"; then
    echo "错误: 无法连接到服务器 $USERNAME@$SERVER_IP:$PORT"
    echo "请检查:"
    echo "1. 服务器IP地址是否正确"
    echo "2. SSH服务是否运行"
    echo "3. 用户名和密钥是否正确"
    echo "4. 端口是否正确"
    exit 1
fi

# 安装Go环境
echo "安装Go环境..."
ssh -p $PORT $USERNAME@$SERVER_IP << 'EOF'
    # 检查Go是否已安装
    if command -v go &> /dev/null; then
        echo "Go已安装，版本: $(go version)"
    else
        echo "安装Go..."
        cd /tmp
        wget -q https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
        sudo rm -rf /usr/local/go
        sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
        
        # 添加到PATH
        echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee -a /etc/profile
        echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
        source ~/.bashrc
        
        # 验证安装
        /usr/local/go/bin/go version
    fi
EOF

# 创建项目目录
echo "创建项目目录..."
ssh -p $PORT $USERNAME@$SERVER_IP "sudo mkdir -p $REMOTE_PATH && sudo chown $USERNAME:$USERNAME $REMOTE_PATH"

# 上传项目文件
echo "上传项目文件..."
scp -P $PORT -r ./* $USERNAME@$SERVER_IP:$REMOTE_PATH/

# 上传Excel数据文件（如果存在）
if [ -f "21-24各省份录取数据(含专业组代码).xlsx" ]; then
    echo "上传Excel数据文件..."
    scp -P $PORT "21-24各省份录取数据(含专业组代码).xlsx" $USERNAME@$SERVER_IP:$REMOTE_PATH/
fi

# 在服务器上安装ClickHouse
echo "安装ClickHouse..."
ssh -p $PORT $USERNAME@$SERVER_IP << EOF
    cd $REMOTE_PATH
    chmod +x scripts/install_clickhouse.sh
    sudo scripts/install_clickhouse.sh
EOF

# 编译项目
echo "编译项目..."
ssh -p $PORT $USERNAME@$SERVER_IP << EOF
    cd $REMOTE_PATH
    export PATH=\$PATH:/usr/local/go/bin
    go mod download
    go build -o gaokao-server main.go
    go build -o import-tool tools/import_excel.go
EOF

# 导入数据（如果数据文件存在）
echo "检查并导入Excel数据..."
ssh -p $PORT $USERNAME@$SERVER_IP << EOF
    cd $REMOTE_PATH
    if [ -f "21-24各省份录取数据(含专业组代码).xlsx" ]; then
        echo "数据文件存在，开始导入..."
        export CLICKHOUSE_HOST=localhost
        export CLICKHOUSE_PORT=9000
        export CLICKHOUSE_USERNAME=default
        export CLICKHOUSE_PASSWORD=
        export CLICKHOUSE_DATABASE=gaokao
        ./import-tool
    else
        echo "警告: 数据文件不存在，跳过数据导入"
        echo "请手动上传数据文件后运行: ./import-tool"
    fi
EOF

# 创建systemd服务文件
echo "创建系统服务..."
ssh -p $PORT $USERNAME@$SERVER_IP << EOF
    sudo tee /etc/systemd/system/gaokao-server.service > /dev/null <<EOL
[Unit]
Description=Gaokao Zhiyuan Server
After=network.target clickhouse-server.service
Requires=clickhouse-server.service

[Service]
Type=simple
User=$USERNAME
WorkingDirectory=$REMOTE_PATH
ExecStart=$REMOTE_PATH/gaokao-server
Restart=always
RestartSec=5
Environment=PORT=8031
Environment=GIN_MODE=release
Environment=CLICKHOUSE_HOST=localhost
Environment=CLICKHOUSE_PORT=9000
Environment=CLICKHOUSE_USERNAME=default
Environment=CLICKHOUSE_PASSWORD=
Environment=CLICKHOUSE_DATABASE=gaokao

[Install]
WantedBy=multi-user.target
EOL

    # 启用并启动服务
    sudo systemctl daemon-reload
    sudo systemctl enable gaokao-server
    sudo systemctl start gaokao-server
    
    # 检查服务状态
    sleep 3
    sudo systemctl status gaokao-server
EOF

# 配置防火墙
echo "配置防火墙..."
ssh -p $PORT $USERNAME@$SERVER_IP << 'EOF'
    # 检查ufw是否安装
    if command -v ufw &> /dev/null; then
        sudo ufw allow 8031/tcp
        sudo ufw allow 9000/tcp
        sudo ufw allow 8123/tcp
        echo "防火墙规则已添加"
    else
        echo "ufw未安装，请手动配置防火墙开放端口 8031, 9000, 8123"
    fi
EOF

# 验证部署
echo "验证部署..."
sleep 5
if ssh -p $PORT $USERNAME@$SERVER_IP "curl -s http://localhost:8031/api/health" | grep -q "ok"; then
    echo "✅ 部署成功！"
    echo ""
    echo "🎉 服务信息:"
    echo "   API地址: http://$SERVER_IP:8031"
    echo "   健康检查: http://$SERVER_IP:8031/api/health"
    echo "   位次查询: http://$SERVER_IP:8031/api/rank/get?score=555"
    echo "   报表查询: http://$SERVER_IP:8031/api/report/get?rank=12000&class_comb=\"123\""
    echo ""
    echo "📋 管理命令:"
    echo "   查看服务状态: sudo systemctl status gaokao-server"
    echo "   查看服务日志: sudo journalctl -u gaokao-server -f"
    echo "   重启服务: sudo systemctl restart gaokao-server"
    echo "   停止服务: sudo systemctl stop gaokao-server"
    echo ""
    echo "⚠️  数据文件提醒:"
    echo "   如果还没有导入数据，请上传Excel文件后运行数据导入工具"
    echo "   scp \"21-24各省份录取数据(含专业组代码).xlsx\" $USERNAME@$SERVER_IP:$REMOTE_PATH/"
    echo "   ssh $USERNAME@$SERVER_IP \"cd $REMOTE_PATH && ./import-tool\""
else
    echo "❌ 部署可能有问题，请检查服务状态"
    ssh -p $PORT $USERNAME@$SERVER_IP "sudo systemctl status gaokao-server"
    exit 1
fi 