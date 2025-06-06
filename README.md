# gaokao-zhiyuan
高考志愿填报辅助系统

高考志愿填报辅助系统 - 这个系统是一个辅助中国高考考生填写高考志愿的后端API服务。

## 🚨 重要提示：数据文件说明

**Excel数据文件未包含在仓库中**：由于 `21-24各省份录取数据(含专业组代码).xlsx` 文件大小为112MB，超过了GitHub的100MB文件大小限制，该文件已从仓库中移除。

**部署前请先准备数据文件**：
- 请确保在项目根目录有 `21-24各省份录取数据(含专业组代码).xlsx` 文件
- 详细的数据文件获取和设置说明请参考：[DATA_SETUP.md](DATA_SETUP.md)

## 系统功能

用户输入自己的分数、位次、以及选课组合，系统返回个性化的院校-专业组推荐清单。

### 交互流程
```
输入: 
- 高考分数：______分
- 全省位次：______名  
- 选科组合：________

输出:
- 院校-专业组推荐清单（45个）
```

## 技术架构

- **后端框架**: Go + Gin
- **数据库**: ClickHouse
- **数据源**: Excel文件 (`21-24各省份录取数据(含专业组代码).xlsx`)

## 项目结构

```
gaokao-zhiyuan/
├── config/          # 配置管理
├── database/        # 数据库操作
├── handlers/        # API处理器
├── models/          # 数据模型
├── scripts/         # 部署脚本
├── tools/           # 工具程序
├── main.go          # 主程序入口
├── Makefile         # 构建脚本
├── README.md        # 项目说明
├── DATA_SETUP.md    # 数据文件设置说明
└── DEPLOYMENT.md    # 部署指南
```

## API接口

### 1. 位次查询接口

根据高考分数查询对应位次。

**请求**:
```bash
GET /api/rank/get?score=555
```

**响应**:
```json
{
    "code": 0,
    "msg": "success", 
    "rank": 72387,
    "year": 2024
}
```

### 2. 报表查询接口

根据位次和选科组合查询推荐院校专业。

**请求**:
```bash
GET /api/report/get?rank=12000&class_comb="123"&page=1&page_size=20
```

**参数说明**:
- `rank`: 用户位次
- `class_comb`: 选科组合字符串
  - 物理=1, 化学=2, 生物=3, 政治=4, 历史=5, 地理=6
  - 例如: 物理+化学+生物 = "123"
- `page`: 页码 (默认1)
- `page_size`: 每页大小 (默认20)

**响应**:
```json
{
    "code": 0,
    "msg": "success",
    "data": {
        "conf": {
            "page": 1,
            "page_size": 20,
            "total_number": 100,
            "total_page": 5
        },
        "list": [
            {
                "id": 1,
                "colledge_code": "A01",
                "colledge_name": "武汉大学", 
                "special_interest_group_code": "01",
                "professional_name": "计算机类",
                "class_demand": "物理+化学",
                "lowest_points": 625,
                "lowest_rank": 3500,
                "description": "国家特色专业，就业率98%"
            }
        ]
    }
}
```

### 3. 健康检查

**请求**:
```bash
GET /api/health
```

## 快速开始

### 1. 环境要求

- Go 1.21+
- ClickHouse 22.0+
- Linux/macOS/Windows

### 2. 本地开发

```bash
# 克隆项目
git clone <repository-url>
cd gaokao-zhiyuan

# ⚠️ 重要：准备数据文件
# 请确保 21-24各省份录取数据(含专业组代码).xlsx 文件存在于项目根目录
# 如果没有此文件，请参考 DATA_SETUP.md 获取

# 下载依赖
make deps

# 安装ClickHouse (Ubuntu/Debian)
sudo scripts/install_clickhouse.sh

# 编译项目
make build

# 导入Excel数据
make import

# 启动服务
make run
```

### 3. 一键部署到远程服务器

```bash
# ⚠️ 部署前确保数据文件已准备好
# 可以先手动上传数据文件到服务器，或修改部署脚本从远程下载

# 部署到远程服务器 (需要SSH密钥认证)
make deploy SERVER=192.168.1.100 USERNAME=root

# 或指定端口
make deploy SERVER=192.168.1.100 USERNAME=ubuntu PORT=2222
```

**部署注意事项**：
- 部署脚本会尝试在服务器上查找Excel数据文件
- 如果文件不存在，请先手动上传或修改部署脚本
- 详细说明请参考 [DEPLOYMENT.md](DEPLOYMENT.md)

部署脚本会自动完成：
- 安装Go环境
- 安装ClickHouse
- 编译项目
- 导入数据
- 配置系统服务
- 配置防火墙

### 4. 环境变量

可以通过环境变量配置服务：

```bash
export PORT=8031                    # 服务端口
export GIN_MODE=release             # Gin模式
export CLICKHOUSE_HOST=localhost    # ClickHouse主机
export CLICKHOUSE_PORT=9000         # ClickHouse端口  
export CLICKHOUSE_USERNAME=default  # ClickHouse用户名
export CLICKHOUSE_PASSWORD=         # ClickHouse密码
export CLICKHOUSE_DATABASE=gaokao   # 数据库名称
```

## 业务逻辑

### 位次查询逻辑
1. 在2024年数据中按分数排序
2. 找到用户分数对应的位置
3. 返回对应专业的最低位次

### 报表查询逻辑  
1. 根据用户位次计算去年等位分
2. 计算分数范围：等位分+20分 到 等位分-30分
3. 根据选科组合过滤专业
4. 在分数范围内查询2024年数据
5. 按专业最低分排序返回结果

## 管理命令

查看所有可用命令：
```bash
make help
```

常用命令：
```bash
make build     # 编译项目
make run       # 运行服务
make import    # 导入数据
make clean     # 清理编译文件
make test      # 运行测试
make fmt       # 格式化代码
```

## 服务管理

部署后可使用systemd管理服务：

```bash
# 查看服务状态
sudo systemctl status gaokao-server

# 查看服务日志
sudo journalctl -u gaokao-server -f

# 重启服务
sudo systemctl restart gaokao-server

# 停止服务  
sudo systemctl stop gaokao-server
```

## 数据表结构

```sql
CREATE TABLE admission_data (
    id UInt64,                           -- 自增ID
    year UInt32,                         -- 年份
    province String,                     -- 省份  
    college_name String,                 -- 院校名称
    college_code String,                 -- 院校代码
    special_interest_group_code String,  -- 专业组代码
    professional_name String,            -- 专业名称
    class_demand String,                 -- 选科要求
    lowest_points Int64,                 -- 录取最低分
    lowest_rank Int64,                   -- 录取最低位次
    description String                   -- 备注
) ENGINE = MergeTree()
ORDER BY (year, lowest_points, lowest_rank)
```

## 选科组合编码

| 科目 | 编码 |
|------|------|
| 物理 | 1    |
| 化学 | 2    |
| 生物 | 3    |
| 政治 | 4    |
| 历史 | 5    |
| 地理 | 6    |

示例：
- 物理+化学+生物 = "123"
- 历史+政治+地理 = "456"

## 故障排除

### 常见问题

1. **连接ClickHouse失败**
   - 检查ClickHouse服务是否运行: `sudo systemctl status clickhouse-server`
   - 检查端口是否开放: `netstat -tlnp | grep 9000`

2. **数据导入失败**
   - 确保Excel文件存在且格式正确
   - 检查数据库连接
   - 查看错误日志获取详细信息
   - 参考 [DATA_SETUP.md](DATA_SETUP.md) 解决数据文件问题

3. **API响应慢**
   - 检查ClickHouse查询性能
   - 考虑添加索引或优化查询语句

## 相关文档

- [数据文件设置说明](DATA_SETUP.md) - 如何获取和配置Excel数据文件
- [部署指南](DEPLOYMENT.md) - 详细的部署步骤和故障排除

## 许可证

[MIT License](LICENSE)

---

如有问题请提交Issue或联系维护者。
