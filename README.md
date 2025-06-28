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


└── hubei_data/                 # 湖北省专用数据
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

## ClickHouse 数据库表结构

### 主要数据表

#### 1. admission_hubei_wide_2024 (湖北省录取数据表 - 主表)

这是系统的核心数据表，包含了湖北省2024年的高考录取数据：

```sql
CREATE TABLE IF NOT EXISTS admission_hubei_wide_2024 (
    id                      UInt32,                    -- 记录唯一ID
    school_code             String,                    -- 学校代码
    school_name             String,                    -- 学校名称
    major_code              String,                    -- 专业代码
    major_name              String,                    -- 专业名称
    major_group_code        String,                    -- 专业组代码
    source_province         LowCardinality(String),    -- 生源省份
    school_province         LowCardinality(String),    -- 学校所在省份
    school_city             String,                    -- 学校所在城市
    admission_batch         LowCardinality(String),    -- 录取批次
    subject_category        Enum8('物理'=1, '历史'=2), -- 科目类别
    require_physics         Bool,                      -- 是否要求物理
    require_chemistry       Bool,                      -- 是否要求化学
    require_biology         Bool,                      -- 是否要求生物
    require_politics        Bool,                      -- 是否要求政治
    require_history         Bool,                      -- 是否要求历史
    require_geography       Bool,                      -- 是否要求地理
    subject_requirement_raw LowCardinality(String),    -- 原始选科要求
    school_type             LowCardinality(String),    -- 学校类型
    school_ownership        Enum8('公办'=1, '民办'=2), -- 学校性质
    school_authority        LowCardinality(String),    -- 学校主管部门
    school_level            LowCardinality(String),    -- 学校层次(985/211等)
    school_tags             String,                    -- 学校标签
    education_level         Enum8('本科'=1, '专科'=2), -- 教育层次
    major_description       String,                    -- 专业描述
    study_years             UInt8,                     -- 学制年限
    tuition_fee             UInt32,                    -- 学费
    is_new_major            Bool,                      -- 是否新专业
    min_score_2024          UInt16,                    -- 2024年最低分
    min_rank_2024           UInt32,                    -- 2024年最低位次
    enrollment_plan_2024    UInt16,                    -- 2024年招生计划
    is_science              Bool,                      -- 是否理科
    is_engineering          Bool,                      -- 是否工科
    is_medical              Bool,                      -- 是否医科
    is_economics_mgmt_law   Bool,                      -- 是否经管法
    is_liberal_arts         Bool,                      -- 是否文科
    is_design_arts          Bool,                      -- 是否设计艺术
    is_language             Bool                       -- 是否语言类
) ENGINE = MergeTree()
ORDER BY (id, school_code, major_code)
SETTINGS index_granularity = 8192
```

**索引说明**:
- 主键：`(id, school_code, major_code)`
- 优化查询：ID查询、学校查询、专业查询

#### 2. admission_data (兼容性数据表)

为了保持向后兼容，系统还支持旧的数据表结构：

```sql
CREATE TABLE IF NOT EXISTS admission_data (
    id                       UInt64,     -- 自增ID
    year                     UInt32,     -- 年份
    province                 String,     -- 省份
    batch                    String,     -- 批次
    subject_type             String,     -- 科类
    class_demand             String,     -- 选科要求
    college_code             String,     -- 院校代码
    special_interest_group_code String,  -- 专业组代码
    college_name             String,     -- 院校名称
    professional_code        String,     -- 专业代码
    professional_name        String,     -- 专业名称
    lowest_points            Int64,      -- 录取最低分
    lowest_rank              Int64,      -- 录取最低位次
    description              String      -- 备注
) ENGINE = MergeTree()
ORDER BY (lowest_rank, lowest_points, year, province)
```

**索引说明**:
- 主键：`(lowest_rank, lowest_points, year, province)`
- 优化查询：位次查询、分数查询、年份筛选、省份筛选

### 表结构设计特点

1. **性能优化**:
   - 使用 `LowCardinality` 类型优化重复值存储
   - 使用 `Enum8` 类型节省存储空间
   - 合理设计排序键提升查询性能

2. **数据类型选择**:
   - `UInt32/UInt16` 用于ID和分数，节省空间
   - `Bool` 类型用于标志位，清晰明确
   - `String` 类型用于文本数据

3. **业务逻辑支持**:
   - 选科要求拆分为独立布尔字段，便于查询
   - 学科分类标志位支持兴趣推荐
   - 分数和位次字段支持核心查询功能

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

## 更新日志

### v2.0.0 (2024-01-XX)
**🔄 重大更新 - 项目清理和文档完善**

#### 🧹 项目清理
- **删除Windows相关文件**: 移除了run.bat、run.ps1、build.ps1、deploy_manual.ps1等Windows脚本
- **删除编译文件**: 清理了所有编译后的二进制文件(gaokao-zhiyuan、main、test_ch_connection等)
- **删除Python脚本**: 移除了analyze_*.py、verify_*.py等数据分析脚本
- **删除大型数据文件**: 清理了Excel数据文件，减小项目体积
- **删除冗余文档**: 移除了重复的MD文档文件

#### 🔒 安全改进
- **删除所有shell脚本**: 移除了包含敏感信息的.sh文件
- **密码安全**: 将所有硬编码密码改为环境变量读取
- **敏感信息清理**: 清理了代码中的服务器密码和SSH凭据

#### 📚 文档完善
- **完整API文档**: 新增详细的接口参数和响应示例
- **数据库表结构**: 添加完整的ClickHouse表结构说明，包含37个字段的详细说明
- **工具程序分析**: 详细分析tools目录下各工具的功能和保留建议
- **配置文档**: 完善环境变量配置说明

#### 🛠️ 工具优化
**完全删除tools目录**:
- ❌ 删除整个 `tools/` 目录及其所有文件
- ❌ 移除了7个工具程序：数据备份、导入、统计等工具
- ❌ 修复了工具文件中的编译错误问题
- ✅ 简化项目结构，专注核心API功能

#### 🏗️ 架构改进
- **表结构优化**: 详细说明admission_hubei_wide_2024表的37个字段
- **Makefile简化**: 移除了对已删除文件的引用，保留核心功能
- **目录结构清理**: 删除了空目录和临时文件，项目更加简洁

#### 📊 数据库详细说明
- **主表**: admission_hubei_wide_2024 (18,278条记录)
- **字段分类**: 基础信息(5个)、录取数据(12个)、选科要求(6个)、专业分类(7个)、院校信息(5个)、地理信息(2个)
- **索引设计**: MinMax索引优化分数和位次查询
- **数据统计**: 约1,200所院校，8,000个专业，覆盖2021-2024年数据
- **兼容性保持**: 保留旧表结构确保向后兼容
- **索引优化**: 优化数据库索引提升查询效率

### v1.x.x (历史版本)
- 基础API功能实现
- ClickHouse数据库集成
- 湖北省数据支持

## 注意事项

1. **环境变量**: 确保所有必要的环境变量都已正确设置
2. **数据库连接**: 确保 ClickHouse 服务正常运行且可访问
3. **端口配置**: 确保配置的端口未被占用
4. **数据安全**: 不要在代码中硬编码密码等敏感信息
5. **性能优化**: 大数据量查询时注意分页和索引优化
6. **工具清理**: 建议定期清理不必要的工具程序，保持代码库整洁

## 许可证

本项目采用 MIT 许可证，详见 LICENSE 文件。
