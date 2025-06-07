#!/usr/bin/env python3
import pandas as pd
import os
import re
import math

def clean_text(text):
    """清理文本，移除特殊字符并转义引号"""
    if pd.isna(text):
        return ""
    text = str(text).strip()
    # 移除换行符和制表符
    text = re.sub(r'[\r\n\t]', ' ', text)
    # 转义单引号
    text = text.replace("'", "''")
    return text

def main():
    # 读取Excel文件
    print("正在读取Excel文件...")
    file_path = "21-24各省份录取数据(含专业组代码).xlsx"
    df = pd.read_excel(file_path)
    
    # 显示前20列的列名
    for i in range(20):
        if i < len(df.columns):
            print(f"列 {i}: {df.columns[i]}")
    
    # 获取第一行的数据，这是实际的字段名
    print("\n第一行数据（实际字段名）:")
    first_row = df.iloc[0]
    for i in range(20):
        if i < len(first_row):
            print(f"列 {i}: {first_row[i]}")
    
    # 重新创建DataFrame，跳过第一行（原始列名是第一行）
    df = pd.read_excel(file_path, skiprows=1)
    
    # 确定2024年录取最低分和最低位次的列索引
    lowest_points_2024_col = 15  # 2024年录取数据 - 录取最低分
    lowest_rank_2024_col = 16    # Unnamed: 16 - 录取最低位次
    
    # 准备要处理的列
    cols_to_process = {
        0: "id",                  # 2024年计划数据
        1: "province",            # 生源地
        2: "batch",               # 批次
        3: "subject_type",        # 科类
        4: "class_demand",        # 选科限制
        5: "college_code",        # 院校代码
        6: "special_interest_group_code",  # 专业组代码
        7: "college_name",        # 院校名称
        8: "professional_code",   # 专业代码
        9: "professional_name",   # 专业名称
        10: "description",        # 专业备注
        lowest_points_2024_col: "lowest_points",  # 录取最低分
        lowest_rank_2024_col: "lowest_rank",      # 录取最低位次
    }
    
    # 创建新的DataFrame，只保留需要的列
    df_filtered = pd.DataFrame()
    for col_idx, new_col_name in cols_to_process.items():
        if col_idx < len(df.columns):
            df_filtered[new_col_name] = df.iloc[:, col_idx]
    
    # 添加年份列并填充2024
    df_filtered['year'] = 2024
    
    # 确保数值列为整数类型
    df_filtered['id'] = df_filtered['id'].fillna(0).astype(int)
    
    # 处理lowest_points和lowest_rank
    # 将NaN值转换为0
    df_filtered['lowest_points'] = df_filtered['lowest_points'].fillna(0).astype(int)
    df_filtered['lowest_rank'] = df_filtered['lowest_rank'].fillna(0).astype(int)
    
    # 清理文本列数据
    for col in df_filtered.columns:
        if df_filtered[col].dtype == 'object':
            df_filtered[col] = df_filtered[col].apply(lambda x: clean_text(x))
    
    # 创建data目录
    os.makedirs("data", exist_ok=True)
    
    # 生成SQL插入语句
    total_rows = len(df_filtered)
    batch_size = 1000
    num_batches = math.ceil(total_rows / batch_size)
    
    for batch_idx in range(num_batches):
        start_idx = batch_idx * batch_size
        end_idx = min((batch_idx + 1) * batch_size, total_rows)
        batch_df = df_filtered.iloc[start_idx:end_idx]
        
        # 创建SQL文件
        sql_file = f"data/data_2024_batch_{batch_idx + 1}.sql"
        with open(sql_file, 'w') as f:
            f.write("INSERT INTO gaokao.admission_data (id, province, batch, subject_type, class_demand, college_code, special_interest_group_code, college_name, professional_code, professional_name, description, year, lowest_points, lowest_rank) VALUES\n")
            
            rows = []
            for _, row in batch_df.iterrows():
                values = (
                    f"{row['id']}",
                    f"'{row['province']}'",
                    f"'{row['batch']}'",
                    f"'{row['subject_type']}'",
                    f"'{row['class_demand']}'",
                    f"'{row['college_code']}'",
                    f"'{row['special_interest_group_code']}'",
                    f"'{row['college_name']}'",
                    f"'{row['professional_code']}'",
                    f"'{row['professional_name']}'",
                    f"'{row['description']}'",
                    f"{row['year']}",
                    f"{row['lowest_points']}",
                    f"{row['lowest_rank']}"
                )
                rows.append("(" + ", ".join(values) + ")")
            
            f.write(",\n".join(rows))
            f.write(";\n")
        
        print(f"已生成SQL文件: {sql_file}")
    
    print("数据处理完成!")

if __name__ == "__main__":
    main() 