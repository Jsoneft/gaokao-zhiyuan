# 高考志愿填报系统 - Windows构建脚本

Write-Host "🚀 开始编译高考志愿填报系统..." -ForegroundColor Green

# 检查Go是否安装
try {
    $goVersion = go version 2>$null
    Write-Host "✅ Go环境: $goVersion" -ForegroundColor Green
} catch {
    Write-Host "❌ 错误: 未找到Go环境" -ForegroundColor Red
    Write-Host "请安装Go语言环境: winget install GoLang.Go" -ForegroundColor Yellow
    exit 1
}

# 创建bin目录
Write-Host "📁 创建bin目录..." -ForegroundColor Cyan
New-Item -ItemType Directory -Force -Path bin | Out-Null

# 下载依赖
Write-Host "📦 下载Go模块依赖..." -ForegroundColor Cyan
go mod download
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ 下载依赖失败" -ForegroundColor Red
    exit 1
}

go mod tidy
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ 整理依赖失败" -ForegroundColor Red
    exit 1
}

# 编译主程序
Write-Host "🔨 编译主程序..." -ForegroundColor Cyan
go build -o bin/gaokao-server.exe main.go
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ 编译主程序失败" -ForegroundColor Red
    exit 1
}

# 编译导入工具
Write-Host "🔨 编译数据导入工具..." -ForegroundColor Cyan
go build -o bin/import-tool.exe tools/import_excel.go
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ 编译导入工具失败" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "🎉 编译完成！" -ForegroundColor Green
Write-Host "📍 可执行文件位于 bin/ 目录中:" -ForegroundColor Yellow
Write-Host "   - bin/gaokao-server.exe   (主服务程序)" -ForegroundColor White
Write-Host "   - bin/import-tool.exe     (数据导入工具)" -ForegroundColor White
Write-Host ""
Write-Host "📋 接下来的步骤:" -ForegroundColor Yellow
Write-Host "   1. 确保ClickHouse数据库已运行" -ForegroundColor White
Write-Host "   2. 将Excel数据文件放在项目根目录" -ForegroundColor White
Write-Host "   3. 运行: .\run.ps1 import  (导入数据)" -ForegroundColor White
Write-Host "   4. 运行: .\run.ps1 run     (启动服务)" -ForegroundColor White 