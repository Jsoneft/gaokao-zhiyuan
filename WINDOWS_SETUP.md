# Windows环境设置指南

## 前提条件安装

### 1. 安装Go语言环境

#### 方法一：通过官网下载安装
1. 访问 [Go官网](https://golang.org/dl/)
2. 下载Windows版本的安装包 (`go1.21.x.windows-amd64.msi`)
3. 运行安装包，按默认设置安装
4. 重启命令行窗口

#### 方法二：通过winget安装（Windows 10/11）
```powershell
winget install GoLang.Go
```

### 2. 验证Go安装
```powershell
go version
```

### 3. 安装Make工具（可选）

#### 方法一：安装Chocolatey + Make
```powershell
# 安装Chocolatey（以管理员身份运行PowerShell）
Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))

# 安装make
choco install make
```

#### 方法二：安装Git for Windows（包含make）
1. 下载 [Git for Windows](https://gitforwindows.org/)
2. 安装时选择包含Linux工具
3. 使用Git Bash终端运行make命令

## Windows下的项目使用指南

### 直接使用Go命令（无需make）

```powershell
# 1. 下载依赖
go mod download
go mod tidy

# 2. 编译项目
# 创建bin目录
New-Item -ItemType Directory -Force -Path bin

# 编译主程序
go build -o bin/gaokao-server.exe main.go

# 编译数据导入工具
go build -o bin/import-tool.exe tools/import_excel.go

# 3. 运行程序
# 导入数据（需要先准备Excel文件）
./bin/import-tool.exe

# 启动服务器
./bin/gaokao-server.exe

# 4. 清理编译文件
Remove-Item -Recurse -Force bin
```

### Windows PowerShell脚本

创建 `build.ps1` 脚本：
```powershell
# build.ps1
Write-Host "编译高考志愿填报系统..."

# 创建bin目录
New-Item -ItemType Directory -Force -Path bin | Out-Null

# 下载依赖
Write-Host "下载Go模块依赖..."
go mod download
go mod tidy

# 编译主程序
Write-Host "编译主程序..."
go build -o bin/gaokao-server.exe main.go

# 编译导入工具
Write-Host "编译导入工具..."
go build -o bin/import-tool.exe tools/import_excel.go

Write-Host "编译完成！"
Write-Host "可执行文件位于 bin/ 目录中"
```

创建 `run.ps1` 脚本：
```powershell
# run.ps1
param(
    [string]$Command = "help"
)

switch ($Command) {
    "build" {
        ./build.ps1
    }
    "run" {
        Write-Host "启动服务器..."
        ./bin/gaokao-server.exe
    }
    "import" {
        Write-Host "导入Excel数据..."
        ./bin/import-tool.exe
    }
    "clean" {
        Write-Host "清理编译文件..."
        Remove-Item -Recurse -Force bin -ErrorAction SilentlyContinue
    }
    default {
        Write-Host "高考志愿填报系统 - Windows版本"
        Write-Host ""
        Write-Host "使用方法:"
        Write-Host "  .\run.ps1 build   - 编译项目"
        Write-Host "  .\run.ps1 run     - 运行服务器"
        Write-Host "  .\run.ps1 import  - 导入Excel数据"
        Write-Host "  .\run.ps1 clean   - 清理编译文件"
        Write-Host ""
        Write-Host "示例:"
        Write-Host "  .\run.ps1 build"
        Write-Host "  .\run.ps1 import"
        Write-Host "  .\run.ps1 run"
    }
}
```

## 环境变量设置

### 临时设置（当前会话有效）
```powershell
$env:PORT = "8031"
$env:GIN_MODE = "release"
$env:CLICKHOUSE_HOST = "localhost"
$env:CLICKHOUSE_PORT = "9000"
$env:CLICKHOUSE_USERNAME = "default"
$env:CLICKHOUSE_PASSWORD = ""
$env:CLICKHOUSE_DATABASE = "gaokao"
```

### 永久设置
```powershell
# 设置用户环境变量
[Environment]::SetEnvironmentVariable("PORT", "8031", "User")
[Environment]::SetEnvironmentVariable("GIN_MODE", "release", "User")
[Environment]::SetEnvironmentVariable("CLICKHOUSE_HOST", "localhost", "User")
[Environment]::SetEnvironmentVariable("CLICKHOUSE_PORT", "9000", "User")
[Environment]::SetEnvironmentVariable("CLICKHOUSE_DATABASE", "gaokao", "User")
```

## Windows下的ClickHouse安装

### 方法一：使用Docker（推荐）
```powershell
# 安装Docker Desktop for Windows
# 下载地址: https://www.docker.com/products/docker-desktop

# 启动ClickHouse容器
docker run -d --name clickhouse-server `
  -p 8123:8123 -p 9000:9000 `
  --ulimit nofile=262144:262144 `
  clickhouse/clickhouse-server:latest

# 创建数据库
docker exec -it clickhouse-server clickhouse-client --query "CREATE DATABASE IF NOT EXISTS gaokao"
```

### 方法二：直接安装
1. 下载ClickHouse Windows版本
2. 解压到指定目录
3. 配置Windows服务
4. 启动服务

## 常见问题解决

### 1. PowerShell执行策略错误
```powershell
# 允许执行脚本（以管理员身份运行）
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

### 2. 端口被占用
```powershell
# 查看端口占用
netstat -ano | findstr :8031

# 结束进程
taskkill /PID <进程ID> /F
```

### 3. Go模块下载慢
```powershell
# 设置Go代理（中国用户）
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOSUMDB=sum.golang.google.cn
```

## 快速开始（Windows版）

1. **安装Go**：
   ```powershell
   winget install GoLang.Go
   ```

2. **克隆项目**：
   ```powershell
   git clone https://github.com/Jsoneft/gaokao-zhiyuan.git
   cd gaokao-zhiyuan
   ```

3. **准备数据文件**：
   - 将 `21-24各省份录取数据(含专业组代码).xlsx` 放在项目根目录

4. **编译和运行**：
   ```powershell
   # 下载依赖
   go mod download

   # 编译
   go build -o bin/gaokao-server.exe main.go
   go build -o bin/import-tool.exe tools/import_excel.go

   # 导入数据（需要ClickHouse运行）
   ./bin/import-tool.exe

   # 启动服务
   ./bin/gaokao-server.exe
   ```

5. **测试API**：
   ```powershell
   # 健康检查
   Invoke-RestMethod -Uri "http://localhost:8031/api/health"

   # 位次查询
   Invoke-RestMethod -Uri "http://localhost:8031/api/rank/get?score=555"
   ```

---

如有问题，请参考项目文档或提交Issue。 