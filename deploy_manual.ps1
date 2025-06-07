# 高考志愿填报系统手动部署脚本
# 使用方法：按步骤执行每个命令

$SERVER_IP = "47.96.103.220"
$PORT = "6189"
$USERNAME = "root"
$REMOTE_PATH = "/opt/gaokao-zhiyuan"

Write-Host "=== 高考志愿填报系统部署指南 ===" -ForegroundColor Green
Write-Host ""

Write-Host "步骤1: 连接到服务器并检查环境" -ForegroundColor Yellow
Write-Host "ssh -p $PORT $USERNAME@$SERVER_IP"
Write-Host "然后执行以下命令："
Write-Host "uname -a"
Write-Host "cat /etc/os-release"
Write-Host ""

Write-Host "步骤2: 更新系统和安装基础软件" -ForegroundColor Yellow
Write-Host "dnf update -y"
Write-Host "dnf install -y wget curl git tar gzip"
Write-Host ""

Write-Host "步骤3: 安装Go环境" -ForegroundColor Yellow
Write-Host "cd /tmp"
Write-Host "wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz"
Write-Host "rm -rf /usr/local/go"
Write-Host "tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz"
Write-Host "echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc"
Write-Host "source ~/.bashrc"
Write-Host "export PATH=$PATH:/usr/local/go/bin"
Write-Host "go version"
Write-Host ""

Write-Host "步骤4: 安装ClickHouse" -ForegroundColor Yellow
Write-Host "dnf install -y yum-utils"
Write-Host "rpm --import https://packages.clickhouse.com/rpm/lts/repodata/repomd.xml.key"
Write-Host "dnf config-manager --add-repo https://packages.clickhouse.com/rpm/lts/clickhouse.repo"
Write-Host "dnf install -y clickhouse-server clickhouse-client"
Write-Host "systemctl enable clickhouse-server"
Write-Host "systemctl start clickhouse-server"
Write-Host "systemctl status clickhouse-server"
Write-Host ""

Write-Host "步骤5: 创建数据库" -ForegroundColor Yellow
Write-Host "clickhouse-client --query 'CREATE DATABASE IF NOT EXISTS gaokao'"
Write-Host ""

Write-Host "步骤6: 创建项目目录" -ForegroundColor Yellow
Write-Host "mkdir -p $REMOTE_PATH"
Write-Host "cd $REMOTE_PATH"
Write-Host ""

Write-Host "步骤7: 上传项目文件 (在本地Windows执行)" -ForegroundColor Yellow
Write-Host "在另一个PowerShell窗口中执行："
Write-Host "scp -P $PORT -r * $USERNAME@${SERVER_IP}:$REMOTE_PATH/"
Write-Host ""

Write-Host "步骤8: 编译项目 (在服务器执行)" -ForegroundColor Yellow
Write-Host "cd $REMOTE_PATH"
Write-Host "export PATH=$PATH:/usr/local/go/bin"
Write-Host "go mod download"
Write-Host "go build -o gaokao-server main.go"
Write-Host "go build -o import-tool tools/import_excel.go"
Write-Host ""

Write-Host "步骤9: 导入数据" -ForegroundColor Yellow
Write-Host "export CLICKHOUSE_HOST=localhost"
Write-Host "export CLICKHOUSE_PORT=9000"
Write-Host "export CLICKHOUSE_USERNAME=default"
Write-Host "export CLICKHOUSE_PASSWORD="
Write-Host "export CLICKHOUSE_DATABASE=gaokao"
Write-Host "./import-tool"
Write-Host ""

Write-Host "步骤10: 创建系统服务" -ForegroundColor Yellow
Write-Host "创建服务文件："
Write-Host "cat > /etc/systemd/system/gaokao-server.service << 'EOF'"
Write-Host "[Unit]"
Write-Host "Description=Gaokao Zhiyuan Server"
Write-Host "After=network.target clickhouse-server.service"
Write-Host "Requires=clickhouse-server.service"
Write-Host ""
Write-Host "[Service]"
Write-Host "Type=simple"
Write-Host "User=root"
Write-Host "WorkingDirectory=$REMOTE_PATH"
Write-Host "ExecStart=$REMOTE_PATH/gaokao-server"
Write-Host "Restart=always"
Write-Host "RestartSec=5"
Write-Host "Environment=PORT=8031"
Write-Host "Environment=GIN_MODE=release"
Write-Host "Environment=CLICKHOUSE_HOST=localhost"
Write-Host "Environment=CLICKHOUSE_PORT=9000"
Write-Host "Environment=CLICKHOUSE_USERNAME=default"
Write-Host "Environment=CLICKHOUSE_PASSWORD="
Write-Host "Environment=CLICKHOUSE_DATABASE=gaokao"
Write-Host ""
Write-Host "[Install]"
Write-Host "WantedBy=multi-user.target"
Write-Host "EOF"
Write-Host ""

Write-Host "步骤11: 启动服务" -ForegroundColor Yellow
Write-Host "systemctl daemon-reload"
Write-Host "systemctl enable gaokao-server"
Write-Host "systemctl start gaokao-server"
Write-Host "systemctl status gaokao-server"
Write-Host ""

Write-Host "步骤12: 配置防火墙" -ForegroundColor Yellow
Write-Host "firewall-cmd --permanent --add-port=8031/tcp"
Write-Host "firewall-cmd --permanent --add-port=9000/tcp"
Write-Host "firewall-cmd --reload"
Write-Host ""

Write-Host "步骤13: 验证接口" -ForegroundColor Yellow
Write-Host "curl http://localhost:8031/api/health"
Write-Host "curl 'http://localhost:8031/api/rank/get?score=555'"
Write-Host ""

Write-Host "外部访问测试:" -ForegroundColor Green
Write-Host "http://${SERVER_IP}:8031/api/health"
Write-Host "http://${SERVER_IP}:8031/api/rank/get?score=555"
Write-Host ""

Write-Host "=== 准备开始部署 ===" -ForegroundColor Green
Write-Host "请先连接到服务器："
Write-Host "ssh -p $PORT $USERNAME@$SERVER_IP" -ForegroundColor Cyan 