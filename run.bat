@echo off
chcp 65001 >nul
echo 高考志愿填报系统 - Windows管理工具
echo.

if "%1"=="" goto help
if "%1"=="help" goto help
if "%1"=="deps" goto deps
if "%1"=="build" goto build
if "%1"=="run" goto run
if "%1"=="import" goto import
if "%1"=="clean" goto clean
if "%1"=="setup" goto setup
goto help

:help
echo 可用命令:
echo   run.bat deps    - 下载依赖
echo   run.bat build   - 编译项目
echo   run.bat run     - 启动服务器
echo   run.bat import  - 导入Excel数据
echo   run.bat clean   - 清理编译文件
echo   run.bat setup   - 环境检查
echo   run.bat help    - 显示帮助
echo.
echo 示例:
echo   run.bat build
echo   run.bat run
goto end

:deps
echo 下载Go模块依赖...
go mod download
go mod tidy
echo 依赖下载完成
goto end

:build
echo 编译项目...
if not exist bin mkdir bin
echo 编译主程序...
go build -o bin/gaokao-server.exe main.go
echo 编译导入工具...
go build -o bin/import-tool.exe tools/import_excel.go
echo 编译完成！
goto end

:run
echo 启动服务器...
if not exist bin/gaokao-server.exe (
    echo 未找到可执行文件，请先运行: run.bat build
    goto end
)
set PORT=8031
set GIN_MODE=release
set CLICKHOUSE_HOST=localhost
set CLICKHOUSE_PORT=9000
set CLICKHOUSE_USERNAME=default
set CLICKHOUSE_PASSWORD=
set CLICKHOUSE_DATABASE=gaokao
echo 服务将在 http://localhost:8031 启动
echo 按 Ctrl+C 停止服务
echo.
bin\gaokao-server.exe
goto end

:import
echo 导入Excel数据...
if not exist bin/import-tool.exe (
    echo 未找到导入工具，请先运行: run.bat build
    goto end
)
set CLICKHOUSE_HOST=localhost
set CLICKHOUSE_PORT=9000
set CLICKHOUSE_USERNAME=default
set CLICKHOUSE_PASSWORD=
set CLICKHOUSE_DATABASE=gaokao
echo 开始导入数据...
bin\import-tool.exe
goto end

:clean
echo 清理编译文件...
if exist bin rmdir /s /q bin
echo 清理完成
goto end

:setup
echo 环境设置向导
echo.
go version >nul 2>&1
if errorlevel 1 (
    echo [ERROR] 未找到Go环境
    echo 请运行: winget install GoLang.Go
    goto end
) else (
    echo [OK] Go环境已安装
)

if exist main.go (echo [OK] main.go) else (echo [ERROR] main.go)
if exist go.mod (echo [OK] go.mod) else (echo [ERROR] go.mod)
if exist config\config.go (echo [OK] config\config.go) else (echo [ERROR] config\config.go)

dir *.xlsx >nul 2>&1
if errorlevel 1 (
    echo [WARNING] Excel数据文件不存在
    echo 请查阅 DATA_SETUP.md 获取数据文件
) else (
    echo [OK] Excel数据文件存在
)

echo.
echo 建议的设置步骤:
echo 1. run.bat deps
echo 2. run.bat build  
echo 3. 启动ClickHouse数据库
echo 4. run.bat import
echo 5. run.bat run
goto end

:end 