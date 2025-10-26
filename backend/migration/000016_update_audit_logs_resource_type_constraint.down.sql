-- 回滚：恢复 audit_logs 表的 resource_type 约束

-- 恢复原有的约束
ALTER TABLE audit_logs
ADD CONSTRAINT chk_audit_logs_resource_type CHECK (
    resource_type IN ('user', 'admin', 'role', 'department', 'job_position', 'resume', 'screening', 'setting', 'attachment', 'conversation', 'message')
);

-- 移除注释
COMMENT ON COLUMN audit_logs.resource_type IS NULL;