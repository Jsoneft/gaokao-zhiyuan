# 高考志愿填报系统部署指南

## 完整的一键部署流程

### 前提条件

1. **本地环境**:
   - Git
   - SSH客户端
   - 有目标服务器的SSH访问权限

2. **目标服务器**:
   - Ubuntu 18.04+ 或 Debian 9+
   - 具有sudo权限的用户
   - SSH服务运行中

### 部署步骤

#### 1. 准备SSH密钥认证

```bash
# 生成SSH密钥（如果还没有）
ssh-keygen -t rsa -b 4096 -C "your_email@example.com"

# 将公钥复制到服务器
ssh-copy-id username@server_ip

# 测试SSH连接
ssh username@server_ip
```

#### 2. 克隆项目并部署

```bash
# 克隆项目
git clone <repository-url>
cd gaokao-zhiyuan

# 确保Excel数据文件存在
ls -la "21-24各省份录取数据(含专业组代码).xlsx"

# 一键部署到服务器
make deploy SERVER=192.168.1.100 USERNAME=root

# 或者指定SSH端口
make deploy SERVER=192.168.1.100 USERNAME=ubuntu PORT=2222
```

#### 3. 验证部署

部署完成后，你会看到：

```
✅ 部署成功！

🎉 服务信息:
   API地址: http://192.168.1.100:8031
   健康检查: http://192.168.1.100:8031/api/health
   位次查询: http://192.168.1.100:8031/api/rank/get?score=555
   报表查询: http://192.168.1.100:8031/api/report/get?rank=12000&class_comb="123"

📋 管理命令:
   查看服务状态: sudo systemctl status gaokao-server
   查看服务日志: sudo journalctl -u gaokao-server -f
   重启服务: sudo systemctl restart gaokao-server
   停止服务: sudo systemctl stop gaokao-server
```

#### 4. 测试API接口

```bash
# 健康检查
curl http://your-server-ip:8031/api/health

# 位次查询
curl "http://your-server-ip:8031/api/rank/get?score=555"

# 报表查询
curl "http://your-server-ip:8031/api/report/get?rank=12000&class_comb=\"123\"&page=1&page_size=5"
```

## 手动部署步骤

如果自动部署脚本失败，可以按以下步骤手动部署：

### 1. 服务器环境准备

```bash
# 连接到服务器
ssh username@server_ip

# 更新系统
sudo apt update && sudo apt upgrade -y

# 安装必要工具
sudo apt install -y wget curl git build-essential
```

### 2. 安装Go环境

```bash
# 下载并安装Go
cd /tmp
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz

# 设置环境变量
echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee -a /etc/profile
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# 验证安装
go version
```

### 3. 安装ClickHouse

```bash
# 添加ClickHouse仓库
sudo apt-get install -y apt-transport-https ca-certificates dirmngr
sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv 8919F6BD2B48D754
echo "deb https://packages.clickhouse.com/deb stable main" | sudo tee /etc/apt/sources.list.d/clickhouse.list

# 安装ClickHouse
sudo apt-get update
sudo apt-get install -y clickhouse-server clickhouse-client

# 启动服务
sudo systemctl enable clickhouse-server
sudo systemctl start clickhouse-server

# 创建数据库
clickhouse-client --query "CREATE DATABASE IF NOT EXISTS gaokao"
```

### 4. 部署应用

```bash
# 创建项目目录
sudo mkdir -p /opt/gaokao-zhiyuan
sudo chown $USER:$USER /opt/gaokao-zhiyuan
cd /opt/gaokao-zhiyuan

# 上传项目文件（从本地）
# scp -r ./* username@server_ip:/opt/gaokao-zhiyuan/

# 编译项目
go mod download
go build -o gaokao-server main.go
go build -o import-tool tools/import_excel.go

# 导入数据
./import-tool

# 创建系统服务
sudo tee /etc/systemd/system/gaokao-server.service > /dev/null <<EOF
[Unit]
Description=Gaokao Zhiyuan Server
After=network.target clickhouse-server.service
Requires=clickhouse-server.service

[Service]
Type=simple
User=$USER
WorkingDirectory=/opt/gaokao-zhiyuan
ExecStart=/opt/gaokao-zhiyuan/gaokao-server
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
EOF

# 启动服务
sudo systemctl daemon-reload
sudo systemctl enable gaokao-server
sudo systemctl start gaokao-server

# 配置防火墙
sudo ufw allow 8031/tcp
sudo ufw allow 9000/tcp
```

## 环境变量配置

可以通过以下方式配置服务：

### 1. 修改systemd服务文件

```bash
sudo systemctl edit gaokao-server
```

添加：
```ini
[Service]
Environment=PORT=8031
Environment=CLICKHOUSE_HOST=localhost
Environment=CLICKHOUSE_PORT=9000
Environment=CLICKHOUSE_DATABASE=gaokao
```

### 2. 使用环境文件

```bash
# 创建环境文件
sudo tee /opt/gaokao-zhiyuan/.env > /dev/null <<EOF
PORT=8031
GIN_MODE=release
CLICKHOUSE_HOST=localhost
CLICKHOUSE_PORT=9000
CLICKHOUSE_USERNAME=default
CLICKHOUSE_PASSWORD=
CLICKHOUSE_DATABASE=gaokao
EOF

# 修改服务文件使用环境文件
sudo systemctl edit gaokao-server
```

添加：
```ini
[Service]
EnvironmentFile=/opt/gaokao-zhiyuan/.env
```

## 维护和监控

### 日志查看

```bash
# 查看服务日志
sudo journalctl -u gaokao-server -f

# 查看ClickHouse日志
sudo tail -f /var/log/clickhouse-server/clickhouse-server.log
```

### 服务管理

```bash
# 查看服务状态
sudo systemctl status gaokao-server

# 重启服务
sudo systemctl restart gaokao-server

# 停止服务
sudo systemctl stop gaokao-server

# 查看服务配置
sudo systemctl cat gaokao-server
```

### 数据库管理

```bash
# 连接ClickHouse
clickhouse-client

# 查看数据库
SHOW DATABASES;

# 查看表
USE gaokao;
SHOW TABLES;

# 查看数据量
SELECT count() FROM admission_data;

# 查看数据样例
SELECT * FROM admission_data LIMIT 5;
```

### 备份和恢复

```bash
# 备份数据
clickhouse-client --query "SELECT * FROM gaokao.admission_data FORMAT Native" > backup.native

# 恢复数据
clickhouse-client --query "INSERT INTO gaokao.admission_data FORMAT Native" < backup.native
```

## 故障排查

### 常见问题

1. **服务无法启动**
   ```bash
   sudo systemctl status gaokao-server
   sudo journalctl -u gaokao-server --no-pager
   ```

2. **数据库连接失败**
   ```bash
   sudo systemctl status clickhouse-server
   clickhouse-client --query "SELECT 1"
   ```

3. **端口被占用**
   ```bash
   sudo netstat -tlnp | grep 8031
   sudo lsof -i :8031
   ```

4. **权限问题**
   ```bash
   sudo chown -R $USER:$USER /opt/gaokao-zhiyuan
   sudo chmod +x /opt/gaokao-zhiyuan/gaokao-server
   ```

### 性能优化

1. **ClickHouse优化**
   ```sql
   # 添加索引
   ALTER TABLE admission_data ADD INDEX idx_year_points (year, lowest_points) TYPE minmax GRANULARITY 1;
   
   # 优化表结构
   OPTIMIZE TABLE admission_data;
   ```

2. **系统资源监控**
   ```bash
   # 查看系统资源
   htop
   
   # 查看磁盘空间
   df -h
   
   # 查看内存使用
   free -h
   ```

## 安全建议

1. **防火墙配置**
   ```bash
   sudo ufw enable
   sudo ufw allow ssh
   sudo ufw allow 8031/tcp
   ```

2. **SSL/TLS配置**
   - 使用Nginx作为反向代理
   - 配置Let's Encrypt证书

3. **数据库安全**
   - 设置ClickHouse密码
   - 限制网络访问

---

如有问题，请查阅错误日志或联系系统管理员。 