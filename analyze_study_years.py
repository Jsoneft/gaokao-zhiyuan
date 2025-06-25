import pandas as pd
import numpy as np

def analyze_excel_structure():
    """深入分析Excel文件的结构，处理复杂的头部"""
    
    excel_file = "最新最新21-24各省本科专业分 1.xlsx"
    
    try:
        # 首先查看原始数据的前几行，不跳过任何行
        print("=== 查看Excel文件的原始结构 ===")
        raw_df = pd.read_excel(excel_file, header=None, nrows=10)
        print("前10行原始数据:")
        for i in range(min(10, len(raw_df))):
            print(f"第{i+1}行: {list(raw_df.iloc[i].values)}")
        
        # 尝试不同的header设置
        print("\n=== 尝试使用第2行作为header ===")
        df_header1 = pd.read_excel(excel_file, header=1, nrows=5)
        print(f"列数: {len(df_header1.columns)}")
        print("列名:")
        for i, col in enumerate(df_header1.columns):
            print(f"{i+1}. {col}")
        
        print("\n前5行数据:")
        print(df_header1.head())
        
        # 查找包含"学制"、"年制"等关键词的列
        print("\n=== 查找学制相关列 ===")
        study_cols = []
        for i, col in enumerate(df_header1.columns):
            col_str = str(col).lower()
            if any(keyword in col_str for keyword in ['学制', '年制', 'study', 'year', '学制年限', '修业年限']):
                study_cols.append((i, col))
        
        if study_cols:
            print(f"找到学制相关列: {study_cols}")
        else:
            print("在列名中未找到明显的学制相关列")
            # 检查数据内容
            print("\n检查前100行数据中是否包含学制信息...")
            for col_idx, col_name in enumerate(df_header1.columns):
                sample_values = df_header1[col_name].dropna().astype(str).head(20).tolist()
                # 检查值中是否包含学制相关信息
                has_study_info = any(
                    any(keyword in str(val).lower() for keyword in ['年', '学制', 'year'])
                    for val in sample_values
                    if str(val) not in ['nan', '']
                )
                if has_study_info:
                    print(f"列 {col_idx+1} ({col_name}) 可能包含学制信息:")
                    print(f"样本值: {sample_values}")
        
        return df_header1
        
    except Exception as e:
        print(f"分析Excel文件时出错: {e}")
        return None

def check_id_and_study_years():
    """检查ID对应关系和学制信息"""
    
    try:
        # 使用正确的header读取新文件
        new_df = pd.read_excel("最新最新21-24各省本科专业分 1.xlsx", header=1)
        
        print(f"\n=== 新文件基本信息 ===")
        print(f"行数: {len(new_df)}")
        print(f"列数: {len(new_df.columns)}")
        
        # 查找ID列
        id_col = None
        for col in new_df.columns:
            if 'id' in str(col).lower() or str(col).strip() == 'id':
                id_col = col
                break
        
        if id_col is None:
            # 检查第一列是否是ID
            first_col = new_df.columns[0]
            first_col_values = new_df[first_col].dropna()
            if len(first_col_values) > 0 and str(first_col_values.iloc[0]).isdigit():
                id_col = first_col
                print(f"假设第一列 '{first_col}' 是ID列")
        
        if id_col:
            print(f"找到ID列: {id_col}")
            print(f"ID唯一值数量: {new_df[id_col].nunique()}")
            print(f"ID范围: {new_df[id_col].min()} - {new_df[id_col].max()}")
            
            # 检查与原文件的对应关系
            original_df = pd.read_excel("21-24各省份录取数据(含专业组代码).xlsx")
            if 'id' in original_df.columns:
                new_ids = set(new_df[id_col].dropna().astype(int))
                original_ids = set(original_df['id'].dropna().astype(int))
                
                common_ids = new_ids.intersection(original_ids)
                coverage = len(common_ids) / len(original_ids) * 100
                
                print(f"\n=== ID对应关系 ===")
                print(f"新文件ID数: {len(new_ids)}")
                print(f"原文件ID数: {len(original_ids)}")
                print(f"共同ID数: {len(common_ids)}")
                print(f"覆盖率: {coverage:.2f}%")
                
                if coverage > 80:
                    print("✅ ID覆盖率良好，可以进行数据合并")
                else:
                    print("❌ ID覆盖率不足，需要进一步检查")
        
        # 详细检查所有列，寻找学制信息
        print(f"\n=== 详细列分析 ===")
        for i, col in enumerate(new_df.columns):
            print(f"\n列 {i+1}: {col}")
            non_null_count = new_df[col].notna().sum()
            unique_count = new_df[col].nunique()
            print(f"  非空值数量: {non_null_count}")
            print(f"  唯一值数量: {unique_count}")
            
            if non_null_count > 0:
                sample_values = new_df[col].dropna().head(10).tolist()
                print(f"  样本值: {sample_values}")
                
                # 检查是否可能是学制信息
                if unique_count < 20 and non_null_count > 1000:  # 学制通常是少数几个值
                    all_values = new_df[col].dropna().unique()
                    print(f"  所有唯一值: {all_values}")
                    
                    # 检查是否包含数字年制信息
                    year_pattern_found = any(
                        str(val) in ['1', '2', '3', '4', '5', '6', '7', '8'] or
                        '年' in str(val) or
                        'year' in str(val).lower()
                        for val in all_values
                    )
                    
                    if year_pattern_found:
                        print(f"  🎯 可能是学制列！")
        
        return new_df
        
    except Exception as e:
        print(f"检查ID和学制信息时出错: {e}")
        return None

def create_mapping_analysis():
    """创建详细的映射分析"""
    
    print("\n=== 创建详细映射分析 ===")
    
    try:
        # 读取两个文件
        new_df = pd.read_excel("最新最新21-24各省本科专业分 1.xlsx", header=1)
        original_df = pd.read_excel("21-24各省份录取数据(含专业组代码).xlsx")
        
        # 保存详细分析结果
        with open("study_years_analysis.txt", "w", encoding="utf-8") as f:
            f.write("学制信息分析报告\n")
            f.write("="*50 + "\n\n")
            
            f.write(f"新文件行数: {len(new_df)}\n")
            f.write(f"新文件列数: {len(new_df.columns)}\n")
            f.write(f"原文件行数: {len(original_df)}\n")
            f.write(f"原文件列数: {len(original_df.columns)}\n\n")
            
            f.write("新文件列信息:\n")
            for i, col in enumerate(new_df.columns):
                non_null = new_df[col].notna().sum()
                unique = new_df[col].nunique()
                f.write(f"{i+1:2d}. {col:30s} 非空:{non_null:8d} 唯一:{unique:6d}\n")
            
            # 查找最可能的学制列
            potential_study_cols = []
            for i, col in enumerate(new_df.columns):
                unique_count = new_df[col].nunique()
                non_null_count = new_df[col].notna().sum()
                
                if 2 <= unique_count <= 15 and non_null_count > 10000:
                    unique_vals = new_df[col].dropna().unique()
                    # 检查是否包含年制信息
                    has_year_info = any(
                        str(val).strip() in ['1', '2', '3', '4', '5', '6', '7', '8'] or
                        '年' in str(val) or
                        'year' in str(val).lower()
                        for val in unique_vals
                    )
                    
                    if has_year_info:
                        potential_study_cols.append((i, col, unique_vals))
            
            f.write(f"\n潜在的学制列:\n")
            for idx, col, vals in potential_study_cols:
                f.write(f"{idx+1}. {col}: {vals}\n")
        
        print("详细分析已保存到 study_years_analysis.txt")
        
        if potential_study_cols:
            print(f"\n找到 {len(potential_study_cols)} 个潜在的学制列:")
            for idx, col, vals in potential_study_cols:
                print(f"  列 {idx+1}: {col}")
                print(f"    唯一值: {vals}")
        
        return potential_study_cols
        
    except Exception as e:
        print(f"创建映射分析时出错: {e}")
        return []

if __name__ == "__main__":
    print("开始分析新Excel文件的学制信息...")
    
    # 步骤1: 分析Excel结构
    df = analyze_excel_structure()
    
    if df is not None:
        # 步骤2: 检查ID对应关系和学制信息
        new_df = check_id_and_study_years()
        
        # 步骤3: 创建详细映射分析
        potential_cols = create_mapping_analysis()
        
        if potential_cols:
            print(f"\n=== 结论 ===")
            print(f"✅ 发现 {len(potential_cols)} 个可能的学制列")
            print("建议进行下一步的数据导入测试")
        else:
            print(f"\n=== 结论 ===")
            print("❌ 未找到明确的学制列，需要进一步分析") 