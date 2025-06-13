-- 湖北省高考志愿数据大宽表 (英文字段名优化版)
-- 优化内容:
-- 1. 所有字段使用英文命名，便于开发和维护
-- 2. 选科限制拆分为独立的布尔字段，支持灵活查询
-- 3. 科类标准化为枚举类型
-- 4. 移除23年及之前的录取数据字段
-- 5. 专业分类标签改为布尔类型
-- 6. 录取分数和位次改为整数类型
-- 7. 为排序字段添加专门索引
-- 生成时间: 2025-06-14

CREATE TABLE IF NOT EXISTS admission_hubei_wide_2024
(
    id                      UInt32 COMMENT '记录唯一标识',
    school_code             String COMMENT '院校代码',
    school_name             String COMMENT '院校名称',
    major_code              String COMMENT '专业代码',
    major_name              String COMMENT '专业名称',
    major_group_code        String COMMENT '专业组代码',
    source_province         LowCardinality(String) COMMENT '生源地，固定为湖北',
    school_province         LowCardinality(String) COMMENT '院校所在省份',
    school_city             String COMMENT '院校所在城市',
    admission_batch         LowCardinality(String) COMMENT '录取批次',
    subject_category        Enum8('物理'=1, '历史'=2) COMMENT '科类：物理类或历史类',
    require_physics         Bool COMMENT '是否要求选择物理',
    require_chemistry       Bool COMMENT '是否要求选择化学',
    require_biology         Bool COMMENT '是否要求选择生物',
    require_politics        Bool COMMENT '是否要求选择政治',
    require_history         Bool COMMENT '是否要求选择历史',
    require_geography       Bool COMMENT '是否要求选择地理',
    subject_requirement_raw LowCardinality(String) COMMENT '原始选科限制描述，用于显示',
    school_type             LowCardinality(String) COMMENT '院校类型',
    school_ownership        Enum8('公办'=1, '民办'=2) COMMENT '公办或民办',
    school_authority        LowCardinality(String) COMMENT '院校隶属单位',
    school_level            LowCardinality(String) COMMENT '院校水平层次',
    school_tags             String COMMENT '院校标签，如985、211等',
    education_level         Enum8('本科'=1, '专科'=2) COMMENT '本科或专科',
    major_description       String COMMENT '专业备注信息',
    study_years             UInt8 COMMENT '学制年数',
    tuition_fee             UInt32 COMMENT '学费（元/年）',
    is_new_major            Bool COMMENT '是否为新增专业',
    min_score_2024          UInt16 COMMENT '2024年专业组最低录取分数',
    min_rank_2024           UInt32 COMMENT '2024年专业组最低录取位次',
    enrollment_plan_2024    UInt16 COMMENT '2024年招生计划数',
    is_science              Bool COMMENT '是否为理科专业',
    is_engineering          Bool COMMENT '是否为工科专业',
    is_medical              Bool COMMENT '是否为医科专业',
    is_economics_mgmt_law   Bool COMMENT '是否为经管法专业',
    is_liberal_arts         Bool COMMENT '是否为文科专业（非经管法）',
    is_design_arts          Bool COMMENT '是否为设计与艺术类专业',
    is_language             Bool COMMENT '是否为语言类专业'
)
ENGINE = MergeTree()
ORDER BY (id, school_code, major_code)
SETTINGS index_granularity = 8192;

-- 核心排序字段索引（用于ORDER BY查询优化）
ALTER TABLE admission_hubei_wide_2024 ADD INDEX idx_min_score_2024 min_score_2024 TYPE minmax GRANULARITY 1;
ALTER TABLE admission_hubei_wide_2024 ADD INDEX idx_min_rank_2024 min_rank_2024 TYPE minmax GRANULARITY 1;

-- 查询示例

-- 1. 查询物理类且要求化学+生物的工科专业，按分数排序
-- SELECT school_name, major_name, min_score_2024, min_rank_2024
-- FROM admission_hubei_wide_2024
-- WHERE subject_category = '物理'
--   AND require_chemistry = true
--   AND require_biology = true
--   AND is_engineering = true
--   AND min_score_2024 IS NOT NULL
-- ORDER BY min_score_2024 DESC;

-- 2. 查询不限选科的985院校，按位次排序
-- SELECT school_name, major_name, min_score_2024, min_rank_2024
-- FROM admission_hubei_wide_2024
-- WHERE require_physics = false
--   AND require_chemistry = false
--   AND require_biology = false
--   AND require_politics = false
--   AND require_history = false
--   AND require_geography = false
--   AND school_tags LIKE '%985%'
--   AND min_rank_2024 IS NOT NULL
-- ORDER BY min_rank_2024 ASC;

-- 3. 查询医科专业，按录取分数排序
-- SELECT school_name, major_name, min_score_2024, min_rank_2024, school_level
-- FROM admission_hubei_wide_2024
-- WHERE is_medical = true
--   AND min_score_2024 IS NOT NULL
-- ORDER BY min_score_2024 DESC
-- LIMIT 50;

-- 4. 查询特定分数段的专业
-- SELECT school_name, major_name, min_score_2024, min_rank_2024
-- FROM admission_hubei_wide_2024
-- WHERE min_score_2024 BETWEEN 600 AND 650
--   AND subject_category = '物理'
-- ORDER BY min_score_2024 DESC;

-- 5. 查询特定位次段的专业
-- SELECT school_name, major_name, min_score_2024, min_rank_2024
-- FROM admission_hubei_wide_2024
-- WHERE min_rank_2024 BETWEEN 1000 AND 5000
--   AND is_engineering = true
-- ORDER BY min_rank_2024 ASC; 