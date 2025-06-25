import pandas as pd
from clickhouse_driver import Client
import sys
from typing import Dict, List, Tuple

def connect_clickhouse():
    """连接到ClickHouse数据库"""
    try:
        client = Client(
            host='localhost',
            port=19000,
            user='default',
            password='',
        )
        return client
    except Exception as e:
        print(f"❌ 连接ClickHouse失败: {e}")
        return None

def check_table_exists(client, database: str, table: str) -> bool:
    """检查表是否存在"""
    try:
        result = client.execute(f"EXISTS TABLE {database}.{table}")
        return result[0][0] == 1
    except Exception as e:
        print(f"❌ 检查表存在性失败: {e}")
        return False

def add_study_years_column(client):
    """为admission_data表添加study_years字段"""
    try:
        # 检查字段是否已存在
        result = client.execute("DESCRIBE gaokao.admission_data")
        columns = [row[0] for row in result]
        
        if 'study_years' in columns:
            print("✅ study_years字段已存在")
            return True
        
        # 添加字段
        print("🔄 添加study_years字段...")
        client.execute("""
            ALTER TABLE gaokao.admission_data 
            ADD COLUMN study_years Nullable(String) DEFAULT NULL
        """)
        print("✅ study_years字段添加成功")
        return True
        
    except Exception as e:
        print(f"❌ 添加字段失败: {e}")
        return False

def load_study_years_data() -> Dict[int, str]:
    """从Excel文件加载学制数据"""
    print("🔄 读取学制数据...")
    
    try:
        # 读取新Excel文件
        df = pd.read_excel("最新最新21-24各省本科专业分 1.xlsx", header=1)
        
        # 提取有学制信息的记录
        has_study_year = df[df['学制'].notna()]
        
        # 创建ID到学制的映射
        study_mapping = {}
        for _, row in has_study_year.iterrows():
            study_id = int(row['id'])
            study_year = str(row['学制']).strip()
            study_mapping[study_id] = study_year
        
        print(f"✅ 成功加载 {len(study_mapping)} 条学制记录")
        return study_mapping
        
    except Exception as e:
        print(f"❌ 加载学制数据失败: {e}")
        return {}

def update_study_years_batch(client, id_study_mapping: Dict[int, str], batch_size: int = 1000):
    """批量更新学制数据"""
    
    items = list(id_study_mapping.items())
    total_batches = (len(items) + batch_size - 1) // batch_size
    
    print(f"🔄 开始批量更新，共 {len(items)} 条记录，分 {total_batches} 批处理...")
    
    success_count = 0
    error_count = 0
    
    for i in range(0, len(items), batch_size):
        batch = items[i:i + batch_size]
        batch_num = i // batch_size + 1
        
        try:
            # 构造批量更新语句
            cases = []
            ids = []
            for record_id, study_year in batch:
                cases.append(f"WHEN {record_id} THEN '{study_year}'")
                ids.append(str(record_id))
            
            if cases:
                update_sql = f"""
                ALTER TABLE gaokao.admission_data 
                UPDATE study_years = CASE id 
                    {' '.join(cases)}
                    ELSE study_years 
                END 
                WHERE id IN ({','.join(ids)})
                """
                
                client.execute(update_sql)
                success_count += len(batch)
                print(f"✅ 批次 {batch_num}/{total_batches} 完成 ({len(batch)} 条记录)")
            
        except Exception as e:
            print(f"❌ 批次 {batch_num} 更新失败: {e}")
            error_count += len(batch)
    
    print(f"📊 更新完成：成功 {success_count} 条，失败 {error_count} 条")
    return success_count, error_count

def verify_import(client) -> bool:
    """验证导入结果"""
    try:
        # 检查总记录数
        total_count = client.execute("SELECT count(*) FROM gaokao.admission_data")[0][0]
        
        # 检查有学制信息的记录数
        study_count = client.execute(
            "SELECT count(*) FROM gaokao.admission_data WHERE study_years IS NOT NULL"
        )[0][0]
        
        # 检查学制值分布
        study_distribution = client.execute("""
            SELECT study_years, count(*) as cnt 
            FROM gaokao.admission_data 
            WHERE study_years IS NOT NULL 
            GROUP BY study_years 
            ORDER BY cnt DESC 
            LIMIT 10
        """)
        
        print(f"\n📊 导入验证结果:")
        print(f"总记录数: {total_count:,}")
        print(f"有学制信息的记录: {study_count:,}")
        print(f"学制覆盖率: {study_count/total_count*100:.2f}%")
        
        print(f"\n学制分布 (前10):")
        for study_year, count in study_distribution:
            print(f"  {study_year}: {count:,} 条")
        
        return study_count > 0
        
    except Exception as e:
        print(f"❌ 验证失败: {e}")
        return False

def update_gitignore():
    """更新.gitignore文件，添加新的Excel文件"""
    try:
        gitignore_path = ".gitignore"
        excel_filename = "最新最新21-24各省本科专业分 1.xlsx"
        
        # 读取现有.gitignore
        try:
            with open(gitignore_path, 'r', encoding='utf-8') as f:
                content = f.read()
        except FileNotFoundError:
            content = ""
        
        # 检查是否已存在
        if excel_filename in content:
            print("✅ .gitignore已包含该Excel文件")
            return True
        
        # 添加到.gitignore
        with open(gitignore_path, 'a', encoding='utf-8') as f:
            if not content.endswith('\n'):
                f.write('\n')
            f.write(f"# 学制信息Excel文件\n")
            f.write(f"{excel_filename}\n")
        
        print("✅ 已将Excel文件添加到.gitignore")
        return True
        
    except Exception as e:
        print(f"❌ 更新.gitignore失败: {e}")
        return False

def main():
    """主执行函数"""
    
    print("=" * 60)
    print("🎯 学制数据导入工具")
    print("=" * 60)
    
    # 1. 连接数据库
    print("\n1️⃣ 连接ClickHouse数据库...")
    client = connect_clickhouse()
    if not client:
        return False
    
    # 2. 检查表是否存在
    print("\n2️⃣ 检查数据表...")
    if not check_table_exists(client, 'gaokao', 'admission_data'):
        print("❌ gaokao.admission_data表不存在")
        return False
    
    # 3. 添加study_years字段
    print("\n3️⃣ 添加study_years字段...")
    if not add_study_years_column(client):
        return False
    
    # 4. 加载学制数据
    print("\n4️⃣ 加载学制数据...")
    study_mapping = load_study_years_data()
    if not study_mapping:
        return False
    
    # 5. 批量更新数据
    print("\n5️⃣ 批量更新学制信息...")
    success_count, error_count = update_study_years_batch(client, study_mapping)
    
    if error_count > 0:
        print(f"⚠️  存在 {error_count} 条记录更新失败")
    
    # 6. 验证导入结果
    print("\n6️⃣ 验证导入结果...")
    if not verify_import(client):
        return False
    
    # 7. 更新.gitignore
    print("\n7️⃣ 更新.gitignore...")
    update_gitignore()
    
    print("\n" + "=" * 60)
    print("🎉 学制数据导入完成！")
    print("=" * 60)
    
    return True

if __name__ == "__main__":
    try:
        success = main()
        sys.exit(0 if success else 1)
    except KeyboardInterrupt:
        print("\n❌ 用户中断操作")
        sys.exit(1)
    except Exception as e:
        print(f"\n❌ 意外错误: {e}")
        sys.exit(1) 