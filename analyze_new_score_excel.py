#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import pandas as pd
import sys
import numpy as np
from datetime import datetime

def analyze_new_score_excel():
    """分析新Excel文件中的24年专业最低分数据"""
    
    print("开始分析 20250626全国22-24各省本科专业分.xlsx...")
    print(f"分析时间: {datetime.now()}")
    print("=" * 60)
    
    try:
        excel_file = "20250626全国22-24各省本科专业分.xlsx"
        print(f"正在读取文件: {excel_file}")
        
        # 尝试不同的header行，因为可能有多行表头
        for header_row in [0, 1, 2, 3]:
            print(f"\n=== 尝试使用第{header_row+1}行作为表头 ===")
            try:
                df_test = pd.read_excel(excel_file, header=header_row, nrows=5)
                print(f"列数: {len(df_test.columns)}")
                print("前几列名:")
                for i, col in enumerate(df_test.columns[:10], 1):
                    print(f"  {i:2d}. '{col}'")
                
                # 查找包含"最低分"或"专业最低分"的列
                score_columns = []
                for col in df_test.columns:
                    col_str = str(col).lower()
                    if any(keyword in col_str for keyword in ['最低分', '专业最低分', '24年', '2024']):
                        if '最低分' in col_str:
                            score_columns.append(col)
                
                if score_columns:
                    print(f"找到可能的分数列: {score_columns}")
                    selected_header = header_row
                    break
                    
            except Exception as e:
                print(f"第{header_row+1}行作为表头时出错: {e}")
                continue
        else:
            print("未找到合适的表头行，使用默认方式")
            selected_header = 1
        
        # 使用选定的表头读取完整数据
        print(f"\n=== 使用第{selected_header+1}行作为表头读取完整数据 ===")
        df = pd.read_excel(excel_file, header=selected_header)
        print(f"总行数: {len(df):,}")
        print(f"总列数: {len(df.columns)}")
        
        # 显示所有列名
        print(f"\n所有列名:")
        for i, col in enumerate(df.columns, 1):
            print(f"  {i:2d}. '{col}'")
        
        # 查找24年专业最低分相关列
        score_columns = []
        id_columns = []
        
        for col in df.columns:
            col_str = str(col).lower()
            # 查找分数相关列
            if any(keyword in col_str for keyword in ['最低分', '专业最低分']):
                if any(year in col_str for year in ['24', '2024']):
                    score_columns.append(col)
            # 查找ID列
            if 'id' in col_str or 'ID' in str(col):
                id_columns.append(col)
        
        print(f"\n找到的分数相关列: {score_columns}")
        print(f"找到的ID相关列: {id_columns}")
        
        if not score_columns:
            # 如果没找到，显示包含"分"的所有列
            all_score_related = [col for col in df.columns if '分' in str(col)]
            print(f"\n所有包含'分'的列: {all_score_related}")
            
            # 显示包含"24"或"2024"的列
            year_related = [col for col in df.columns if '24' in str(col) or '2024' in str(col)]
            print(f"所有包含'24'或'2024'的列: {year_related}")
            
            # 手动查看前几行数据
            print(f"\n前5行数据样本:")
            sample_cols = df.columns[:15] if len(df.columns) >= 15 else df.columns
            print(df[sample_cols].head())
            
            return
        
        # 分析找到的分数列
        score_column = score_columns[0]
        print(f"\n=== 分析分数列: '{score_column}' ===")
        
        # 基本统计
        total_rows = len(df)
        non_null_scores = df[score_column].notna().sum()
        null_scores = df[score_column].isna().sum()
        
        print(f"总记录数: {total_rows:,}")
        print(f"非空分数: {non_null_scores:,} ({non_null_scores/total_rows*100:.1f}%)")
        print(f"空值数量: {null_scores:,} ({null_scores/total_rows*100:.1f}%)")
        
        if non_null_scores > 0:
            score_data = df[score_column].dropna()
            print(f"\n分数数据分析:")
            print(f"数据类型: {score_data.dtype}")
            
            # 查看样本数据
            print(f"前10个分数样本:")
            for i, score in enumerate(score_data.head(10), 1):
                print(f"  {i:2d}. {score}")
            
            # 尝试数值转换
            try:
                numeric_scores = pd.to_numeric(score_data, errors='coerce')
                valid_numeric = numeric_scores.notna().sum()
                print(f"\n可转换为数值的记录: {valid_numeric:,}")
                
                if valid_numeric > 0:
                    print(f"分数范围: {numeric_scores.min():.0f} - {numeric_scores.max():.0f}")
                    print(f"平均分数: {numeric_scores.mean():.1f}")
            except Exception as e:
                print(f"数值转换错误: {e}")
        
        # 分析ID列
        if id_columns:
            id_column = id_columns[0]
            print(f"\n=== 分析ID列: '{id_column}' ===")
            
            id_data = df[id_column].dropna()
            print(f"非空ID数: {len(id_data):,}")
            print(f"唯一ID数: {id_data.nunique():,}")
            
            # ID样本
            print(f"前10个ID样本:")
            for i, id_val in enumerate(id_data.head(10), 1):
                print(f"  {i:2d}. {id_val}")
        
        # 保存样本数据
        if score_columns and (id_columns or len(df.columns) > 0):
            print(f"\n=== 保存样本数据 ===")
            
            # 选择关键列
            key_columns = []
            if id_columns:
                key_columns.append(id_columns[0])
            key_columns.append(score_columns[0])
            
            # 添加其他有用列
            for col in df.columns:
                if any(keyword in str(col) for keyword in ['院校', '专业', '省份', '科类', '学校']):
                    if col not in key_columns:
                        key_columns.append(col)
            
            # 限制列数
            key_columns = key_columns[:10]
            
            sample_data = df[key_columns].head(100)
            sample_file = "new_score_sample.csv"
            sample_data.to_csv(sample_file, index=False, encoding='utf-8')
            print(f"样本数据已保存到: {sample_file}")
            print(f"包含列: {list(sample_data.columns)}")
        
        print(f"\n分析完成!")
        return True
        
    except Exception as e:
        print(f"分析过程中出错: {e}")
        import traceback
        traceback.print_exc()
        return False

if __name__ == "__main__":
    analyze_new_score_excel() 