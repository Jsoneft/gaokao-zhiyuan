# 湖北省高考志愿数据库文档

## 📊 项目概述

本项目为湖北省2024年高考志愿填报提供数据分析支持，包含18,278条本科专业录取数据。数据库基于ClickHouse构建，支持高性能的多维度查询分析。

## 🗄️ 数据库信息

### 连接配置
- **数据库类型**: ClickHouse 25.5.2.47
- **主机**: localhost
- **TCP端口**: 19000 (命令行客户端)
- **HTTP端口**: 18123 (DB工具连接)
- **用户名**: default
- **密码**: (空)
- **数据库**: default

### 表信息
- **表名**: `admission_hubei_wide_2024`
- **记录数**: 18,278条
- **数据范围**: 湖北省2024年本科专业录取数据
- **更新时间**: 2024年6月

## 📋 表结构

### 基本信息字段
| 字段名 | 类型 | 描述 | 示例 |
|--------|------|------|------|
| `id` | String | 唯一标识符 | "42_1001_01_物理" |
| `province_code` | String | 省份代码 | "42" |
| `school_code` | String | 院校代码 | "1001" |
| `school_name` | String | 院校名称 | "清华大学" |
| `major_code` | String | 专业代码 | "01" |
| `major_name` | String | 专业名称 | "理科试验班类" |
| `subject_category` | Enum8 | 科类 | '物理'=1, '历史'=2 |
| `group_code` | String | 专业组代码 | "01" |

### 选科要求字段
| 字段名 | 类型 | 描述 |
|--------|------|------|
| `require_physics` | Bool | 是否要求物理 |
| `require_chemistry` | Bool | 是否要求化学 |
| `require_biology` | Bool | 是否要求生物 |
| `require_politics` | Bool | 是否要求政治 |
| `require_history` | Bool | 是否要求历史 |
| `require_geography` | Bool | 是否要求地理 |

### 录取数据字段
| 字段名 | 类型 | 描述 | 示例 |
|--------|------|------|------|
| `min_score_2024` | UInt16 | 2024年最低分 | 692 |
| `min_rank_2024` | UInt32 | 2024年最低位次 | 156 |
| `plan_count_2024` | UInt16 | 2024年招生计划数 | 30 |

### 专业分类标签
| 字段名 | 类型 | 描述 |
|--------|------|------|
| `is_science` | Bool | 是否为理科类 |
| `is_engineering` | Bool | 是否为工科类 |
| `is_medical` | Bool | 是否为医科类 |
| `is_economics_mgmt_law` | Bool | 是否为经管法类 |
| `is_liberal_arts` | Bool | 是否为文科类 |
| `is_design_arts` | Bool | 是否为设计艺术类 |
| `is_language` | Bool | 是否为语言类 |

### 其他字段
| 字段名 | 类型 | 描述 |
|--------|------|------|
| `school_level` | String | 院校层次 |
| `school_nature` | String | 院校性质 |
| `school_location` | String | 院校所在地 |
| `major_duration` | String | 学制 |
| `tuition_fee` | String | 学费 |
| `remarks` | String | 备注信息 |

## 🔍 常用查询示例

### 1. 基础查询
```sql
-- 查看表结构
DESCRIBE admission_hubei_wide_2024;

-- 统计总记录数
SELECT COUNT(*) FROM admission_hubei_wide_2024;

-- 查看数据样本
SELECT * FROM admission_hubei_wide_2024 LIMIT 5;
```

### 2. 分数段分析
```sql
-- 查看高分专业（600分以上）
SELECT 
    school_name, 
    major_name, 
    min_score_2024, 
    min_rank_2024
FROM admission_hubei_wide_2024 
WHERE min_score_2024 >= 600 
ORDER BY min_score_2024 DESC 
LIMIT 20;

-- 分数段分布统计
SELECT 
    CASE 
        WHEN min_score_2024 >= 650 THEN '650+'
        WHEN min_score_2024 >= 600 THEN '600-649'
        WHEN min_score_2024 >= 550 THEN '550-599'
        WHEN min_score_2024 >= 500 THEN '500-549'
        ELSE '500以下'
    END as score_range,
    COUNT(*) as count
FROM admission_hubei_wide_2024 
GROUP BY score_range 
ORDER BY score_range;
```

### 3. 选科要求分析
```sql
-- 查看物理+化学要求的专业
SELECT school_name, major_name, min_score_2024
FROM admission_hubei_wide_2024 
WHERE require_physics = 1 AND require_chemistry = 1
ORDER BY min_score_2024 DESC 
LIMIT 10;

-- 统计各选科组合的专业数量
SELECT 
    require_physics,
    require_chemistry,
    require_biology,
    COUNT(*) as major_count
FROM admission_hubei_wide_2024 
GROUP BY require_physics, require_chemistry, require_biology
ORDER BY major_count DESC;
```

### 4. 专业分类分析
```sql
-- 工科类专业分析
SELECT 
    school_name, 
    major_name, 
    min_score_2024
FROM admission_hubei_wide_2024 
WHERE is_engineering = 1 
ORDER BY min_score_2024 DESC 
LIMIT 15;

-- 各专业类别统计
SELECT 
    '理科' as category, COUNT(*) as count FROM admission_hubei_wide_2024 WHERE is_science = 1
UNION ALL
SELECT 
    '工科' as category, COUNT(*) as count FROM admission_hubei_wide_2024 WHERE is_engineering = 1
UNION ALL
SELECT 
    '医科' as category, COUNT(*) as count FROM admission_hubei_wide_2024 WHERE is_medical = 1
UNION ALL
SELECT 
    '经管法' as category, COUNT(*) as count FROM admission_hubei_wide_2024 WHERE is_economics_mgmt_law = 1;
```

### 5. 院校分析
```sql
-- 985/211院校统计
SELECT 
    school_level,
    COUNT(*) as major_count,
    AVG(min_score_2024) as avg_score,
    MIN(min_score_2024) as min_score,
    MAX(min_score_2024) as max_score
FROM admission_hubei_wide_2024 
GROUP BY school_level 
ORDER BY avg_score DESC;

-- 特定院校的专业分布
SELECT 
    major_name,
    min_score_2024,
    min_rank_2024,
    plan_count_2024
FROM admission_hubei_wide_2024 
WHERE school_name = '华中科技大学'
ORDER BY min_score_2024 DESC;
```

### 6. 综合查询
```sql
-- 适合特定选科组合的高性价比专业
SELECT 
    school_name,
    major_name,
    min_score_2024,
    min_rank_2024,
    school_level
FROM admission_hubei_wide_2024 
WHERE require_physics = 1 
    AND require_chemistry = 1 
    AND require_biology = 0
    AND is_engineering = 1
    AND min_score_2024 BETWEEN 580 AND 620
ORDER BY min_score_2024 ASC;
```

## 📈 性能优化

### 索引策略
表已创建以下索引以优化查询性能：
- `min_score_2024`: MinMax索引，优化分数范围查询
- `min_rank_2024`: MinMax索引，优化位次范围查询

### 查询优化建议
1. **分数查询**: 使用范围查询而非精确匹配
2. **选科查询**: 利用布尔字段进行高效过滤
3. **排序查询**: 优先使用已索引字段排序
4. **聚合查询**: 利用ClickHouse的列式存储优势

## 🔧 维护操作

### 数据更新
```sql
-- 插入新数据（示例）
INSERT INTO admission_hubei_wide_2024 VALUES (...);

-- 更新现有数据（ClickHouse不支持UPDATE，需要重新插入）
-- 建议使用 INSERT INTO ... SELECT 进行数据迁移
```

### 备份与恢复
```bash
# 导出数据
clickhouse-client --port 19000 --query "SELECT * FROM admission_hubei_wide_2024 FORMAT CSV" > backup.csv

# 导入数据
clickhouse-client --port 19000 --query "INSERT INTO admission_hubei_wide_2024 FORMAT CSV" < backup.csv
```

## 📞 技术支持

### 服务器管理
- **启动服务器**: 使用配置文件启动ClickHouse
- **停止服务器**: `pkill -f clickhouse-server`
- **查看日志**: `/Users/jarviszuo/clickhouse_server/logs/`

### 连接工具推荐
- **DBeaver**: 免费的通用数据库工具
- **DataGrip**: JetBrains的专业数据库IDE
- **ClickHouse官方客户端**: 命令行工具

### 常见问题
1. **连接失败**: 检查端口配置，HTTP使用18123，TCP使用19000
2. **查询慢**: 检查是否使用了合适的WHERE条件和索引
3. **内存不足**: 调整查询的LIMIT或使用分页查询

---

*最后更新: 2024年6月14日*
*数据版本: 2024年湖北省本科专业录取数据* 