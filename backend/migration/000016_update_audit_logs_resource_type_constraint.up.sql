-- 移除 audit_logs 表的 resource_type 约束检查
-- 原因：为了提高系统扩展性，避免每次新增资源类型都需要修改数据库约束

-- 删除 resource_type 约束
ALTER TABLE audit_logs DROP CONSTRAINT IF EXISTS chk_audit_logs_resource_type;

-- 添加注释说明
COMMENT ON COLUMN audit_logs.resource_type IS '资源类型：由应用层代码控制有效值，不在数据库层面做约束检查';