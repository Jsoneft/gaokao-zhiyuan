#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import pandas as pd
import numpy as np

def analyze_score_columns():
    """分析各个分数列，确定哪个是2024年专业最低分"""
    
    print("分析各个分数列...")
    print("=" * 60)
    
    try:
        # 读取Excel文件
        excel_file = "20250626全国22-24各省本科专业分.xlsx"
        df = pd.read_excel(excel_file, header=1)  # 使用第2行作为表头
        
        print(f"总记录数: {len(df):,}")
        
        # 所有分数相关列
        score_columns = ['专业组最低分', '专业最低分', '专业组最低分.1', '最低分', '最低分.1', '最低分.2']
        
        print(f"\n=== 分析各个分数列 ===")
        
        for i, col in enumerate(score_columns, 1):
            print(f"\n{i}. 分析列: '{col}'")
            
            if col in df.columns:
                col_data = df[col]
                non_null_count = col_data.notna().sum()
                null_count = col_data.isna().sum()
                
                print(f"   非空数量: {non_null_count:,} ({non_null_count/len(df)*100:.1f}%)")
                print(f"   空值数量: {null_count:,} ({null_count/len(df)*100:.1f}%)")
                
                if non_null_count > 0:
                    # 查看数据样本
                    sample_data = col_data.dropna().head(10)
                    print(f"   样本数据: {list(sample_data)}")
                    
                    # 尝试数值转换
                    try:
                        numeric_data = pd.to_numeric(col_data, errors='coerce')
                        valid_numeric = numeric_data.notna().sum()
                        if valid_numeric > 0:
                            print(f"   数值范围: {numeric_data.min():.0f} - {numeric_data.max():.0f}")
                            print(f"   平均值: {numeric_data.mean():.1f}")
                    except:
                        pass
            else:
                print(f"   列不存在")
        
        # 根据列位置分析可能的年份对应关系
        print(f"\n=== 根据列位置推测年份对应关系 ===")
        
        # 从列名和位置推测：
        # - 专业组最低分 (15) 和 专业最低分 (17) 可能是2024年
        # - 专业组最低分.1 (19) 可能是2023年  
        # - 最低分 (22) 可能是2022年
        # - 最低分.1 (25) 可能是2021年
        
        year_mapping = {
            '专业最低分': '2024年',
            '专业组最低分': '2024年',
            '专业组最低分.1': '2023年',
            '最低分': '2022年',
            '最低分.1': '2021年',
            '最低分.2': '可能更早年份'
        }
        
        for col, year in year_mapping.items():
            if col in df.columns:
                non_null_count = df[col].notna().sum()
                print(f"{col:15} -> {year:10} (非空: {non_null_count:,})")
        
        # 重点分析专业最低分（很可能是2024年的）
        print(f"\n=== 重点分析 '专业最低分' 列（推测为2024年数据）===")
        
        if '专业最低分' in df.columns:
            score_col = '专业最低分'
            id_col = 'id'
            
            # 筛选有效数据
            valid_data = df[(df[id_col].notna()) & (df[score_col].notna())].copy()
            print(f"有ID且有分数的记录: {len(valid_data):,}")
            
            if len(valid_data) > 0:
                # 分析分数分布
                scores = pd.to_numeric(valid_data[score_col], errors='coerce')
                valid_scores = scores.dropna()
                
                print(f"有效分数记录: {len(valid_scores):,}")
                if len(valid_scores) > 0:
                    print(f"分数范围: {valid_scores.min():.0f} - {valid_scores.max():.0f}")
                    print(f"平均分数: {valid_scores.mean():.1f}")
                    print(f"中位数: {valid_scores.median():.0f}")
                    
                    # 分数区间分布
                    print(f"\n分数区间分布:")
                    bins = [0, 300, 400, 500, 600, 700, 800]
                    score_dist = pd.cut(valid_scores, bins=bins, include_lowest=True)
                    print(score_dist.value_counts().sort_index())
                
                # 保存样本数据用于验证
                sample_file = "score_validation_sample.csv"
                sample_data = valid_data[['id', score_col, '院校名称', '专业名称', '生源地']].copy()
                sample_data.rename(columns={score_col: 'score_2024'}, inplace=True)
                sample_data.head(200).to_csv(sample_file, index=False, encoding='utf-8')
                print(f"\n验证样本已保存到: {sample_file}")
        
        print(f"\n分析完成!")
        return True
        
    except Exception as e:
        print(f"分析过程中出错: {e}")
        import traceback
        traceback.print_exc()
        return False

if __name__ == "__main__":
    analyze_score_columns() 