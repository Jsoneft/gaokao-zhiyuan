ATTACH TABLE _ UUID '64fedb27-d8fc-4275-a977-4849f2f5b742'
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
    `year` UInt16,
    `lowest_points` Int64,
    `lowest_rank` Int64
)
ENGINE = MergeTree
ORDER BY (lowest_rank, lowest_points, year, province)
SETTINGS index_granularity = 8192
