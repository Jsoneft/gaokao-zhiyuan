# 湖北省高考志愿数据分析系统

## 项目概述

本项目成功完成了两张Excel表的分析、合并和ClickHouse数据库建设，为湖北省高考志愿填报提供了强大的数据分析基础。

## 数据处理成果

### 原始数据
- **表1**: `21-24各省份录取数据(含专业组代码).xlsx` (112MB, 584,117行)
- **表2**: `最新最新21-24各省本科专业分.xlsx` (96MB, 584,117行)

### 湖北省数据筛选
- **筛选后数据**: 18,430行（湖北省专用数据）
- **有效录取数据**: 18,278条（含2024年录取分数和位次）
- **ID匹配率**: 100%（湖北省内无重复ID问题）

### 数据质量
- **科类分布**: 物理类 13,627条 (74.6%)，历史类 4,651条 (25.4%)
- **公私性质**: 公办 15,066条 (82.4%)，民办 3,212条 (17.6%)
- **教育层次**: 本科 18,278条 (100%)

## 数据库设计

### 表结构优化
- **表名**: `admission_hubei_wide_2024`
- **字段数**: 37个字段，全英文命名
- **记录数**: 18,278条
- **存储引擎**: ClickHouse MergeTree

### 核心字段设计

#### 基础信息
- `id`: 记录唯一标识
- `school_code/school_name`: 院校代码/名称
- `major_code/major_name`: 专业代码/名称
- `major_group_code`: 专业组代码

#### 选科限制优化
原始字段 `'化与生'` 等字符串拆分为6个布尔字段：
- `require_physics/chemistry/biology`: 理科选科要求
- `require_politics/history/geography`: 文科选科要求

#### 枚举类型优化
- `subject_category`: Enum8('物理'=1, '历史'=2)
- `school_ownership`: Enum8('公办'=1, '民办'=2)
- `education_level`: Enum8('本科'=1, '专科'=2)

#### 专业分类标签
7个布尔字段支持灵活查询：
- `is_science`: 理科专业 (1,215个, 6.6%)
- `is_engineering`: 工科专业 (8,557个, 46.8%)
- `is_medical`: 医科专业 (1,502个, 8.2%)
- `is_economics_mgmt_law`: 经管法专业 (4,900个, 26.8%)
- `is_liberal_arts`: 文科专业 (1,901个, 10.4%)
- `is_design_arts`: 设计与艺术类 (2,203个, 12.1%)
- `is_language`: 语言类专业 (1,521个, 8.3%)

#### 录取数据
- `min_score_2024`: 2024年最低录取分数 (372-692分)
- `min_rank_2024`: 2024年最低录取位次 (9-167,641位)
- `enrollment_plan_2024`: 2024年招生计划数

### 索引设计
- **主键排序**: (id, school_code, major_code)
- **MinMax索引**: min_score_2024, min_rank_2024
- **查询性能**: 平均0.2秒内完成复杂查询

## 技术实现

### 环境配置
- **ClickHouse版本**: 25.5.2.47 (官方构建)
- **运行模式**: ClickHouse Local
- **数据目录**: `~/clickhouse_local_data`
- **平台**: macOS (Apple Silicon)

### 文件结构
```
gaokao-zhiyuan/
├── hubei_data/
│   ├── table1_hubei.csv              # 湖北省表1数据
│   ├── table2_hubei.csv              # 湖北省表2数据
│   ├── create_hubei_optimized_en.sql # 建表SQL
│   └── insert_data_fixed.sql         # 插入数据SQL
├── generate_insert_sql.py            # 数据转换脚本
├── fix_insert_sql.py                 # 数据修复脚本
├── test_clickhouse_final.py          # 最终测试脚本
└── sample_queries.sql                # 示例查询文件
```

## 使用指南

### 启动ClickHouse
```bash
/opt/homebrew/Caskroom/clickhouse/25.5.2.47-stable/clickhouse-macos-aarch64 local --path ~/clickhouse_local_data
```

### 常用查询示例

#### 1. 基础统计
```sql
-- 查看表结构
DESCRIBE admission_hubei_wide_2024;

-- 统计总记录数
SELECT COUNT(*) FROM admission_hubei_wide_2024;

-- 科类分布
SELECT subject_category, COUNT(*) as count 
FROM admission_hubei_wide_2024 
GROUP BY subject_category;
```

#### 2. 高分专业查询
```sql
-- 高分专业TOP10
SELECT school_name, major_name, min_score_2024, min_rank_2024
FROM admission_hubei_wide_2024 
WHERE min_score_2024 IS NOT NULL
ORDER BY min_score_2024 DESC 
LIMIT 10;
```

#### 3. 专业分类查询
```sql
-- 工科专业分析
SELECT school_name, major_name, min_score_2024
FROM admission_hubei_wide_2024
WHERE is_engineering = true
ORDER BY min_score_2024 DESC LIMIT 10;

-- 医科专业查询
SELECT school_name, major_name, min_score_2024
FROM admission_hubei_wide_2024
WHERE is_medical = true
ORDER BY min_score_2024 DESC LIMIT 10;
```

#### 4. 选科要求查询
```sql
-- 要求化学+生物的专业
SELECT school_name, major_name, min_score_2024
FROM admission_hubei_wide_2024
WHERE require_chemistry = true AND require_biology = true
ORDER BY min_score_2024 DESC LIMIT 10;
```

#### 5. 分数段分析
```sql
-- 600-650分段专业数
SELECT COUNT(*) 
FROM admission_hubei_wide_2024 
WHERE min_score_2024 BETWEEN 600 AND 650;

-- 分数段分布
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
WHERE min_score_2024 IS NOT NULL
GROUP BY score_range
ORDER BY min(min_score_2024) DESC;
```

#### 6. 复合条件查询
```sql
-- 工科+选科要求+分数段
SELECT school_name, major_name, min_score_2024, min_rank_2024
FROM admission_hubei_wide_2024 
WHERE is_engineering = true 
  AND require_chemistry = true 
  AND min_score_2024 BETWEEN 550 AND 600
ORDER BY min_score_2024 DESC;
```

## 性能表现

### 查询性能测试结果
- **基础统计查询**: ~0.2秒
- **复合条件查询**: ~0.2秒
- **排序查询**: ~0.2秒
- **聚合查询**: ~0.2秒

### 数据完整性
- ✅ 18,278条有效录取数据
- ✅ 100% ID匹配率
- ✅ 完整的专业分类标签
- ✅ 优化的选科限制字段
- ✅ 标准化的枚举类型

## 项目亮点

### 1. 数据质量优化
- 解决了原始数据中的重复记录问题
- 实现了湖北省数据的精准筛选
- 完成了字段类型的标准化

### 2. 数据库设计优化
- 英文字段命名，便于程序化查询
- 布尔字段拆分，支持灵活的选科查询
- 枚举类型优化，节省存储空间
- 合理的索引设计，提升查询性能

### 3. 查询功能丰富
- 支持多维度专业分类查询
- 支持复杂的选科要求组合
- 支持分数段和位次范围查询
- 支持院校和专业的多重筛选

### 4. 技术架构先进
- 使用ClickHouse列式存储，查询性能优异
- Local模式部署，无需复杂的服务器配置
- 完整的数据处理和验证流程

## 应用场景

### 1. 高考志愿填报
- 根据分数查询可报考专业
- 根据选科组合筛选专业
- 分析专业录取难度和趋势

### 2. 教育数据分析
- 专业热度和分布分析
- 院校竞争力评估
- 选科政策影响分析

### 3. 决策支持系统
- 为教育部门提供数据支撑
- 为学校提供专业设置参考
- 为学生提供科学的志愿填报建议

## 总结

本项目成功实现了：
1. ✅ 两张大型Excel表的深度分析和合并
2. ✅ 湖北省专用数据的精准提取和优化
3. ✅ 高性能ClickHouse数据库的建设
4. ✅ 丰富的查询功能和优异的性能表现
5. ✅ 完整的数据处理和验证流程

数据库现已就绪，可以为湖北省高考志愿填报提供强大的数据分析支持！

---

**技术支持**: ClickHouse Local 25.5.2.47  
**数据更新**: 2024年录取数据  
**覆盖范围**: 湖北省全部本科专业  
**查询性能**: 平均响应时间 < 0.3秒 