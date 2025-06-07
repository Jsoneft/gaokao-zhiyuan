#!/usr/bin/env python3

def split_sql_file(input_file, output_dir, batch_size=10000):
    """将大SQL文件分割成多个较小的SQL文件"""
    print(f"分割SQL文件: {input_file}")
    
    with open(input_file, 'r') as f:
        # 读取INSERT语句（第一行）
        insert_statement = f.readline().strip()
        
        batch_num = 1
        batch_lines = []
        total_lines = 0
        
        for line in f:
            line = line.strip()
            if not line:
                continue
                
            batch_lines.append(line)
            
            # 如果达到了批次大小，写入文件
            if len(batch_lines) >= batch_size:
                write_batch(output_dir, batch_num, insert_statement, batch_lines)
                batch_num += 1
                total_lines += len(batch_lines)
                print(f"已写入批次 {batch_num-1}, 总计 {total_lines} 行")
                batch_lines = []
        
        # 写入最后一批
        if batch_lines:
            write_batch(output_dir, batch_num, insert_statement, batch_lines)
            total_lines += len(batch_lines)
            print(f"已写入批次 {batch_num}, 总计 {total_lines} 行")
            
    print(f"分割完成! 共 {batch_num} 个文件, {total_lines} 行数据")

def write_batch(output_dir, batch_num, insert_statement, lines):
    """将一批数据写入文件"""
    output_file = f"{output_dir}/batch_{batch_num}.sql"
    with open(output_file, 'w') as f:
        f.write(insert_statement + "\n")
        
        # 最后一行不需要逗号
        for i, line in enumerate(lines):
            if i == len(lines) - 1 and line.endswith(','):
                line = line[:-1]
            f.write(line + "\n")
        
        # 添加结束分号
        if not lines[-1].endswith(';'):
            f.write(";\n")

if __name__ == "__main__":
    split_sql_file("../data_2024_new.sql", "../sql_chunks", 10000) 