import pandas as pd
import sys

def analyze_new_excel():
    """分析新Excel文件的结构和学制信息"""
    
    excel_file = "最新最新21-24各省本科专业分 1.xlsx"
    
    try:
        # 读取Excel文件
        print("正在读取Excel文件...")
        df = pd.read_excel(excel_file)
        
        print(f"文件行数: {len(df)}")
        print(f"文件列数: {len(df.columns)}")
        print("\n列名:")
        for i, col in enumerate(df.columns):
            print(f"{i+1}. {col}")
        
        print("\n前5行数据:")
        print(df.head())
        
        # 检查是否有id列
        if 'id' in df.columns:
            print(f"\nid列信息:")
            print(f"id唯一值数量: {df['id'].nunique()}")
            print(f"id总数量: {len(df['id'])}")
            print(f"是否有重复id: {df['id'].duplicated().any()}")
            print(f"id范围: {df['id'].min()} - {df['id'].max()}")
        
        # 寻找学制相关的列
        print("\n寻找学制相关的列:")
        study_related_cols = []
        for col in df.columns:
            if any(keyword in str(col).lower() for keyword in ['学制', 'study', 'year', '年制', '年']):
                study_related_cols.append(col)
        
        if study_related_cols:
            print(f"找到可能的学制相关列: {study_related_cols}")
            for col in study_related_cols:
                print(f"\n{col} 列统计:")
                print(f"唯一值数量: {df[col].nunique()}")
                print(f"唯一值: {df[col].unique()}")
                print(f"空值数量: {df[col].isnull().sum()}")
                print(f"值统计:")
                print(df[col].value_counts())
        else:
            print("未找到明显的学制相关列，显示所有列的唯一值数量:")
            for col in df.columns:
                print(f"{col}: {df[col].nunique()} 个唯一值")
        
        # 保存样本数据用于进一步分析
        sample_file = "new_excel_sample.csv"
        df.head(100).to_csv(sample_file, index=False, encoding='utf-8')
        print(f"\n已保存前100行样本数据到: {sample_file}")
        
        return df
        
    except Exception as e:
        print(f"读取Excel文件时出错: {e}")
        return None

def check_id_correspondence():
    """检查新Excel文件的id与现有数据的对应关系"""
    
    try:
        # 读取新Excel文件
        new_df = pd.read_excel("最新最新21-24各省本科专业分 1.xlsx")
        
        # 读取原始Excel文件进行对比
        original_df = pd.read_excel("21-24各省份录取数据(含专业组代码).xlsx")
        
        print("数据对应关系分析:")
        print(f"新文件记录数: {len(new_df)}")
        print(f"原文件记录数: {len(original_df)}")
        
        if 'id' in new_df.columns and 'id' in original_df.columns:
            new_ids = set(new_df['id'].unique())
            original_ids = set(original_df['id'].unique())
            
            print(f"新文件唯一id数: {len(new_ids)}")
            print(f"原文件唯一id数: {len(original_ids)}")
            
            # 检查交集
            common_ids = new_ids.intersection(original_ids)
            print(f"共同id数: {len(common_ids)}")
            
            # 检查差集
            only_in_new = new_ids - original_ids
            only_in_original = original_ids - new_ids
            
            print(f"仅在新文件中的id数: {len(only_in_new)}")
            print(f"仅在原文件中的id数: {len(only_in_original)}")
            
            if len(only_in_new) > 0:
                print(f"仅在新文件中的id示例: {list(only_in_new)[:10]}")
            if len(only_in_original) > 0:
                print(f"仅在原文件中的id示例: {list(only_in_original)[:10]}")
            
            # 检查是否有一对一的关系
            coverage_rate = len(common_ids) / len(original_ids) * 100
            print(f"id覆盖率: {coverage_rate:.2f}%")
            
            return {
                'new_df': new_df,
                'original_df': original_df,
                'common_ids': common_ids,
                'coverage_rate': coverage_rate,
                'can_merge': coverage_rate > 80  # 如果覆盖率超过80%认为可以合并
            }
        
    except Exception as e:
        print(f"检查id对应关系时出错: {e}")
        return None

if __name__ == "__main__":
    print("=== 分析新Excel文件 ===")
    df = analyze_new_excel()
    
    if df is not None:
        print("\n=== 检查数据对应关系 ===")
        result = check_id_correspondence()
        
        if result:
            print(f"\n=== 结论 ===")
            if result['can_merge']:
                print("✅ 数据可以合并，建议继续进行学制字段的导入")
            else:
                print("❌ 数据覆盖率不足，需要进一步检查数据一致性") 