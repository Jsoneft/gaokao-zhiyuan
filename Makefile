# 高考志愿填报系统 Makefile

.PHONY: build run clean test deps fmt lint help

# 默认目标
help:
	@echo "高考志愿填报系统 - 可用命令:"
	@echo ""
	@echo "  build        编译项目"
	@echo "  run          运行服务器"
	@echo "  clean        清理编译文件"
	@echo "  test         运行测试"
	@echo "  deps         下载依赖"
	@echo "  fmt          格式化代码"
	@echo "  lint         代码质量检查"
	@echo ""
	@echo "示例:"
	@echo "  make build"
	@echo "  make run"

# 创建bin目录
bin:
	mkdir -p bin

# 编译项目
build: bin
	@echo "编译主程序..."
	go build -o bin/gaokao-server main.go
	@echo "编译完成！"

# 运行服务器
run: build
	@echo "启动服务器..."
	./bin/gaokao-server

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

# 检查代码质量
lint:
	@echo "检查代码质量..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint 未安装，跳过代码检查"; \
		echo "安装命令: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# 开发环境设置说明
dev-setup:
	@echo "本地开发环境设置:"
	@echo ""
	@echo "1. 安装ClickHouse:"
	@echo "   - macOS: brew install clickhouse"
	@echo "   - Ubuntu: https://clickhouse.com/docs/en/install"
	@echo ""
	@echo "2. 环境变量设置:"
	@echo "   export PORT=8031"
	@echo "   export CLICKHOUSE_HOST=localhost"
	@echo "   export CLICKHOUSE_PORT=9000"
	@echo "   export CLICKHOUSE_DATABASE=gaokao"
	@echo "   export CLICKHOUSE_USERNAME=default"
	@echo "   export CLICKHOUSE_PASSWORD="
	@echo ""
	@echo "3. 启动ClickHouse并创建数据库:"
	@echo "   clickhouse-client --query \"CREATE DATABASE IF NOT EXISTS gaokao\"" 