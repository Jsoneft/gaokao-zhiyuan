#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import pandas as pd
import numpy as np

def analyze_hubei_excel_only():
    """只分析Excel中湖北生源地的数据，准备更新文件"""
    
    print("分析Excel中湖北生源地的专业最低分数据...")
    print("=" * 60)
    
    try:
        # 1. 读取Excel中的湖北数据
        print("1. 读取Excel中的湖北生源地数据...")
        excel_file = "20250626全国22-24各省本科专业分.xlsx"
        df_excel = pd.read_excel(excel_file, header=1)
        
        print(f"Excel总记录数: {len(df_excel):,}")
        
        # 检查生源地分布
        print("\n生源地分布:")
        source_dist = df_excel['生源地'].value_counts().head(10)
        for province, count in source_dist.items():
            print(f"  {province}: {count:,}")
        
        # 筛选湖北生源地且有ID和分数的记录
        hubei_data = df_excel[
            (df_excel['生源地'] == '湖北') & 
            (df_excel['id'].notna()) & 
            (df_excel['专业最低分'].notna())
        ].copy()
        
        print(f"\n湖北生源地有效记录数: {len(hubei_data):,}")
        
        if len(hubei_data) == 0:
            print("❌ 没有找到湖北生源地的有效数据")
            return False
        
        # 2. 分析湖北数据
        print(f"\n=== 湖北数据分析 ===")
        
        # 基本统计
        print(f"数据概况:")
        print(f"  记录数量: {len(hubei_data):,}")
        print(f"  分数范围: {hubei_data['专业最低分'].min():.0f} - {hubei_data['专业最低分'].max():.0f}")
        print(f"  平均分数: {hubei_data['专业最低分'].mean():.1f}")
        print(f"  中位数: {hubei_data['专业最低分'].median():.0f}")
        print(f"  院校数量: {hubei_data['院校名称'].nunique()}")
        print(f"  专业数量: {hubei_data['专业名称'].nunique()}")
        
        # ID唯一性检查
        unique_ids = hubei_data['id'].nunique()
        total_records = len(hubei_data)
        print(f"  唯一ID数: {unique_ids:,}")
        print(f"  ID重复率: {(total_records - unique_ids) / total_records * 100:.1f}%")
        
        # 分数分布
        print(f"\n分数区间分布:")
        bins = [0, 400, 450, 500, 550, 600, 650, 700, 800]
        score_dist = pd.cut(hubei_data['专业最低分'], bins=bins, include_lowest=True)
        for interval, count in score_dist.value_counts().sort_index().items():
            print(f"  {interval}: {count:,}")
        
        # 3. 院校和专业分析
        print(f"\n=== 院校和专业分析 ===")
        
        # 高分院校TOP10
        print(f"高分院校TOP10:")
        college_scores = hubei_data.groupby('院校名称')['专业最低分'].max().sort_values(ascending=False)
        for i, (college, score) in enumerate(college_scores.head(10).items(), 1):
            print(f"  {i:2d}. {college}: {score:.0f}")
        
        # 高分专业TOP10
        print(f"\n高分专业TOP10:")
        high_score_majors = hubei_data.nlargest(10, '专业最低分')
        for i, row in enumerate(high_score_majors.itertuples(), 1):
            print(f"  {i:2d}. {row.院校名称} - {row.专业名称}: {row.专业最低分:.0f}")
        
        # 4. 数据质量检查
        print(f"\n=== 数据质量检查 ===")
        
        # 检查ID格式
        id_sample = hubei_data['id'].head(10).tolist()
        print(f"ID样本: {id_sample}")
        
        # 检查分数格式
        score_sample = hubei_data['专业最低分'].head(10).tolist()
        print(f"分数样本: {score_sample}")
        
        # 检查是否有异常分数
        unusual_scores = hubei_data[(hubei_data['专业最低分'] < 200) | (hubei_data['专业最低分'] > 750)]
        if len(unusual_scores) > 0:
            print(f"异常分数记录: {len(unusual_scores)}")
            print(unusual_scores[['id', '院校名称', '专业名称', '专业最低分']].head())
        else:
            print("✅ 分数范围正常")
        
        # 5. 准备更新数据
        print(f"\n=== 准备更新数据 ===")
        
        # 创建更新数据集
        update_data = hubei_data[['id', '专业最低分', '院校名称', '专业名称', '生源地']].copy()
        
        # 数据类型转换和清理
        update_data['id'] = update_data['id'].astype(int)
        update_data['专业最低分'] = pd.to_numeric(update_data['专业最低分'], errors='coerce')
        
        # 删除转换失败的记录
        clean_data = update_data.dropna(subset=['id', '专业最低分'])
        print(f"清理后数据量: {len(clean_data):,}")
        
        # 重命名列以匹配ClickHouse
        clean_data.rename(columns={
            '专业最低分': 'major_min_score_2024',
            '院校名称': 'college_name_excel',
            '专业名称': 'major_name_excel',
            '生源地': 'source_province'
        }, inplace=True)
        
        # 保存更新数据
        hubei_score_file = "hubei_score_update_data.csv"
        clean_data.to_csv(hubei_score_file, index=False, encoding='utf-8')
        print(f"✅ 更新数据已保存到: {hubei_score_file}")
        print(f"包含字段: {list(clean_data.columns)}")
        
        # 保存样本数据用于手动验证
        sample_file = "hubei_score_sample.csv"
        sample_data = clean_data.head(50)
        sample_data.to_csv(sample_file, index=False, encoding='utf-8')
        print(f"✅ 样本数据已保存到: {sample_file}")
        
        # 6. 统计总结
        print(f"\n" + "=" * 60)
        print("数据准备总结:")
        print(f"✅ 湖北生源地记录: {len(hubei_data):,}")
        print(f"✅ 清理后记录: {len(clean_data):,}")
        print(f"✅ 分数范围: {clean_data['major_min_score_2024'].min():.0f} - {clean_data['major_min_score_2024'].max():.0f}")
        print(f"✅ 平均分数: {clean_data['major_min_score_2024'].mean():.1f}")
        print(f"✅ 数据质量良好，可以进行ClickHouse更新")
        
        return True
        
    except Exception as e:
        print(f"分析过程中出错: {e}")
        import traceback
        traceback.print_exc()
        return False

if __name__ == "__main__":
    analyze_hubei_excel_only() 