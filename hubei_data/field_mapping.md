# 字段映射表 - 中文到英文

## 基础标识字段
| 中文字段名 | 英文字段名 | 类型 | 说明 |
|-----------|-----------|------|------|
| id | id | UInt32 | 记录唯一标识 |
| 院校代码 | school_code | String | 院校代码 |
| 院校名称 | school_name | String | 院校名称 |
| 专业代码 | major_code | String | 专业代码 |
| 专业名称 | major_name | String | 专业名称 |
| 专业组代码 | major_group_code | String | 专业组代码 |

## 地域和批次信息
| 中文字段名 | 英文字段名 | 类型 | 说明 |
|-----------|-----------|------|------|
| 生源地 | source_province | LowCardinality(String) | 生源地，固定为湖北 |
| 所在省 | school_province | LowCardinality(String) | 院校所在省份 |
| 城市 | school_city | String | 院校所在城市 |
| 批次 | admission_batch | LowCardinality(String) | 录取批次 |

## 科类和选科限制
| 中文字段名 | 英文字段名 | 类型 | 说明 |
|-----------|-----------|------|------|
| 科类 | subject_category | Enum8('物理'=1, '历史'=2) | 科类：物理类或历史类 |
| - | require_physics | Bool | 是否要求选择物理 |
| - | require_chemistry | Bool | 是否要求选择化学 |
| - | require_biology | Bool | 是否要求选择生物 |
| - | require_politics | Bool | 是否要求选择政治 |
| - | require_history | Bool | 是否要求选择历史 |
| - | require_geography | Bool | 是否要求选择地理 |
| 选科限制 | subject_requirement_raw | LowCardinality(String) | 原始选科限制描述，用于显示 |

## 院校基本信息
| 中文字段名 | 英文字段名 | 类型 | 说明 |
|-----------|-----------|------|------|
| 类型 | school_type | LowCardinality(String) | 院校类型 |
| 公私性质 | school_ownership | Enum8('公办'=1, '民办'=2) | 公办或民办 |
| 隶属单位 | school_authority | LowCardinality(String) | 院校隶属单位 |
| 院校水平 | school_level | LowCardinality(String) | 院校水平层次 |
| 院校标签 | school_tags | String | 院校标签，如985、211等 |
| 本科/专科 | education_level | Enum8('本科'=1, '专科'=2) | 本科或专科 |

## 专业信息
| 中文字段名 | 英文字段名 | 类型 | 说明 |
|-----------|-----------|------|------|
| 专业备注 | major_description | String | 专业备注信息 |
| 学制 | study_years | UInt8 | 学制年数 |
| 学费 | tuition_fee | UInt32 | 学费（元/年） |
| 新增专业 | is_new_major | Bool | 是否为新增专业 |

## 2024年录取数据
| 中文字段名 | 英文字段名 | 类型 | 说明 |
|-----------|-----------|------|------|
| 专业组最低分_2024 | min_score_2024 | UInt16 | 2024年专业组最低录取分数 |
| 专业组最低位次_2024 | min_rank_2024 | UInt32 | 2024年专业组最低录取位次 |
| 计划数_2024 | enrollment_plan_2024 | UInt16 | 2024年招生计划数 |

## 专业分类标签
| 中文字段名 | 英文字段名 | 类型 | 说明 |
|-----------|-----------|------|------|
| 理科 | is_science | Bool | 是否为理科专业 |
| 工科 | is_engineering | Bool | 是否为工科专业 |
| 医科 | is_medical | Bool | 是否为医科专业 |
| 经管法 | is_economics_mgmt_law | Bool | 是否为经管法专业 |
| 文科（非经管法） | is_liberal_arts | Bool | 是否为文科专业（非经管法） |
| 设计与艺术类 | is_design_arts | Bool | 是否为设计与艺术类专业 |
| 语言类 | is_language | Bool | 是否为语言类专业 |

## 索引说明

### 主键索引
```sql
ORDER BY (id, school_code, major_code)
```

### 排序优化索引
```sql
-- 分数排序索引（用于按分数排序查询）
ALTER TABLE gaokao_hubei_optimized ADD INDEX idx_min_score_2024 min_score_2024 TYPE minmax GRANULARITY 1;

-- 位次排序索引（用于按位次排序查询）
ALTER TABLE gaokao_hubei_optimized ADD INDEX idx_min_rank_2024 min_rank_2024 TYPE minmax GRANULARITY 1;
```

### 索引类型说明
- **minmax索引**: 适用于数值范围查询和排序，存储每个granule的最小值和最大值
- **GRANULARITY 1**: 每个index granule对应1个data granule，提供最精确的过滤

## 常用查询模式

### 1. 按分数排序查询
```sql
SELECT school_name, major_name, min_score_2024, min_rank_2024
FROM gaokao_hubei_optimized
WHERE min_score_2024 IS NOT NULL
ORDER BY min_score_2024 DESC;
```

### 2. 按位次排序查询
```sql
SELECT school_name, major_name, min_score_2024, min_rank_2024
FROM gaokao_hubei_optimized
WHERE min_rank_2024 IS NOT NULL
ORDER BY min_rank_2024 ASC;
```

### 3. 分数段查询
```sql
SELECT school_name, major_name, min_score_2024
FROM gaokao_hubei_optimized
WHERE min_score_2024 BETWEEN 600 AND 650
ORDER BY min_score_2024 DESC;
```

### 4. 位次段查询
```sql
SELECT school_name, major_name, min_rank_2024
FROM gaokao_hubei_optimized
WHERE min_rank_2024 BETWEEN 1000 AND 5000
ORDER BY min_rank_2024 ASC;
``` 