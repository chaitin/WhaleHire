# SQL 迁移脚本生成工具

基于当前 Ent schema 自动生成 SQL 迁移脚本的工具。

> 💡 不可完全依赖该生成工具，目前该工具并不完善，会出现问题，需要人工检查核实

````

## 使用方法

```bash
# 在 backend 目录下执行
go run script/create_migration.go <migration_name>
````

## 功能特性

### 自动化生成

- 自动获取下一个迁移版本号
- 基于当前 Ent schema 生成完整的 SQL DDL
- 生成格式化的、可读性强的 SQL 语句
- 自动创建 up 和 down 迁移文件对

### 文件输出

生成的文件保存在 `migration/` 目录下：

- `XXXXXX_<migration_name>.up.sql` - 包含格式化的数据库 schema
- `XXXXXX_<migration_name>.down.sql` - 包含回滚注释模板

### SQL 格式化

生成的 SQL 具有良好的可读性：

- CREATE TABLE 语句按列分行显示
- 适当的缩进和换行
- PRIMARY KEY 和 FOREIGN KEY 约束清晰格式化
- CREATE INDEX 语句结构化显示

## 使用示例

```bash
# 生成初始数据库 schema
go run script/create_migration.go init_schema

# 生成用户相关表的迁移
go run script/create_migration.go add_user_tables
```

生成的文件示例：

```
migration/000001_init_schema.up.sql
migration/000001_init_schema.down.sql
```

## 注意事项

- **down 文件需要手动编写**：工具只在 down 文件中生成注释模板，回滚 SQL 需手动添加
- **完整 schema 生成**：工具生成当前完整的数据库 schema，适用于初始化或重建场景
- **路径自动处理**：使用绝对路径确保在任何目录下运行都能正确生成文件

## 兼容性

生成的迁移文件与 golang-migrate 系统完全兼容，可直接用于数据库迁移管理。
