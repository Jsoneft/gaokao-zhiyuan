ATTACH TABLE _ UUID 'ed6cebf8-d91c-4501-b82f-4f8a515a3dbf'
(
    `id` UInt64,
    `province` String,
    `batch` String,
    `subject_type` String,
    `class_demand` String,
    `college_code` String,
    `special_interest_group_code` String,
    `college_name` String,
    `professional_code` String,
    `professional_name` String,
    `description` String,
    `year` UInt32,
    `lowest_points` Int64,
    `lowest_rank` Int64
)
ENGINE = MergeTree
ORDER BY (lowest_rank, lowest_points, year, province)
SETTINGS index_granularity = 8192
