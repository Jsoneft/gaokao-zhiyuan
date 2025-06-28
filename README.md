# 高考志愿填报系统 API

## 项目简介

高考志愿填报系统是一个基于 Go 语言开发的 Web API 服务，主要提供高考分数位次查询和志愿填报建议功能。系统使用 ClickHouse 作为数据库，提供高性能的数据查询服务。

## 技术栈

- **后端语言**: Go 1.21+
- **Web框架**: Gin
- **数据库**: ClickHouse
- **配置管理**: 环境变量 + .env 文件

## 目录结构

```
gaokao-zhiyuan/
├── main.go                     # 主程序入口
├── go.mod                      # Go 模块依赖
├── go.sum                      # 依赖版本锁定
├── config/
│   └── config.go              # 配置管理
├── database/
│   └── clickhouse.go          # ClickHouse 数据库连接和操作
├── handlers/
│   └── handlers.go            # HTTP 请求处理器
├── models/
│   └── models.go              # 数据模型定义
├── tools/                      # 工具程序
│   ├── ch_inspect.go          # ClickHouse 数据检查工具
│   ├── clickhouse_stats.go    # 数据统计工具
│   ├── hubei_stats.go         # 湖北数据统计
│   ├── stat_query.go          # 查询统计工具
│   ├── hubei_fixed.go         # 湖北数据修复工具
│   ├── province_source_stats.go # 省份数据统计
│   ├── export_clickhouse.go   # 数据导出工具
│   ├── verify_hubei_ids.go    # 湖北ID验证工具
│   ├── simple_verify.go       # 简单验证工具
│   ├── update_major_scores.go # 专业分数更新工具
│   └── hubei_import/          # 湖北数据导入工具
├── data/                       # 数据文件目录
├── hubei_data/                 # 湖北省专用数据
├── scripts/                    # 脚本文件
├── user_files/                 # 用户上传文件
├── flags/                      # 功能标志文件
├── bin/                        # 编译后的二进制文件
└── preprocessed_configs/       # 预处理配置文件
```

## API 接口文档

### 1. 健康检查

**接口地址**: `GET /api/health`

**功能**: 检查服务健康状态

**响应示例**:
```json
{
  "code": 0,
  "msg": "服务正常运行"
}
```

### 2. 分数位次查询

**接口地址**: `GET /api/rank/get`

**功能**: 根据分数查询对应的位次

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| score | float | 是 | 高考分数 |

**请求示例**:
```
GET /api/rank/get?score=555
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "success",
  "rank": 45678,
  "year": 2024,
  "score": 555
}
```

### 3. 高级位次查询

**接口地址**: `POST /api/v1/query_rank`

**功能**: 根据多个条件查询位次

**请求参数**:
```json
{
  "province": "湖北",
  "year": 2024,
  "score": 555,
  "subject_type": "物理",
  "class_demand": ["物", "化", "生"]
}
```

**参数说明**:
| 参数名 | 类型 | 必填 | 默认值 | 说明 |
|--------|------|------|--------|------|
| province | string | 否 | "湖北" | 省份 |
| year | int | 否 | 2024 | 年份 |
| score | int64 | 是 | - | 高考分数 |
| subject_type | string | 否 | "物理" | 科目类型 |
| class_demand | []string | 否 | ["物","化","生"] | 选科要求 |

**响应示例**:
```json
{
  "code": 0,
  "msg": "success",
  "rank": 45678,
  "year": 2024,
  "province": "湖北",
  "subject_type": "物理",
  "score": 555
}
```

### 4. 志愿填报报表查询

**接口地址**: `GET /api/report/get`

**功能**: 根据位次和条件查询推荐的院校专业

**请求参数**:
| 参数名 | 类型 | 必填 | 默认值 | 说明 |
|--------|------|------|--------|------|
| rank | int64 | 是 | - | 位次 |
| class_first_choise | string | 否 | - | 首选科目 |
| class_optional_choise | string | 否 | - | 可选科目(JSON数组字符串) |
| province | string | 否 | - | 省份 |
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 10 | 每页数量(最大100) |
| college_location | string | 否 | - | 院校地区(JSON数组字符串) |
| interest | string | 否 | - | 兴趣方向(JSON数组字符串) |
| strategy | int | 否 | 0 | 填报策略 |

**请求示例**:
```
GET /api/report/get?rank=50000&class_first_choise=物理&class_optional_choise=["化学","生物"]&province=湖北&page=1&page_size=10&college_location=["湖北"]&interest=["理科","工科"]&strategy=0
```

**响应示例**:
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "conf": {
      "page": 1,
      "page_size": 10,
      "total_number": 1500,
      "total_page": 150
    },
    "list": [
      {
        "id": 12345,
        "college_name": "华中科技大学",
        "college_code": "10487",
        "professional_name": "计算机科学与技术",
        "class_demand": "物理+化学",
        "college_province": "湖北",
        "college_city": "武汉",
        "college_ownership": "公办",
        "college_type": "综合",
        "college_authority": "教育部",
        "college_level": "985",
        "education_level": "本科",
        "tuition_fee": 5850,
        "study_years": "4",
        "lowest_points": 580,
        "lowest_rank": 12000,
        "major_min_score_2024": 585,
        "is_new_major": false
      }
    ]
  }
}
```

## 配置文件结构

### 环境变量配置

系统通过环境变量进行配置，支持 `.env` 文件：

```bash
# 服务配置
PORT=8031                           # 服务端口
GIN_MODE=release                    # Gin运行模式 (debug/release)

# ClickHouse 数据库配置
CLICKHOUSE_HOST=localhost           # ClickHouse 主机地址
CLICKHOUSE_PORT=19000              # ClickHouse 端口
CLICKHOUSE_USERNAME=default         # ClickHouse 用户名
CLICKHOUSE_PASSWORD=               # ClickHouse 密码
CLICKHOUSE_DATABASE=gaokao         # ClickHouse 数据库名
```

### 配置加载逻辑

配置通过 `config/config.go` 加载：

```go
type Config struct {
    Port               string  // 服务端口
    GinMode            string  // Gin运行模式
    ClickHouseHost     string  // ClickHouse主机
    ClickHousePort     int     // ClickHouse端口
    ClickHouseUser     string  // ClickHouse用户名
    ClickHousePassword string  // ClickHouse密码
    ClickHouseDatabase string  // ClickHouse数据库名
}
```

## 数据库表结构

### 主要数据表

#### admission_hubei_wide_2024 (湖北省录取数据表)

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | UInt32 | 记录ID |
| school_code | String | 学校代码 |
| school_name | String | 学校名称 |
| major_code | String | 专业代码 |
| major_name | String | 专业名称 |
| major_group_code | String | 专业组代码 |
| source_province | String | 生源省份 |
| school_province | String | 学校省份 |
| school_city | String | 学校城市 |
| admission_batch | String | 录取批次 |
| subject_category | Enum8 | 科目类别('物理'=1, '历史'=2) |
| require_physics | Bool | 是否要求物理 |
| require_chemistry | Bool | 是否要求化学 |
| require_biology | Bool | 是否要求生物 |
| require_politics | Bool | 是否要求政治 |
| require_history | Bool | 是否要求历史 |
| require_geography | Bool | 是否要求地理 |
| subject_requirement_raw | String | 原始选科要求 |
| school_type | String | 学校类型 |
| school_ownership | Enum8 | 学校性质('公办'=1, '民办'=2) |
| school_authority | String | 学校主管部门 |
| school_level | String | 学校层次 |
| school_tags | String | 学校标签 |
| education_level | Enum8 | 教育层次('本科'=1, '专科'=2) |
| major_description | String | 专业描述 |
| study_years | UInt8 | 学制年限 |
| tuition_fee | UInt32 | 学费 |
| is_new_major | Bool | 是否新专业 |
| min_score_2024 | UInt16 | 2024年最低分 |
| min_rank_2024 | UInt32 | 2024年最低位次 |
| enrollment_plan_2024 | UInt16 | 2024年招生计划 |
| is_science | Bool | 是否理科 |
| is_engineering | Bool | 是否工科 |
| is_medical | Bool | 是否医科 |
| is_economics_mgmt_law | Bool | 是否经管法 |
| is_liberal_arts | Bool | 是否文科 |
| is_design_arts | Bool | 是否设计艺术 |
| is_language | Bool | 是否语言类 |

## 部署说明

### 1. 环境准备

- Go 1.21+ 
- ClickHouse 数据库
- Linux/Windows/macOS 系统

### 2. 编译运行

```bash
# 安装依赖
go mod download

# 编译
go build -o gaokao-zhiyuan main.go

# 运行
./gaokao-zhiyuan
```

### 3. Docker 部署

```bash
# 构建镜像
docker build -t gaokao-zhiyuan .

# 运行容器
docker run -d \
  -p 8031:8031 \
  -e CLICKHOUSE_HOST=your_clickhouse_host \
  -e CLICKHOUSE_PORT=19000 \
  -e CLICKHOUSE_USERNAME=default \
  -e CLICKHOUSE_PASSWORD=your_password \
  -e CLICKHOUSE_DATABASE=gaokao \
  gaokao-zhiyuan
```

## 开发说明

### 项目结构说明

- `main.go`: 程序入口，设置路由和中间件
- `config/`: 配置管理模块
- `database/`: 数据库连接和操作
- `handlers/`: HTTP请求处理
- `models/`: 数据模型定义
- `tools/`: 各种工具程序

### 添加新接口

1. 在 `models/models.go` 中定义数据结构
2. 在 `database/clickhouse.go` 中添加数据库操作方法
3. 在 `handlers/handlers.go` 中添加HTTP处理方法
4. 在 `main.go` 中添加路由

### 数据库操作

系统使用 ClickHouse 作为主数据库，主要操作包括：
- 位次查询
- 院校专业查询
- 数据统计分析

## 注意事项

1. **环境变量**: 确保所有必要的环境变量都已正确设置
2. **数据库连接**: 确保 ClickHouse 服务正常运行且可访问
3. **端口配置**: 确保配置的端口未被占用
4. **数据安全**: 不要在代码中硬编码密码等敏感信息
5. **性能优化**: 大数据量查询时注意分页和索引优化

## 许可证

本项目采用 MIT 许可证，详见 LICENSE 文件。
