# 项目清理说明

本文档记录了高考志愿填报系统项目的清理过程和结果。

## 清理内容

本次清理主要删除了以下冗余文件：

1. **重复的SQL批处理文件**
   - 删除了 `/data/data_2024_batch_*.sql` 文件，因为它们与 `/scripts/data` 目录下的文件内容相同
   - 这些数据已经存储在 ClickHouse 数据库中，并且仍有备份在 `/scripts/data` 目录

2. **旧的根目录SQL文件**
   - 删除了根目录下的 `data_2021.sql`、`data_2022.sql`、`data_2023.sql`、`data_2024.sql` 等文件
   - 这些文件内容都已经分批存储在 `/scripts/data` 目录下

3. **冗余的 SQL 片段目录**
   - 删除了 `/sql_chunks` 目录，该目录内容与 `/scripts/data` 目录重复

4. **空的或未完成的工具文件**
   - 删除了 `tools/create_db.go` 和 `tools/province_compare_data.go` 等空文件或未完成的工具文件

5. **空目录**
   - 删除了 `metadata_dropped`、`format_schemas`、`tmp` 等空目录

## 项目结构说明

清理后的项目主要保留以下关键部分：

1. **数据文件**
   - `/scripts/data` 目录：包含所有批次的SQL数据文件，是原始数据的主要存储位置
   - `setup_clickhouse.sql`：ClickHouse数据库初始化脚本

2. **Go工具脚本**
   - `tools/province_source_stats.go`：按省份统计数据分布
   - `tools/hubei_fixed.go`：湖北省详细数据分析
   - `tools/clickhouse_stats.go`：ClickHouse数据库统计工具
   - 其他数据分析和查询工具

3. **清理脚本**
   - `clean_clickhouse_data.sh`：用于清理冗余数据文件的脚本

## 数据库连接信息

项目使用远程ClickHouse数据库，连接信息如下：
- 服务器：43.248.188.28:26890
- 数据库：gaokao
- 用户名：default
- 密码：vfdeuiclgb

## 注意事项

1. 如需重新导入数据，可使用 `/scripts/data` 目录下的SQL文件
2. 本次清理不影响系统功能，只删除了冗余和重复的文件
3. `/store` 目录下的文件是ClickHouse数据库的一部分，不建议手动删除 