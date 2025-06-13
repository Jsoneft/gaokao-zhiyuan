#!/usr/bin/env python3
"""
ClickHouse 服务器启动脚本
用于启动湖北省高考志愿数据库服务
"""

import subprocess
import time
import sys
import os

# ClickHouse可执行文件路径
CLICKHOUSE_PATH = "/opt/homebrew/Caskroom/clickhouse/25.5.2.47-stable/clickhouse-macos-aarch64"
CONFIG_FILE = "/Users/jarviszuo/clickhouse_server/config.xml"

def start_server():
    """启动ClickHouse服务器"""
    print("🔄 启动ClickHouse服务器...")
    
    # 停止可能存在的进程
    subprocess.run("pkill -f clickhouse-server", shell=True, capture_output=True)
    time.sleep(2)
    
    # 启动服务器
    server_cmd = f"{CLICKHOUSE_PATH} server --config-file={CONFIG_FILE} --daemon"
    
    try:
        result = subprocess.run(server_cmd, shell=True, capture_output=True, text=True)
        if result.returncode == 0:
            print("✅ ClickHouse服务器启动成功")
            
            # 等待服务器就绪
            print("⏳ 等待服务器就绪...")
            for i in range(30):
                time.sleep(1)
                test_cmd = f"{CLICKHOUSE_PATH} client --port 19000 --query 'SELECT 1'"
                test_result = subprocess.run(test_cmd, shell=True, capture_output=True)
                if test_result.returncode == 0:
                    print("✅ 服务器已就绪")
                    return True
            
            print("❌ 服务器启动超时")
            return False
        else:
            print(f"❌ 服务器启动失败: {result.stderr}")
            return False
    except Exception as e:
        print(f"❌ 启动异常: {e}")
        return False

def test_connection():
    """测试连接"""
    print("🔄 测试连接...")
    
    test_queries = [
        ("SELECT COUNT(*) FROM admission_hubei_wide_2024", "数据记录数"),
        ("SELECT version()", "ClickHouse版本"),
    ]
    
    for query, desc in test_queries:
        cmd = f'{CLICKHOUSE_PATH} client --port 19000 --query "{query}"'
        result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
        if result.returncode == 0:
            print(f"✅ {desc}: {result.stdout.strip()}")
        else:
            print(f"❌ {desc}失败: {result.stderr}")

def main():
    """主函数"""
    print("湖北省高考志愿数据库启动脚本")
    print("="*50)
    
    # 检查ClickHouse可执行文件
    if not os.path.exists(CLICKHOUSE_PATH):
        print(f"❌ ClickHouse可执行文件不存在: {CLICKHOUSE_PATH}")
        return False
    
    # 检查配置文件
    if not os.path.exists(CONFIG_FILE):
        print(f"❌ 配置文件不存在: {CONFIG_FILE}")
        return False
    
    # 启动服务器
    if not start_server():
        return False
    
    # 测试连接
    test_connection()
    
    # 显示连接信息
    print("\n" + "="*50)
    print("🔗 数据库连接信息")
    print("="*50)
    print("主机: localhost")
    print("TCP端口: 19000 (命令行客户端)")
    print("HTTP端口: 18123 (DB工具连接)")
    print("用户名: default")
    print("密码: (空)")
    print("数据库: default")
    print("表名: admission_hubei_wide_2024")
    
    print("\n🎉 数据库服务已启动，可以开始使用！")
    print("💡 停止服务器: pkill -f clickhouse-server")
    
    return True

if __name__ == "__main__":
    try:
        success = main()
        sys.exit(0 if success else 1)
    except KeyboardInterrupt:
        print("\n\n⚠️  用户中断操作")
        print("停止ClickHouse服务器...")
        subprocess.run("pkill -f clickhouse-server", shell=True)
        sys.exit(0) 