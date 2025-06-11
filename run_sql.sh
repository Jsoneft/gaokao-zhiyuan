#!/bin/bash

# 远程ClickHouse服务器参数
CH_HOST="43.248.188.28"
CH_PORT="26890"
CH_USER="default"
CH_PASSWORD="vfdeuiclgb"
CH_DATABASE="gaokao"

# 执行SQL查询
echo "按省份统计高考数据条数："
clickhouse-client --host="$CH_HOST" --port="$CH_PORT" --user="$CH_USER" --password="$CH_PASSWORD" --database="$CH_DATABASE" --query="
SELECT 
    province, 
    COUNT(*) as count,
    ROUND(COUNT(*) / (SELECT COUNT(*) FROM admission_data) * 100, 2) as percentage
FROM admission_data 
GROUP BY province 
ORDER BY count DESC
FORMAT PrettyCompact
"

# 查询总记录数
echo -e "\n总记录数："
clickhouse-client --host="$CH_HOST" --port="$CH_PORT" --user="$CH_USER" --password="$CH_PASSWORD" --database="$CH_DATABASE" --query="
SELECT COUNT(*) as total_count FROM admission_data
FORMAT PrettyCompact
" 