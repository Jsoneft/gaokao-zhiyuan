#!/usr/bin/env python3
import subprocess
import os
import time
import sys

def run_command(command):
    """运行系统命令并打印输出"""
    print(f"执行命令: {command}")
    result = subprocess.run(command, shell=True, capture_output=True, text=True)
    if result.stdout:
        print(result.stdout)
    if result.stderr:
        print(f"错误: {result.stderr}")
    return result.returncode == 0

def create_database_schema():
    """创建ClickHouse数据库和表结构"""
    # 创建schema.sql文件
    schema_sql = """
    CREATE DATABASE IF NOT EXISTS gaokao;
    
    CREATE TABLE IF NOT EXISTS gaokao.admission_data (
        id UInt64,
        province String,
        batch String,
        subject_type String,
        class_demand String,
        college_code String,
        special_interest_group_code String,
        college_name String,
        professional_code String,
        professional_name String,
        description String,
        year UInt32,
        lowest_points Int64,
        lowest_rank Int64
    ) ENGINE = MergeTree()
    ORDER BY (lowest_rank, lowest_points, year, province)
    """
    
    with open("create_schema.sql", "w") as f:
        f.write(schema_sql)
    
    # 执行SQL创建数据库和表
    success = run_command("clickhouse client --multiquery < create_schema.sql")
    if not success:
        print("创建数据库和表结构失败!")
        return False
    
    print("数据库和表结构创建成功!")
    return True

def process_excel_data():
    """处理Excel数据，生成SQL文件"""
    if not os.path.exists("process_excel.py"):
        print("错误: 找不到Excel处理脚本!")
        return False
    
    success = run_command("python3 process_excel.py")
    if not success:
        print("处理Excel数据失败!")
        return False
    
    print("Excel数据处理成功!")
    return True

def import_data():
    """导入数据到ClickHouse"""
    # 检查data目录是否存在
    data_dir = "data"
    if not os.path.exists(data_dir):
        print(f"错误: 数据目录 {data_dir} 不存在!")
        return False
    
    # 获取所有2024年数据的SQL文件
    sql_files = [f for f in os.listdir(data_dir) if f.startswith("data_2024_") and f.endswith(".sql")]
    if not sql_files:
        print(f"错误: 在 {data_dir} 目录中没有找到2024年的SQL文件!")
        return False
    
    # 按批次排序
    sql_files.sort()
    
    for sql_file in sql_files:
        file_path = os.path.join(data_dir, sql_file)
        print(f"正在导入文件: {file_path}...")
        
        import_query = f"""
        clickhouse client --multiquery < {file_path}
        """
        
        success = run_command(import_query)
        if not success:
            print(f"导入 {sql_file} 失败!")
            return False
        
        print(f"{sql_file} 导入成功!")
    
    return True

def verify_data():
    """验证数据是否导入成功"""
    count_query = "SELECT year, count() FROM gaokao.admission_data GROUP BY year ORDER BY year"
    print("验证导入的数据...")
    run_command(f"clickhouse client --query \"{count_query}\"")
    
    return True

def main():
    # 检查ClickHouse是否运行
    print("检查ClickHouse服务状态...")
    if not run_command("clickhouse client --query \"SELECT 1\""):
        print("错误: 无法连接到ClickHouse服务器!")
        return 1
    
    # 清理旧数据
    print("清理旧数据...")
    run_command("rm -rf data create_schema.sql")
    run_command("clickhouse client --query \"DROP DATABASE IF EXISTS gaokao\"")
    
    # 创建目录
    os.makedirs("data", exist_ok=True)
    
    # 执行数据处理步骤
    steps = [
        ("创建数据库和表结构", create_database_schema),
        ("处理Excel数据", process_excel_data),
        ("导入数据", import_data),
        ("验证数据", verify_data)
    ]
    
    for step_name, step_func in steps:
        print(f"\n=== {step_name} ===")
        if not step_func():
            print(f"错误: {step_name}失败!")
            return 1
        print(f"{step_name}完成!")
    
    print("\n所有操作完成，2024年数据已成功导入到ClickHouse!")
    return 0

if __name__ == "__main__":
    sys.exit(main()) 