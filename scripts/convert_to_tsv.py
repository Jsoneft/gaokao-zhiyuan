#!/usr/bin/env python3
import re

def process_sql_to_tsv(input_file, output_file):
    """处理SQL文件并转换为TSV格式"""
    print(f"处理文件: {input_file} -> {output_file}")
    count = 0
    
    with open(input_file, 'r') as f, open(output_file, 'w') as out:
        # 跳过第一行（INSERT INTO语句）
        next(f)
        
        for line in f:
            line = line.strip()
            if not line or line == ';':
                continue
                
            # 去掉结尾的逗号和分号
            if line.endswith(','):
                line = line[:-1]
            if line.endswith(';'):
                line = line[:-1]
                
            # 提取括号内的内容
            match = re.match(r'\((.*)\)', line)
            if not match:
                continue
                
            values = match.group(1)
            
            # 解析CSV格式（处理引号内的逗号）
            parts = []
            current = ""
            in_quotes = False
            
            for char in values:
                if char == "'" and (not current or current[-1] != '\\'):
                    in_quotes = not in_quotes
                    
                if char == ',' and not in_quotes:
                    parts.append(current.strip())
                    current = ""
                else:
                    current += char
                    
            if current:
                parts.append(current.strip())
                
            # 清理每个字段
            cleaned = []
            for part in parts:
                if part.startswith("'") and part.endswith("'"):
                    # 去掉引号
                    cleaned.append(part[1:-1].replace("''", "'"))
                else:
                    cleaned.append(part)
            
            # 写入TSV行
            out.write("\t".join(cleaned) + "\n")
            count += 1
            
            if count % 10000 == 0:
                print(f"已处理 {count} 行...")
    
    print(f"转换完成! 共处理 {count} 行")

if __name__ == "__main__":
    process_sql_to_tsv("../data_2024_new.sql", "../data_2024.tsv") 