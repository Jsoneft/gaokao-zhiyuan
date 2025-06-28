#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import pandas as pd
import clickhouse_connect
import numpy as np

def verify_id_mapping():
    """验证Excel中的ID是否与ClickHouse数据匹配"""
    
    print("验证ID映射关系...")
    print("=" * 60)
    
    try:
        # 1. 读取Excel中的ID样本
        print("1. 读取Excel样本数据...")
        excel_file = "20250626全国22-24各省本科专业分.xlsx"
        df_excel = pd.read_excel(excel_file, header=1)
        
        # 筛选有ID和分数的记录
        valid_excel = df_excel[(df_excel['id'].notna()) & (df_excel['专业最低分'].notna())].copy()
        print(f"Excel中有效记录数: {len(valid_excel):,}")
        
        # 取样本进行验证
        sample_ids = valid_excel['id'].head(100).tolist()
        print(f"样本ID数量: {len(sample_ids)}")
        
        # 2. 连接ClickHouse
        print("\n2. 连接ClickHouse...")
        client = clickhouse_connect.get_client(
            host='43.248.188.28',
            port=42914,
            username='default',
            password='',
            database='default'
        )
        
        # 3. 检查目标表结构
        print("\n3. 检查目标表结构...")
        table_info_query = "DESCRIBE TABLE default.admission_hubei_wide_2024"
        table_structure = client.query(table_info_query)
        print("表结构:")
        for row in table_structure.result_rows:
            print(f"  {row[0]:25} {row[1]:15} {row[2] if len(row) > 2 else ''}")
        
        # 4. 查询ClickHouse中的ID
        print(f"\n4. 查询ClickHouse中的匹配ID...")
        id_list_str = ','.join([str(int(id_val)) for id_val in sample_ids])
        
        ch_query = f"""
        SELECT id, college_name, major_name, min_score_2024, min_rank_2024
        FROM default.admission_hubei_wide_2024 
        WHERE id IN ({id_list_str})
        ORDER BY id
        LIMIT 50
        """
        
        ch_result = client.query(ch_query)
        ch_rows = ch_result.result_rows
        
        print(f"ClickHouse中匹配的记录数: {len(ch_rows)}")
        
        if len(ch_rows) > 0:
            print("\n前10条匹配记录:")
            for i, row in enumerate(ch_rows[:10], 1):
                print(f"  {i:2d}. ID: {row[0]}, 院校: {row[1]}, 专业: {row[2]}, 分数: {row[3]}")
        
        # 5. 对比分析
        print(f"\n5. 对比分析...")
        
        # 创建ClickHouse ID映射
        ch_id_map = {row[0]: row for row in ch_rows}
        ch_ids = set(ch_id_map.keys())
        excel_sample_ids = set([int(id_val) for id_val in sample_ids])
        
        matched_ids = ch_ids.intersection(excel_sample_ids)
        unmatched_excel_ids = excel_sample_ids - ch_ids
        
        print(f"Excel样本ID数: {len(excel_sample_ids)}")
        print(f"ClickHouse匹配ID数: {len(ch_ids)}")
        print(f"成功匹配的ID数: {len(matched_ids)}")
        print(f"Excel中未匹配的ID数: {len(unmatched_excel_ids)}")
        print(f"匹配率: {len(matched_ids)/len(excel_sample_ids)*100:.1f}%")
        
        # 6. 详细对比匹配的记录
        if len(matched_ids) > 0:
            print(f"\n6. 详细对比前10条匹配记录...")
            
            matched_comparison = []
            for i, match_id in enumerate(list(matched_ids)[:10], 1):
                # Excel数据
                excel_row = valid_excel[valid_excel['id'] == match_id].iloc[0]
                # ClickHouse数据
                ch_row = ch_id_map[match_id]
                
                comparison = {
                    'id': match_id,
                    'excel_college': excel_row['院校名称'],
                    'ch_college': ch_row[1],
                    'excel_major': excel_row['专业名称'],
                    'ch_major': ch_row[2],
                    'excel_score': excel_row['专业最低分'],
                    'ch_score': ch_row[3]
                }
                matched_comparison.append(comparison)
                
                print(f"\n{i}. ID: {match_id}")
                print(f"   Excel院校: {comparison['excel_college']}")
                print(f"   CH院校:   {comparison['ch_college']}")
                print(f"   Excel专业: {comparison['excel_major']}")
                print(f"   CH专业:   {comparison['ch_major']}")
                print(f"   Excel分数: {comparison['excel_score']}")
                print(f"   CH分数:   {comparison['ch_score']}")
        
        # 7. 检查是否有重复的ID（笛卡尔积检查）
        print(f"\n7. 笛卡尔积检查...")
        
        duplicate_check_query = f"""
        SELECT id, COUNT(*) as count
        FROM default.admission_hubei_wide_2024 
        WHERE id IN ({id_list_str})
        GROUP BY id
        HAVING count > 1
        ORDER BY count DESC
        """
        
        duplicate_result = client.query(duplicate_check_query)
        duplicate_rows = duplicate_result.result_rows
        
        if len(duplicate_rows) > 0:
            print(f"发现重复ID数量: {len(duplicate_rows)}")
            print("重复ID详情:")
            for row in duplicate_rows[:5]:
                print(f"  ID: {row[0]}, 重复次数: {row[1]}")
        else:
            print("✅ 没有发现重复ID，无笛卡尔积问题")
        
        # 8. 总结
        print(f"\n" + "=" * 60)
        print("验证总结:")
        print(f"✅ Excel文件包含 {len(valid_excel):,} 条有效记录")
        print(f"✅ ID匹配率: {len(matched_ids)/len(excel_sample_ids)*100:.1f}% ({len(matched_ids)}/{len(excel_sample_ids)})")
        print(f"✅ 无笛卡尔积问题" if len(duplicate_rows) == 0 else f"⚠️  发现 {len(duplicate_rows)} 个重复ID")
        
        if len(matched_ids) > 0:
            print(f"✅ 数据结构匹配，可以进行更新")
            return True
        else:
            print(f"❌ 没有匹配的ID，需要检查数据源")
            return False
        
    except Exception as e:
        print(f"验证过程中出错: {e}")
        import traceback
        traceback.print_exc()
        return False

if __name__ == "__main__":
    verify_id_mapping() 