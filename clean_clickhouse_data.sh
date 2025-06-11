#!/bin/bash

# 清理脚本 - 用于删除冗余的ClickHouse数据文件
# 这些文件可以在需要时通过ClickHouse数据库重新生成

echo "开始清理冗余数据文件..."

# 删除冗余的批次SQL文件
if [ -d "data" ]; then
  find data -name "data_2024_batch_*.sql" -type f -delete
  echo "已删除data目录下的批次SQL文件"
fi

# 删除旧的/未使用的SQL数据文件
rm -f data_202*.sql
echo "已删除根目录下的年份SQL文件"

# 删除sql_chunks目录
if [ -d "sql_chunks" ]; then
  rm -rf sql_chunks
  echo "已删除sql_chunks目录"
fi

# 清理空的或未使用的工具文件
if [ -f "tools/create_db.go" ]; then
  rm -f tools/create_db.go
  echo "已删除create_db.go"
fi

if [ -f "tools/province_compare_data.go" ]; then
  rm -f tools/province_compare_data.go
  echo "已删除province_compare_data.go"
fi

# 清理空目录
for dir in "metadata_dropped" "format_schemas" "tmp"; do
  if [ -d "$dir" ]; then
    rmdir "$dir" 2>/dev/null
    if [ $? -eq 0 ]; then
      echo "已删除空目录: $dir"
    fi
  fi
done

# 清理重复的store目录下的数据文件
echo "注意: store目录下的数据是ClickHouse数据库的一部分，删除它们可能会影响数据库功能"
echo "如果确认要删除，请手动运行: rm -rf store"

echo "清理完成！"
echo "如果需要重新加载数据，请使用scripts/data目录下的SQL文件" 