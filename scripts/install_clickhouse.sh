#!/bin/bash

# ClickHouse 安装和配置脚本
# 适用于 Ubuntu/Debian 系统

set -e

echo "开始安装 ClickHouse..."

# 更新系统包
sudo apt-get update

# 安装必要的工具
sudo apt-get install -y apt-transport-https ca-certificates dirmngr

# 添加 ClickHouse 官方 GPG 密钥
sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv 8919F6BD2B48D754

# 添加 ClickHouse 官方仓库
echo "deb https://packages.clickhouse.com/deb stable main" | sudo tee /etc/apt/sources.list.d/clickhouse.list

# 更新包列表
sudo apt-get update

# 安装 ClickHouse 服务器和客户端
sudo apt-get install -y clickhouse-server clickhouse-client

# 创建数据目录
sudo mkdir -p /var/lib/clickhouse
sudo mkdir -p /var/log/clickhouse-server
sudo chown clickhouse:clickhouse /var/lib/clickhouse
sudo chown clickhouse:clickhouse /var/log/clickhouse-server

# 配置 ClickHouse
sudo tee /etc/clickhouse-server/config.d/listen.xml > /dev/null <<EOF
<clickhouse>
    <listen_host>0.0.0.0</listen_host>
    <http_port>8123</http_port>
    <tcp_port>9000</tcp_port>
</clickhouse>
EOF

# 配置用户
sudo tee /etc/clickhouse-server/users.d/default.xml > /dev/null <<EOF
<clickhouse>
    <users>
        <default>
            <password></password>
            <networks>
                <ip>::/0</ip>
            </networks>
            <profile>default</profile>
            <quota>default</quota>
            <access_management>1</access_management>
        </default>
    </users>
</clickhouse>
EOF

# 启动 ClickHouse 服务
sudo systemctl enable clickhouse-server
sudo systemctl start clickhouse-server

# 等待服务启动
echo "等待 ClickHouse 服务启动..."
sleep 5

# 检查服务状态
if sudo systemctl is-active --quiet clickhouse-server; then
    echo "ClickHouse 服务启动成功！"
else
    echo "ClickHouse 服务启动失败！"
    sudo systemctl status clickhouse-server
    exit 1
fi

# 创建数据库
echo "创建数据库..."
clickhouse-client --host localhost --port 9000 --query "CREATE DATABASE IF NOT EXISTS gaokao"

echo "ClickHouse 安装和配置完成！"
echo "数据库连接信息："
echo "  主机: localhost"
echo "  端口: 9000 (TCP), 8123 (HTTP)"
echo "  数据库: gaokao"
echo "  用户: default"
echo "  密码: (空)"

# 显示防火墙配置建议
echo ""
echo "如需远程访问，请配置防火墙："
echo "  sudo ufw allow 9000"
echo "  sudo ufw allow 8123" 