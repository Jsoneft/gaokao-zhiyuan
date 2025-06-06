# 高考志愿填报系统 Makefile

.PHONY: build run import clean test docker-build docker-run deploy help

# 默认目标
help:
	@echo "高考志愿填报系统 - 可用命令:"
	@echo ""
	@echo "  build        编译项目"
	@echo "  run          运行服务器"
	@echo "  import       导入Excel数据"
	@echo "  clean        清理编译文件"
	@echo "  test         运行测试"
	@echo "  deps         下载依赖"
	@echo "  fmt          格式化代码"
	@echo "  deploy       部署到远程服务器 (需要参数 SERVER=ip USERNAME=user)"
	@echo ""
	@echo "示例:"
	@echo "  make build"
	@echo "  make run"
	@echo "  make import"
	@echo "  make deploy SERVER=192.168.1.100 USERNAME=root"

# 编译项目
build:
	@echo "编译主程序..."
	go build -o bin/gaokao-server main.go
	@echo "编译导入工具..."
	go build -o bin/import-tool tools/import_excel.go
	@echo "编译完成！"

# 运行服务器
run: build
	@echo "启动服务器..."
	./bin/gaokao-server

# 导入Excel数据
import: build
	@echo "导入Excel数据..."
	./bin/import-tool

# 下载依赖
deps:
	@echo "下载Go模块依赖..."
	go mod download
	go mod tidy

# 格式化代码
fmt:
	@echo "格式化Go代码..."
	go fmt ./...

# 运行测试
test:
	@echo "运行测试..."
	go test ./...

# 清理编译文件
clean:
	@echo "清理编译文件..."
	rm -rf bin/
	go clean

# 创建bin目录
bin:
	mkdir -p bin

# 编译时创建bin目录
build: bin

# 部署到远程服务器
deploy:
	@if [ -z "$(SERVER)" ] || [ -z "$(USERNAME)" ]; then \
		echo "错误: 需要指定服务器信息"; \
		echo "使用方法: make deploy SERVER=<IP> USERNAME=<用户名> [PORT=<端口>]"; \
		echo "例如: make deploy SERVER=192.168.1.100 USERNAME=root"; \
		echo "例如: make deploy SERVER=192.168.1.100 USERNAME=ubuntu PORT=2222"; \
		exit 1; \
	fi
	@echo "部署到服务器 $(USERNAME)@$(SERVER)..."
	chmod +x scripts/deploy.sh
	./scripts/deploy.sh $(SERVER) $(USERNAME) $(PORT)

# 本地开发环境设置
dev-setup:
	@echo "设置本地开发环境..."
	@echo "请确保已安装ClickHouse，如未安装请运行:"
	@echo "  sudo scripts/install_clickhouse.sh"
	@echo ""
	@echo "环境变量设置:"
	@echo "  export PORT=8031"
	@echo "  export CLICKHOUSE_HOST=localhost"
	@echo "  export CLICKHOUSE_PORT=9000"
	@echo "  export CLICKHOUSE_DATABASE=gaokao"

# 检查代码质量
lint:
	@echo "检查代码质量..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint 未安装，跳过代码检查"; \
		echo "安装命令: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# 生成API文档
docs:
	@echo "生成API文档..."
	@echo "API接口文档:"
	@echo ""
	@echo "1. 健康检查:"
	@echo "   GET /api/health"
	@echo ""
	@echo "2. 位次查询:"
	@echo "   GET /api/rank/get?score=555"
	@echo ""
	@echo "3. 报表查询:"
	@echo "   GET /api/report/get?rank=12000&class_comb=\"123\"&page=1&page_size=20" 