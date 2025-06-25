import pandas as pd
from clickhouse_driver import Client
import os

def perform_pre_import_checks():
    """执行学制数据导入前的全面检查"""
    
    print("🔍 学制数据导入前检查")
    print("=" * 50)
    
    checks_passed = 0
    total_checks = 7
    
    # 检查1: Excel文件存在性
    print("\n1️⃣ 检查Excel文件...")
    excel_file = "最新最新21-24各省本科专业分 1.xlsx"
    if os.path.exists(excel_file):
        file_size = os.path.getsize(excel_file) / (1024 * 1024)  # MB
        print(f"✅ Excel文件存在 ({file_size:.1f} MB)")
        checks_passed += 1
    else:
        print(f"❌ Excel文件不存在: {excel_file}")
    
    # 检查2: ClickHouse连接
    print("\n2️⃣ 检查ClickHouse连接...")
    try:
        client = Client(
            host='localhost',
            port=19000,
            user='default',
            password='',
        )
        version = client.execute("SELECT version()")[0][0]
        print(f"✅ ClickHouse连接成功 (版本: {version})")
        checks_passed += 1
    except Exception as e:
        print(f"❌ ClickHouse连接失败: {e}")
        client = None
    
    # 检查3: 数据库和表存在性
    print("\n3️⃣ 检查数据库和表...")
    if client:
        try:
            # 检查gaokao数据库
            databases = client.execute("SHOW DATABASES")
            database_names = [row[0] for row in databases]
            
            if 'gaokao' in database_names:
                print("✅ gaokao数据库存在")
                
                # 检查admission_data表
                tables = client.execute("SHOW TABLES FROM gaokao")
                table_names = [row[0] for row in tables]
                
                if 'admission_data' in table_names:
                    print("✅ admission_data表存在")
                    checks_passed += 1
                else:
                    print("❌ admission_data表不存在")
            else:
                print("❌ gaokao数据库不存在")
        except Exception as e:
            print(f"❌ 检查数据库表失败: {e}")
    
    # 检查4: 表结构
    print("\n4️⃣ 检查表结构...")
    if client:
        try:
            schema = client.execute("DESCRIBE gaokao.admission_data")
            columns = [row[0] for row in schema]
            
            required_columns = ['id', 'province', 'college_name', 'professional_name']
            missing_columns = [col for col in required_columns if col not in columns]
            
            if not missing_columns:
                print("✅ 表结构包含必要字段")
                if 'study_years' in columns:
                    print("⚠️  study_years字段已存在，将会被更新")
                else:
                    print("💡 study_years字段不存在，将会被添加")
                checks_passed += 1
            else:
                print(f"❌ 表结构缺少字段: {missing_columns}")
        except Exception as e:
            print(f"❌ 检查表结构失败: {e}")
    
    # 检查5: 当前数据量
    print("\n5️⃣ 检查当前数据量...")
    if client:
        try:
            count = client.execute("SELECT count(*) FROM gaokao.admission_data")[0][0]
            print(f"📊 当前记录数: {count:,}")
            
            if count > 0:
                print("✅ 表中有数据，可以进行学制字段更新")
                checks_passed += 1
            else:
                print("⚠️  表中无数据，请先导入基础数据")
        except Exception as e:
            print(f"❌ 检查数据量失败: {e}")
    
    # 检查6: 学制数据预分析
    print("\n6️⃣ 检查学制数据...")
    try:
        df = pd.read_excel(excel_file, header=1)
        
        total_records = len(df)
        has_study_info = df['学制'].notna().sum()
        unique_ids = df['id'].nunique()
        
        print(f"📊 Excel文件记录数: {total_records:,}")
        print(f"📊 有学制信息记录数: {has_study_info:,}")
        print(f"📊 唯一ID数: {unique_ids:,}")
        print(f"📊 学制覆盖率: {has_study_info/total_records*100:.2f}%")
        
        # 显示学制值分布
        study_values = df['学制'].value_counts().head(5)
        print("📊 主要学制值:")
        for value, count in study_values.items():
            print(f"   {value}: {count:,} 条")
        
        if has_study_info > 0:
            print("✅ 学制数据可用")
            checks_passed += 1
        else:
            print("❌ 无有效学制数据")
            
    except Exception as e:
        print(f"❌ 检查学制数据失败: {e}")
    
    # 检查7: 系统资源
    print("\n7️⃣ 检查系统资源...")
    try:
        import psutil
        
        # 内存检查
        memory = psutil.virtual_memory()
        memory_gb = memory.total / (1024**3)
        memory_available_gb = memory.available / (1024**3)
        
        print(f"💾 总内存: {memory_gb:.1f} GB")
        print(f"💾 可用内存: {memory_available_gb:.1f} GB")
        
        if memory_available_gb > 2:
            print("✅ 内存充足")
            checks_passed += 1
        else:
            print("⚠️  可用内存不足，可能影响处理速度")
            checks_passed += 1  # 仍然可以执行，只是速度慢
            
    except ImportError:
        print("💡 无法检查系统资源 (需要安装psutil)")
        checks_passed += 1  # 跳过此检查
    except Exception as e:
        print(f"⚠️  系统资源检查失败: {e}")
        checks_passed += 1  # 跳过此检查
    
    # 总结
    print("\n" + "=" * 50)
    print(f"📋 检查结果: {checks_passed}/{total_checks} 项通过")
    
    if checks_passed >= total_checks - 1:  # 允许1项检查失败
        print("✅ 系统准备就绪，可以执行导入")
        
        print("\n🚀 执行建议:")
        print("1. 确保有足够时间完成导入 (预计5-10分钟)")
        print("2. 导入过程中不要关闭程序")
        print("3. 如遇到错误，可以重新运行脚本")
        
        return True
    else:
        print("❌ 系统未准备就绪，请解决上述问题后重试")
        return False

def show_execution_plan():
    """显示执行计划"""
    print("\n" + "=" * 50)
    print("📋 执行计划")
    print("=" * 50)
    
    steps = [
        "1. 连接ClickHouse数据库",
        "2. 检查admission_data表存在性",
        "3. 为表添加study_years字段 (如果不存在)",
        "4. 从Excel读取学制数据",
        "5. 批量更新数据库记录",
        "6. 验证导入结果",
        "7. 更新.gitignore文件"
    ]
    
    for step in steps:
        print(f"   {step}")
    
    print("\n⚠️  注意事项:")
    print("   • 此操作会修改数据库表结构")
    print("   • 大约需要5-10分钟完成")
    print("   • 建议在非高峰时段执行")
    print("   • 确保有数据库备份 (如需要)")

if __name__ == "__main__":
    print("🔍 学制数据导入前置检查工具")
    print("=" * 50)
    
    # 执行检查
    ready = perform_pre_import_checks()
    
    # 显示执行计划
    if ready:
        show_execution_plan()
        
        print("\n" + "=" * 50)
        confirm = input("👆 确认要执行学制数据导入吗？(y/N): ").strip().lower()
        
        if confirm == 'y':
            print("\n🚀 请运行: python import_study_years.py")
        else:
            print("❌ 用户取消操作")
    else:
        print("\n�� 请解决上述问题后重新运行检查") 