# 项目清理文档

## 已清理的敏感信息

本文档记录了项目中已清理的敏感信息，确保代码安全。

## 清理内容

### 1. Shell脚本文件
- 已删除所有 .sh 文件，这些文件包含了部署脚本和密码信息

### 2. 硬编码密码
- 已将所有硬编码的数据库密码改为环境变量读取
- 已将所有硬编码的SSH密码移除

### 3. 配置文件
- 确保所有配置都通过环境变量管理
- 移除了配置文件中的明文密码

## 安全建议

1. **环境变量**: 使用环境变量管理所有敏感信息
2. **访问控制**: 确保数据库和服务器有适当的访问控制
3. **密码策略**: 使用强密码并定期更换
4. **网络安全**: 配置防火墙和网络访问控制

## 环境变量配置

```bash
# ClickHouse 配置
export CLICKHOUSE_HOST=localhost
export CLICKHOUSE_PORT=19000
export CLICKHOUSE_USERNAME=default
export CLICKHOUSE_PASSWORD=your_secure_password
export CLICKHOUSE_DATABASE=gaokao

# 服务配置
export PORT=8031
export GIN_MODE=release
```

## 注意事项

- 不要在代码中硬编码任何密码或敏感信息
- 使用 .env 文件进行本地开发配置
- 确保 .env 文件已添加到 .gitignore
- 生产环境使用系统环境变量或安全的配置管理工具

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
- 密码：已移除

## 注意事项

1. 如需重新导入数据，可使用 `/scripts/data` 目录下的SQL文件
2. 本次清理不影响系统功能，只删除了冗余和重复的文件
3. `/store` 目录下的文件是ClickHouse数据库的一部分，不建议手动删除 