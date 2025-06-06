# 数据文件设置说明

## 问题说明

由于Excel数据文件 `21-24各省份录取数据(含专业组代码).xlsx` 大小为112MB，超过了GitHub的100MB文件大小限制，因此该文件已从Git仓库中移除。

## 解决方案

### 1. 本地开发

如果你已经有数据文件：
1. 将 `21-24各省份录取数据(含专业组代码).xlsx` 放在项目根目录
2. 运行 `make import` 导入数据

### 2. 获取数据文件

如果你没有数据文件，可以通过以下方式获取：

#### 方案一：从原始来源获取
- 联系项目维护者获取数据文件
- 从官方教育部门或相关机构获取高考录取数据

#### 方案二：使用Git LFS（推荐）
如果你有权限配置仓库，可以使用Git LFS：

```bash
# 安装Git LFS
git lfs install

# 追踪Excel文件
git lfs track "*.xlsx"

# 添加并提交
git add .gitattributes
git add "21-24各省份录取数据(含专业组代码).xlsx"
git commit -m "使用Git LFS管理大文件"
git push
```

#### 方案三：外部存储
将数据文件上传到云存储服务（如OSS、S3等），然后在部署脚本中下载：

```bash
# 在部署脚本中添加下载步骤
wget "https://your-storage-url/21-24各省份录取数据(含专业组代码).xlsx"
```

### 3. 数据文件格式要求

Excel文件应包含以下列（按顺序）：
1. 年份 (Year)
2. 省份 (Province) 
3. 院校名称 (College Name)
4. 院校代码 (College Code)
5. 专业组代码 (Special Interest Group Code)
6. 专业名称 (Professional Name)
7. 选科要求 (Class Demand)
8. 录取最低分 (Lowest Points)
9. 录取最低位次 (Lowest Rank)
10. 备注 (Description)

### 4. 替代数据源

如果无法获取原始数据文件，可以：

1. **创建测试数据**：
   ```bash
   # 运行测试数据生成器（如果有）
   go run tools/generate_test_data.go
   ```

2. **使用公开数据**：
   - 从教育部官网获取历年高考录取数据
   - 从各省教育考试院获取本省数据
   - 整理成相同格式的Excel文件

3. **分批导入数据**：
   - 将大文件拆分成多个小文件
   - 分别导入到数据库中

## 部署注意事项

### 远程部署时

如果使用自动部署脚本 `make deploy`，需要确保：

1. **手动上传数据文件**：
   ```bash
   scp "21-24各省份录取数据(含专业组代码).xlsx" username@server:/opt/gaokao-zhiyuan/
   ```

2. **修改部署脚本**：
   编辑 `scripts/deploy.sh`，在数据导入之前添加数据文件检查：
   ```bash
   # 检查数据文件是否存在
   if [ ! -f "21-24各省份录取数据(含专业组代码).xlsx" ]; then
       echo "错误: 数据文件不存在，请先上传数据文件"
       exit 1
   fi
   ```

3. **或者从远程下载**：
   在部署脚本中添加下载步骤：
   ```bash
   # 下载数据文件
   wget -O "21-24各省份录取数据(含专业组代码).xlsx" "https://your-data-source-url/data.xlsx"
   ```

## 验证数据导入

导入完成后，可以通过以下方式验证：

```bash
# 连接ClickHouse
clickhouse-client

# 检查数据量
SELECT count() FROM gaokao.admission_data;

# 检查年份分布
SELECT year, count() FROM gaokao.admission_data GROUP BY year ORDER BY year;

# 检查数据样例
SELECT * FROM gaokao.admission_data LIMIT 5;
```

## 故障排除

如果导入失败：

1. **检查文件格式**：确保Excel文件格式正确
2. **检查文件路径**：确保文件在正确的位置
3. **检查权限**：确保程序有读取文件的权限
4. **查看日志**：检查导入工具的错误日志

---

如有问题，请联系项目维护者或查阅相关文档。 