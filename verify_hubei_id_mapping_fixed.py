#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import pandas as pd
import subprocess
import os
import tempfile

def verify_hubei_id_mapping_fixed():
    """验证Excel中湖北生源地的ID是否与ClickHouse数据匹配（使用clickhouse-client）"""
    
    print("验证湖北生源地ID映射关系...")
    print("=" * 60)
    
    try:
        # 1. 读取Excel中的湖北数据
        print("1. 读取Excel中的湖北生源地数据...")
        excel_file = "20250626全国22-24各省本科专业分.xlsx"
        df_excel = pd.read_excel(excel_file, header=1)
        
        # 筛选湖北生源地且有ID和分数的记录
        hubei_data = df_excel[
            (df_excel['生源地'] == '湖北') & 
            (df_excel['id'].notna()) & 
            (df_excel['专业最低分'].notna())
        ].copy()
        
        print(f"Excel中湖北生源地有效记录数: {len(hubei_data):,}")
        
        if len(hubei_data) == 0:
            print("❌ 没有找到湖北生源地的有效数据")
            return False
        
        # 取样本进行验证
        sample_size = min(100, len(hubei_data))
        sample_data = hubei_data.head(sample_size)
        sample_ids = sample_data['id'].tolist()
        print(f"湖北样本ID数量: {len(sample_ids)}")
        
        # 显示样本数据概况
        print(f"\n湖北数据概况:")
        print(f"  分数范围: {hubei_data['专业最低分'].min():.0f} - {hubei_data['专业最低分'].max():.0f}")
        print(f"  平均分数: {hubei_data['专业最低分'].mean():.1f}")
        print(f"  院校数量: {hubei_data['院校名称'].nunique()}")
        print(f"  专业数量: {hubei_data['专业名称'].nunique()}")
        
        # 2. 使用clickhouse-client查询
        print("\n2. 使用clickhouse-client查询ClickHouse...")
        
        # 检查表结构
        print("\n检查目标表结构...")
        structure_query = "DESCRIBE TABLE default.admission_hubei_wide_2024"
        
        try:
            result = subprocess.run([
                "clickhouse-client", 
                "--host", "43.248.188.28",
                "--port", "42914", 
                "--user", "default",
                "--database", "default",
                "--query", structure_query
            ], capture_output=True, text=True, timeout=30)
            
            if result.returncode == 0:
                print("表结构:")
                lines = result.stdout.strip().split('\n')
                for line in lines[:10]:  # 只显示前10个字段
                    print(f"  {line}")
                if len(lines) > 10:
                    print(f"  ... 还有 {len(lines)-10} 个字段")
            else:
                print(f"❌ 查询表结构失败: {result.stderr}")
                return False
        except Exception as e:
            print(f"❌ 连接ClickHouse失败: {e}")
            return False
        
        # 3. 查询匹配的ID
        print(f"\n3. 查询ClickHouse中的匹配ID...")
        id_list_str = ','.join([str(int(id_val)) for id_val in sample_ids])
        
        ch_query = f"""
        SELECT id, college_name, major_name, min_score_2024, min_rank_2024
        FROM default.admission_hubei_wide_2024 
        WHERE id IN ({id_list_str})
        ORDER BY id
        LIMIT 50
        """
        
        try:
            result = subprocess.run([
                "clickhouse-client", 
                "--host", "43.248.188.28",
                "--port", "42914", 
                "--user", "default",
                "--database", "default",
                "--query", ch_query
            ], capture_output=True, text=True, timeout=60)
            
            if result.returncode == 0:
                ch_lines = result.stdout.strip().split('\n')
                ch_rows = []
                for line in ch_lines:
                    if line.strip():
                        parts = line.split('\t')
                        if len(parts) >= 5:
                            try:
                                id_val = int(parts[0])
                                ch_rows.append((id_val, parts[1], parts[2], parts[3], parts[4]))
                            except:
                                continue
                
                print(f"ClickHouse中匹配的记录数: {len(ch_rows)}")
                
                if len(ch_rows) > 0:
                    print("\n前10条匹配记录:")
                    for i, row in enumerate(ch_rows[:10], 1):
                        print(f"  {i:2d}. ID: {row[0]}, 院校: {row[1]}, 专业: {row[2]}, 分数: {row[3]}")
                
            else:
                print(f"❌ 查询匹配ID失败: {result.stderr}")
                return False
        except Exception as e:
            print(f"❌ 查询ClickHouse失败: {e}")
            return False
        
        # 4. 对比分析
        print(f"\n4. 对比分析...")
        
        # 创建ClickHouse ID映射
        ch_id_map = {row[0]: row for row in ch_rows}
        ch_ids = set(ch_id_map.keys())
        excel_sample_ids = set([int(id_val) for id_val in sample_ids])
        
        matched_ids = ch_ids.intersection(excel_sample_ids)
        unmatched_excel_ids = excel_sample_ids - ch_ids
        
        print(f"Excel湖北样本ID数: {len(excel_sample_ids)}")
        print(f"ClickHouse匹配ID数: {len(ch_ids)}")
        print(f"成功匹配的ID数: {len(matched_ids)}")
        print(f"Excel中未匹配的ID数: {len(unmatched_excel_ids)}")
        
        if len(excel_sample_ids) > 0:
            match_rate = len(matched_ids) / len(excel_sample_ids) * 100
            print(f"匹配率: {match_rate:.1f}%")
        else:
            match_rate = 0
        
        # 5. 详细对比匹配的记录
        if len(matched_ids) > 0:
            print(f"\n5. 详细对比前10条匹配记录...")
            
            for i, match_id in enumerate(list(matched_ids)[:10], 1):
                # Excel数据
                excel_row = sample_data[sample_data['id'] == match_id].iloc[0]
                # ClickHouse数据
                ch_row = ch_id_map[match_id]
                
                print(f"\n{i}. ID: {match_id}")
                print(f"   生源地:   {excel_row['生源地']}")
                print(f"   Excel院校: {excel_row['院校名称']}")
                print(f"   CH院校:   {ch_row[1]}")
                print(f"   Excel专业: {excel_row['专业名称']}")
                print(f"   CH专业:   {ch_row[2]}")
                print(f"   Excel分数: {excel_row['专业最低分']}")
                print(f"   CH分数:   {ch_row[3]}")
        
        # 6. 检查是否有重复的ID（笛卡尔积检查）
        print(f"\n6. 笛卡尔积检查...")
        
        duplicate_check_query = f"""
        SELECT id, COUNT(*) as count
        FROM default.admission_hubei_wide_2024 
        WHERE id IN ({id_list_str})
        GROUP BY id
        HAVING count > 1
        ORDER BY count DESC
        """
        
        try:
            result = subprocess.run([
                "clickhouse-client", 
                "--host", "43.248.188.28",
                "--port", "42914", 
                "--user", "default",
                "--database", "default",
                "--query", duplicate_check_query
            ], capture_output=True, text=True, timeout=30)
            
            if result.returncode == 0:
                duplicate_lines = result.stdout.strip().split('\n')
                duplicate_rows = []
                for line in duplicate_lines:
                    if line.strip():
                        parts = line.split('\t')
                        if len(parts) >= 2:
                            try:
                                id_val = int(parts[0])
                                count_val = int(parts[1])
                                duplicate_rows.append((id_val, count_val))
                            except:
                                continue
                
                if len(duplicate_rows) > 0:
                    print(f"发现重复ID数量: {len(duplicate_rows)}")
                    print("重复ID详情:")
                    for row in duplicate_rows[:5]:
                        print(f"  ID: {row[0]}, 重复次数: {row[1]}")
                else:
                    print("✅ 没有发现重复ID，无笛卡尔积问题")
            else:
                print(f"重复检查查询失败: {result.stderr}")
                duplicate_rows = []
        except Exception as e:
            print(f"重复检查失败: {e}")
            duplicate_rows = []
        
        # 7. 总结和建议
        print(f"\n" + "=" * 60)
        print("验证总结:")
        print(f"✅ Excel湖北数据: {len(hubei_data):,} 条记录")
        print(f"✅ ID匹配率: {match_rate:.1f}% ({len(matched_ids)}/{len(excel_sample_ids)})")
        print(f"✅ 无笛卡尔积问题" if len(duplicate_rows) == 0 else f"⚠️  发现 {len(duplicate_rows)} 个重复ID")
        
        if len(matched_ids) > 0 and match_rate > 50:
            print(f"✅ 匹配率可接受，可以进行数据更新")
            
            # 保存匹配的湖北数据用于后续更新
            hubei_score_file = "hubei_score_update_verified.csv"
            update_data = hubei_data[['id', '专业最低分', '院校名称', '专业名称', '生源地']].copy()
            update_data.rename(columns={'专业最低分': 'major_min_score_2024'}, inplace=True)
            update_data.to_csv(hubei_score_file, index=False, encoding='utf-8')
            print(f"✅ 验证后的更新数据已保存到: {hubei_score_file}")
            
            return True
        else:
            print(f"❌ 匹配率过低（{match_rate:.1f}%），需要检查数据源")
            return False
        
    except Exception as e:
        print(f"验证过程中出错: {e}")
        import traceback
        traceback.print_exc()
        return False

if __name__ == "__main__":
    verify_hubei_id_mapping_fixed() 