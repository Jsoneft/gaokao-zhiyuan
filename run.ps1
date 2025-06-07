param(
    [string]$Command = "help"
)

function Show-Help {
    Write-Host "高考志愿填报系统 - Windows版本管理工具" -ForegroundColor Green
    Write-Host ""
    Write-Host "可用命令:" -ForegroundColor Yellow
    Write-Host "  deps       - 下载依赖" -ForegroundColor Cyan
    Write-Host "  build      - 编译项目" -ForegroundColor Cyan
    Write-Host "  run        - 启动服务器" -ForegroundColor Cyan
    Write-Host "  import     - 导入Excel数据" -ForegroundColor Cyan
    Write-Host "  clean      - 清理编译文件" -ForegroundColor Cyan
    Write-Host "  test       - 运行测试" -ForegroundColor Cyan
    Write-Host "  setup      - 环境设置向导" -ForegroundColor Cyan
    Write-Host "  health     - 健康检查" -ForegroundColor Cyan
    Write-Host "  help       - 显示此帮助信息" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "示例:" -ForegroundColor Yellow
    Write-Host "  .\run.ps1 deps" -ForegroundColor White
    Write-Host "  .\run.ps1 build" -ForegroundColor White
    Write-Host "  .\run.ps1 run" -ForegroundColor White
}

function Download-Deps {
    Write-Host "下载Go模块依赖..." -ForegroundColor Green
    go mod download
    go mod tidy
    Write-Host "依赖下载完成" -ForegroundColor Green
}

function Build-Project {
    Write-Host "编译项目..." -ForegroundColor Green
    
    New-Item -ItemType Directory -Force -Path bin | Out-Null
    
    Write-Host "编译主程序..." -ForegroundColor Cyan
    go build -o bin/gaokao-server.exe main.go
    
    Write-Host "编译导入工具..." -ForegroundColor Cyan
    go build -o bin/import-tool.exe tools/import_excel.go
    
    Write-Host "编译完成！" -ForegroundColor Green
}

function Run-Server {
    Write-Host "启动服务器..." -ForegroundColor Green
    
    if (-not (Test-Path "bin/gaokao-server.exe")) {
        Write-Host "未找到可执行文件，请先运行: .\run.ps1 build" -ForegroundColor Red
        return
    }
    
    $env:PORT = "8031"
    $env:GIN_MODE = "release"
    $env:CLICKHOUSE_HOST = "localhost"
    $env:CLICKHOUSE_PORT = "9000"
    $env:CLICKHOUSE_USERNAME = "default"
    $env:CLICKHOUSE_PASSWORD = ""
    $env:CLICKHOUSE_DATABASE = "gaokao"
    
    Write-Host "服务将在 http://localhost:8031 启动" -ForegroundColor Cyan
    Write-Host "按 Ctrl+C 停止服务" -ForegroundColor Yellow
    
    .\bin\gaokao-server.exe
}

function Import-Data {
    Write-Host "导入Excel数据..." -ForegroundColor Green
    
    if (-not (Test-Path "bin/import-tool.exe")) {
        Write-Host "未找到导入工具，请先运行: .\run.ps1 build" -ForegroundColor Red
        return
    }
    
    $excelFiles = Get-ChildItem -Filter "*.xlsx"
    if ($excelFiles.Count -eq 0) {
        Write-Host "未找到Excel数据文件" -ForegroundColor Red
        Write-Host "请将Excel文件放在项目根目录" -ForegroundColor Yellow
        return
    }
    
    $env:CLICKHOUSE_HOST = "localhost"
    $env:CLICKHOUSE_PORT = "9000"
    $env:CLICKHOUSE_USERNAME = "default"
    $env:CLICKHOUSE_PASSWORD = ""
    $env:CLICKHOUSE_DATABASE = "gaokao"
    
    Write-Host "开始导入数据..." -ForegroundColor Cyan
    .\bin\import-tool.exe
}

function Clean-Project {
    Write-Host "清理编译文件..." -ForegroundColor Green
    
    if (Test-Path "bin") {
        Remove-Item -Recurse -Force bin
        Write-Host "已删除 bin/ 目录" -ForegroundColor Green
    }
    
    go clean
    Write-Host "清理完成" -ForegroundColor Green
}

function Test-Project {
    Write-Host "运行测试..." -ForegroundColor Green
    go test ./...
}

function Setup-Environment {
    Write-Host "环境设置向导" -ForegroundColor Green
    Write-Host ""
    
    try {
        $goVersion = go version 2>$null
        Write-Host "Go环境: $goVersion" -ForegroundColor Green
    } catch {
        Write-Host "未找到Go环境" -ForegroundColor Red
        Write-Host "请运行: winget install GoLang.Go" -ForegroundColor Yellow
        return
    }
    
    $requiredFiles = @("main.go", "go.mod", "config/config.go")
    foreach ($file in $requiredFiles) {
        if (Test-Path $file) {
            Write-Host "✓ $file" -ForegroundColor Green
        } else {
            Write-Host "✗ $file" -ForegroundColor Red
        }
    }
    
    $excelFiles = Get-ChildItem -Filter "*.xlsx"
    if ($excelFiles.Count -gt 0) {
        Write-Host "Excel数据文件存在" -ForegroundColor Green
    } else {
        Write-Host "Excel数据文件不存在" -ForegroundColor Yellow
        Write-Host "请查阅 DATA_SETUP.md 获取数据文件" -ForegroundColor Cyan
    }
    
    Write-Host ""
    Write-Host "建议的设置步骤:" -ForegroundColor Yellow
    Write-Host "1. .\run.ps1 deps" -ForegroundColor White
    Write-Host "2. .\run.ps1 build" -ForegroundColor White
    Write-Host "3. 启动ClickHouse数据库" -ForegroundColor White
    Write-Host "4. .\run.ps1 import" -ForegroundColor White
    Write-Host "5. .\run.ps1 run" -ForegroundColor White
}

function Test-Health {
    Write-Host "检查API健康状态..." -ForegroundColor Green
    
    try {
        $response = Invoke-RestMethod -Uri "http://localhost:8031/api/health" -TimeoutSec 5
        Write-Host "API服务正常运行" -ForegroundColor Green
        Write-Host "响应: $($response | ConvertTo-Json)" -ForegroundColor Cyan
    } catch {
        Write-Host "无法连接到API服务" -ForegroundColor Red
        Write-Host "请确认服务是否已启动: .\run.ps1 run" -ForegroundColor Yellow
    }
}

switch ($Command.ToLower()) {
    "deps" { Download-Deps }
    "build" { Build-Project }
    "run" { Run-Server }
    "import" { Import-Data }
    "clean" { Clean-Project }
    "test" { Test-Project }
    "setup" { Setup-Environment }
    "health" { Test-Health }
    "help" { Show-Help }
    default { Show-Help }
} 