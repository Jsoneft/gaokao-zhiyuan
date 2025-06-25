import pandas as pd

def analyze_study_years():
    """分析学制列的具体值"""
    
    # 读取新Excel文件，使用第2行作为header
    new_df = pd.read_excel("最新最新21-24各省本科专业分 1.xlsx", header=1)
    
    print("=== 学制列详细分析 ===")
    study_col = '学制'
    
    # 查看学制列的所有唯一值
    study_values = new_df[study_col].dropna().unique()
    print(f"学制列的所有唯一值 ({len(study_values)} 个):")
    for val in sorted(study_values):
        count = new_df[new_df[study_col] == val].shape[0]
        print(f"  {val}: {count} 条记录")
    
    # 检查与原文件的ID对应关系
    print("\n=== ID对应关系检查 ===")
    
    # 读取原始文件，也尝试不同的header设置
    try:
        original_df = pd.read_excel("21-24各省份录取数据(含专业组代码).xlsx")
        if 'id' not in original_df.columns:
            # 尝试使用header=1
            original_df = pd.read_excel("21-24各省份录取数据(含专业组代码).xlsx", header=1)
    except:
        print("❌ 无法读取原始Excel文件")
        return False, None
    
    print(f"原始文件列名: {list(original_df.columns)[:10]}")
    
    # 找到ID列
    id_col_original = None
    for col in original_df.columns:
        if 'id' in str(col).lower() or str(col).strip().lower() == 'id':
            id_col_original = col
            break
    
    if id_col_original is None:
        # 假设第一列是ID
        id_col_original = original_df.columns[0]
        print(f"假设原始文件第一列 '{id_col_original}' 是ID列")
    
    # ID对应关系
    new_ids = set(new_df['id'].dropna().astype(int))
    original_ids = set(original_df[id_col_original].dropna().astype(int))
    
    common_ids = new_ids.intersection(original_ids)
    only_in_new = new_ids - original_ids
    only_in_original = original_ids - new_ids
    
    print(f"新文件ID数量: {len(new_ids):,}")
    print(f"原文件ID数量: {len(original_ids):,}")
    print(f"共同ID数量: {len(common_ids):,}")
    print(f"仅在新文件中的ID: {len(only_in_new):,}")
    print(f"仅在原文件中的ID: {len(only_in_original):,}")
    print(f"ID覆盖率: {len(common_ids)/len(original_ids)*100:.2f}%")
    
    # 检查有学制信息的记录中，ID的覆盖情况
    has_study_year = new_df[new_df[study_col].notna()]
    study_ids = set(has_study_year['id'].astype(int))
    study_common_ids = study_ids.intersection(original_ids)
    
    print(f"\n=== 学制信息覆盖分析 ===")
    print(f"有学制信息的记录数: {len(has_study_year):,}")
    print(f"有学制信息的唯一ID数: {len(study_ids):,}")
    print(f"学制ID与原文件的重合数: {len(study_common_ids):,}")
    print(f"学制信息对原文件的覆盖率: {len(study_common_ids)/len(original_ids)*100:.2f}%")
    
    # 分析每个学制值的分布
    print(f"\n=== 学制分布分析 ===")
    study_counts = has_study_year[study_col].value_counts()
    for value, count in study_counts.head(10).items():  # 只显示前10个最常见的
        coverage = len(set(has_study_year[has_study_year[study_col] == value]['id'].astype(int)).intersection(original_ids))
        print(f"{value}: {count:,} 条记录, 覆盖原文件 {coverage:,} 个ID")
    
    # 检查是否存在一对多的情况（检查前100个共同ID以节省时间）
    print(f"\n=== 数据一致性检查 ===")
    
    conflicts = []
    test_ids = list(common_ids)[:100]  # 只检查前100个共同ID
    for common_id in test_ids:
        study_values_for_id = has_study_year[has_study_year['id'] == common_id][study_col].dropna().unique()
        if len(study_values_for_id) > 1:
            conflicts.append((common_id, study_values_for_id))
    
    if conflicts:
        print(f"发现 {len(conflicts)} 个ID有多个不同的学制值:")
        for cid, values in conflicts[:5]:  # 只显示前5个
            print(f"  ID {cid}: {values}")
    else:
        print("✅ 在检查的样本中没有发现ID对应多个不同学制值的情况")
    
    # 生成映射建议
    print(f"\n=== 映射建议 ===")
    coverage_rate = len(study_common_ids) / len(original_ids)
    if coverage_rate > 0.6:  # 降低阈值到60%
        print("✅ 建议进行学制字段的导入：")
        print(f"  1. 覆盖率为 {coverage_rate*100:.2f}%，数据质量良好")
        print("  2. 可以为ClickHouse表添加study_years字段")
        print("  3. 字段类型建议为String类型")
        print("  4. 主要学制值：四年、五年、三年等")
        
        return True, has_study_year, coverage_rate
    else:
        print("❌ 不建议导入：")
        print(f"  覆盖率只有 {coverage_rate*100:.2f}%，可能导致数据不完整")
        return False, None, coverage_rate

if __name__ == "__main__":
    success, study_data, coverage = analyze_study_years()
    
    if success:
        print(f"\n✅ 数据分析完成，覆盖率: {coverage*100:.2f}%")
        print("准备生成导入脚本...") 