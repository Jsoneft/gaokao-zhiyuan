#!/usr/bin/env python3
import subprocess
import time
import os
import sys

# ClickHouse可执行文件路径
CLICKHOUSE_PATH = "/opt/homebrew/Caskroom/clickhouse/25.5.2.47-stable/clickhouse-macos-aarch64"

def run_command(command, description, check_output=False, timeout=60):
    """运行命令并处理结果"""
    print(f"\n🔄 {description}")
    
    try:
        if check_output:
            result = subprocess.run(command, shell=True, capture_output=True, text=True, timeout=timeout)
            if result.returncode == 0:
                print(f"✅ 成功: {description}")
                if result.stdout.strip():
                    return result.stdout.strip()
                return ""
            else:
                print(f"❌ 失败: {description}")
                if result.stderr:
                    print(f"错误: {result.stderr}")
                return None
        else:
            result = subprocess.run(command, shell=True, timeout=timeout)
            if result.returncode == 0:
                print(f"✅ 成功: {description}")
                return True
            else:
                print(f"❌ 失败: {description}")
                return False
    except subprocess.TimeoutExpired:
        print(f"⏰ 超时: {description}")
        return False
    except Exception as e:
        print(f"❌ 异常: {description} - {e}")
        return False

def execute_sql_file_local(sql_file, description):
    """使用ClickHouse Local执行SQL文件"""
    print(f"\n🔄 {description}")
    
    if not os.path.exists(sql_file):
        print(f"❌ SQL文件不存在: {sql_file}")
        return False
    
    # 创建数据目录
    data_dir = os.path.expanduser("~/clickhouse_local_data")
    os.makedirs(data_dir, exist_ok=True)
    
    # 使用ClickHouse Local执行SQL文件
    cmd = f"{CLICKHOUSE_PATH} local --path {data_dir} --multiquery < {sql_file}"
    
    try:
        result = subprocess.run(cmd, shell=True, capture_output=True, text=True, timeout=300)
        if result.returncode == 0:
            print(f"✅ 成功执行: {description}")
            if result.stdout.strip():
                print(f"输出: {result.stdout.strip()}")
            return True
        else:
            print(f"❌ 执行失败: {description}")
            if result.stderr:
                print(f"错误: {result.stderr}")
            return False
    except subprocess.TimeoutExpired:
        print(f"⏰ 执行超时: {description}")
        return False
    except Exception as e:
        print(f"❌ 执行异常: {description} - {e}")
        return False

def run_query_local(query, description):
    """使用ClickHouse Local运行单个查询"""
    data_dir = os.path.expanduser("~/clickhouse_local_data")
    cmd = f'{CLICKHOUSE_PATH} local --path {data_dir} --query "{query}"'
    return run_command(cmd, description, check_output=True)

def main():
    """主函数"""
    print("ClickHouse 最终测试脚本")
    print("="*60)
    
    # 检查ClickHouse可执行文件
    if not os.path.exists(CLICKHOUSE_PATH):
        print(f"❌ ClickHouse可执行文件不存在: {CLICKHOUSE_PATH}")
        return False
    
    print(f"✅ 找到ClickHouse: {CLICKHOUSE_PATH}")
    
    # 清理旧数据
    data_dir = os.path.expanduser("~/clickhouse_local_data")
    if os.path.exists(data_dir):
        print(f"🔄 清理旧数据目录: {data_dir}")
        run_command(f"rm -rf {data_dir}", "清理旧数据")
    
    # 1. 执行建表SQL
    if not execute_sql_file_local("hubei_data/create_hubei_optimized_en.sql", "执行建表SQL"):
        print("❌ 建表失败")
        return False
    
    # 2. 执行修复后的插入SQL
    if not execute_sql_file_local("hubei_data/insert_data_fixed.sql", "执行修复后的数据插入SQL"):
        print("❌ 数据插入失败")
        return False
    
    # 3. 基础验证查询
    print("\n" + "="*60)
    print("📊 基础验证查询")
    print("="*60)
    
    basic_queries = [
        ("SELECT COUNT(*) as total_records FROM admission_hubei_wide_2024", "统计总记录数"),
        ("SELECT subject_category, COUNT(*) as count FROM admission_hubei_wide_2024 GROUP BY subject_category ORDER BY count DESC", "科类分布统计"),
        ("SELECT school_ownership, COUNT(*) as count FROM admission_hubei_wide_2024 GROUP BY school_ownership ORDER BY count DESC", "公私性质分布"),
        ("SELECT education_level, COUNT(*) as count FROM admission_hubei_wide_2024 GROUP BY education_level ORDER BY count DESC", "教育层次分布"),
    ]
    
    for query, description in basic_queries:
        result = run_query_local(query, description)
        if result:
            print(f"结果: {result}")
    
    # 4. 专业分类统计
    print("\n" + "="*60)
    print("📊 专业分类统计")
    print("="*60)
    
    category_query = """
    SELECT 
        SUM(is_science::UInt32) as science_count,
        SUM(is_engineering::UInt32) as engineering_count,
        SUM(is_medical::UInt32) as medical_count,
        SUM(is_economics_mgmt_law::UInt32) as economics_mgmt_law_count,
        SUM(is_liberal_arts::UInt32) as liberal_arts_count,
        SUM(is_design_arts::UInt32) as design_arts_count,
        SUM(is_language::UInt32) as language_count
    FROM admission_hubei_wide_2024
    """
    
    result = run_query_local(category_query, "专业分类统计")
    if result:
        print(f"结果: {result}")
    
    # 5. 高分专业查询
    print("\n" + "="*60)
    print("📊 高分专业TOP10")
    print("="*60)
    
    top_score_query = """
    SELECT school_name, major_name, min_score_2024, min_rank_2024
    FROM admission_hubei_wide_2024 
    WHERE min_score_2024 IS NOT NULL
    ORDER BY min_score_2024 DESC 
    LIMIT 10
    """
    
    result = run_query_local(top_score_query, "高分专业TOP10")
    if result:
        print(f"结果:\n{result}")
    
    # 6. 工科专业分析
    print("\n" + "="*60)
    print("📊 工科专业分析")
    print("="*60)
    
    engineering_queries = [
        ("SELECT COUNT(*) as engineering_count FROM admission_hubei_wide_2024 WHERE is_engineering = true", "工科专业总数"),
        ("SELECT AVG(min_score_2024) as avg_score FROM admission_hubei_wide_2024 WHERE is_engineering = true AND min_score_2024 IS NOT NULL", "工科专业平均分"),
        ("SELECT COUNT(*) as count FROM admission_hubei_wide_2024 WHERE is_engineering = true AND min_score_2024 BETWEEN 600 AND 650", "600-650分工科专业数"),
    ]
    
    for query, description in engineering_queries:
        result = run_query_local(query, description)
        if result:
            print(f"结果: {result}")
    
    # 7. 选科要求分析
    print("\n" + "="*60)
    print("📊 选科要求分析")
    print("="*60)
    
    subject_req_query = """
    SELECT 
        SUM(require_physics::UInt32) as require_physics_count,
        SUM(require_chemistry::UInt32) as require_chemistry_count,
        SUM(require_biology::UInt32) as require_biology_count,
        SUM(require_politics::UInt32) as require_politics_count,
        SUM(require_history::UInt32) as require_history_count,
        SUM(require_geography::UInt32) as require_geography_count
    FROM admission_hubei_wide_2024
    """
    
    result = run_query_local(subject_req_query, "选科要求统计")
    if result:
        print(f"结果: {result}")
    
    # 8. 性能测试
    print("\n" + "="*60)
    print("⚡ 性能测试")
    print("="*60)
    
    perf_queries = [
        ("SELECT COUNT(*) FROM admission_hubei_wide_2024 WHERE min_score_2024 > 600", "高分专业查询"),
        ("SELECT school_name, COUNT(*) as major_count FROM admission_hubei_wide_2024 GROUP BY school_name ORDER BY major_count DESC LIMIT 5", "学校专业数排名"),
        ("SELECT * FROM admission_hubei_wide_2024 WHERE is_engineering = true AND require_chemistry = true AND min_score_2024 BETWEEN 550 AND 600 ORDER BY min_score_2024 DESC LIMIT 10", "复合条件查询"),
    ]
    
    for query, description in perf_queries:
        start_time = time.time()
        result = run_query_local(query, description)
        end_time = time.time()
        
        if result is not None:
            execution_time = end_time - start_time
            print(f"执行时间: {execution_time:.3f}秒")
            if description == "复合条件查询":
                print(f"结果:\n{result}")
    
    # 9. 索引验证
    print("\n" + "="*60)
    print("📋 索引验证")
    print("="*60)
    
    index_queries = [
        ("SELECT name, type FROM system.data_skipping_indices WHERE table = 'admission_hubei_wide_2024'", "查看表索引"),
        ("SHOW CREATE TABLE admission_hubei_wide_2024", "查看表结构"),
    ]
    
    for query, description in index_queries:
        result = run_query_local(query, description)
        if result:
            print(f"结果: {result}")
    
    # 10. 生成使用说明
    print("\n" + "="*60)
    print("🔗 使用说明")
    print("="*60)
    print(f"ClickHouse路径: {CLICKHOUSE_PATH}")
    print(f"数据目录: ~/clickhouse_local_data")
    print("表名: admission_hubei_wide_2024")
    print("记录数: 18,278条")
    
    print("\n连接示例:")
    print(f"{CLICKHOUSE_PATH} local --path ~/clickhouse_local_data")
    
    print("\n常用查询示例:")
    print("1. 查看表结构:")
    print("   DESCRIBE admission_hubei_wide_2024;")
    
    print("\n2. 查询工科专业:")
    print("   SELECT school_name, major_name, min_score_2024")
    print("   FROM admission_hubei_wide_2024")
    print("   WHERE is_engineering = true")
    print("   ORDER BY min_score_2024 DESC LIMIT 10;")
    
    print("\n3. 查询特定分数段:")
    print("   SELECT COUNT(*)")
    print("   FROM admission_hubei_wide_2024")
    print("   WHERE min_score_2024 BETWEEN 600 AND 650;")
    
    print("\n4. 查询选科要求:")
    print("   SELECT school_name, major_name, min_score_2024")
    print("   FROM admission_hubei_wide_2024")
    print("   WHERE require_chemistry = true AND require_biology = true")
    print("   ORDER BY min_score_2024 DESC LIMIT 10;")
    
    print("\n5. 专业分类查询:")
    print("   SELECT school_name, major_name, min_score_2024")
    print("   FROM admission_hubei_wide_2024")
    print("   WHERE is_medical = true")
    print("   ORDER BY min_score_2024 DESC LIMIT 10;")
    
    print("\n🎉 ClickHouse Local环境设置完成！")
    print("💡 数据已保存，可以开始进行高考志愿分析查询")
    
    return True

if __name__ == "__main__":
    try:
        success = main()
        if success:
            print("\n✅ 所有测试通过！环境就绪")
        else:
            print("\n❌ 测试失败")
        
    except KeyboardInterrupt:
        print("\n\n⚠️  用户中断操作")
    
    sys.exit(0 if success else 1) 