-- 创建 screening_node_runs 表
CREATE TABLE screening_node_runs (
    "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "task_id" UUID NOT NULL,
    "task_resume_id" UUID NOT NULL,
    "node_key" VARCHAR(50),
    "status" VARCHAR(20) NOT NULL DEFAULT 'pending',
    "attempt_no" INTEGER NOT NULL DEFAULT 1,
    "trace_id" VARCHAR(100),
    "agent_version" VARCHAR(50),
    "model_name" VARCHAR(100),
    "model_provider" VARCHAR(50),
    "llm_params" JSONB,
    "input_payload" JSONB,
    "output_payload" JSONB,
    "error_message" TEXT,
    "tokens_input" INTEGER DEFAULT 0,
    "tokens_output" INTEGER DEFAULT 0,
    "total_cost" DECIMAL(10, 6) DEFAULT 0.0,
    "started_at" TIMESTAMP WITH TIME ZONE,
    "finished_at" TIMESTAMP WITH TIME ZONE,
    "duration_ms" INTEGER DEFAULT 0,
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL DEFAULT now(),
    "deleted_at" timestamptz NULL
);
-- 添加外键约束
ALTER TABLE screening_node_runs
ADD CONSTRAINT fk_screening_node_runs_task_id FOREIGN KEY (task_id) REFERENCES screening_tasks(id) ON DELETE CASCADE;
ALTER TABLE screening_node_runs
ADD CONSTRAINT fk_screening_node_runs_task_resume_id FOREIGN KEY (task_resume_id) REFERENCES screening_task_resumes(id) ON DELETE CASCADE;
-- 创建索引
CREATE INDEX idx_screening_node_runs_task_id ON screening_node_runs(task_id);
CREATE INDEX idx_screening_node_runs_task_resume_id ON screening_node_runs(task_resume_id);
CREATE INDEX idx_screening_node_runs_node_key ON screening_node_runs(node_key);
CREATE INDEX idx_screening_node_runs_status ON screening_node_runs(status);
CREATE INDEX idx_screening_node_runs_trace_id ON screening_node_runs(trace_id);
CREATE INDEX idx_screening_node_runs_created_at ON screening_node_runs(created_at);
-- 创建复合索引
CREATE INDEX idx_screening_node_runs_task_resume_node ON screening_node_runs(task_resume_id, node_key);
CREATE INDEX idx_screening_node_runs_task_status ON screening_node_runs(task_id, status);
-- 添加状态检查约束
ALTER TABLE screening_node_runs
ADD CONSTRAINT chk_screening_node_runs_status CHECK (
        status IN ('pending', 'running', 'completed', 'failed')
    );
-- 添加 attempt_no 检查约束
ALTER TABLE screening_node_runs
ADD CONSTRAINT chk_screening_node_runs_attempt_no CHECK (attempt_no > 0);
-- 添加 tokens 检查约束
ALTER TABLE screening_node_runs
ADD CONSTRAINT chk_screening_node_runs_tokens_input CHECK (tokens_input >= 0);
ALTER TABLE screening_node_runs
ADD CONSTRAINT chk_screening_node_runs_tokens_output CHECK (tokens_output >= 0);
-- 添加 total_cost 检查约束
ALTER TABLE screening_node_runs
ADD CONSTRAINT chk_screening_node_runs_total_cost CHECK (total_cost >= 0);
-- 添加 duration_ms 检查约束
ALTER TABLE screening_node_runs
ADD CONSTRAINT chk_screening_node_runs_duration_ms CHECK (duration_ms >= 0);
-- 添加注释
COMMENT ON TABLE screening_node_runs IS '筛选节点运行记录表';
COMMENT ON COLUMN screening_node_runs.id IS '主键ID';
COMMENT ON COLUMN screening_node_runs.task_id IS '筛选任务ID';
COMMENT ON COLUMN screening_node_runs.task_resume_id IS '任务简历ID';
COMMENT ON COLUMN screening_node_runs.node_key IS '节点类型，参考 domain.TaskMetaDataNode 等常量';
COMMENT ON COLUMN screening_node_runs.status IS '运行状态: pending/running/completed/failed';
COMMENT ON COLUMN screening_node_runs.attempt_no IS '第几次尝试';
COMMENT ON COLUMN screening_node_runs.trace_id IS '链路追踪ID';
COMMENT ON COLUMN screening_node_runs.agent_version IS 'Agent版本号';
COMMENT ON COLUMN screening_node_runs.model_name IS '使用的模型名称';
COMMENT ON COLUMN screening_node_runs.model_provider IS '模型提供商';
COMMENT ON COLUMN screening_node_runs.llm_params IS 'LLM参数配置(JSON)';
COMMENT ON COLUMN screening_node_runs.input_payload IS '输入数据(JSON)';
COMMENT ON COLUMN screening_node_runs.output_payload IS '输出数据(JSON)';
COMMENT ON COLUMN screening_node_runs.error_message IS '错误信息';
COMMENT ON COLUMN screening_node_runs.tokens_input IS '输入token数';
COMMENT ON COLUMN screening_node_runs.tokens_output IS '输出token数';
COMMENT ON COLUMN screening_node_runs.total_cost IS '总成本';
COMMENT ON COLUMN screening_node_runs.started_at IS '开始时间';
COMMENT ON COLUMN screening_node_runs.finished_at IS '结束时间';
COMMENT ON COLUMN screening_node_runs.duration_ms IS '执行时长(毫秒)';
COMMENT ON COLUMN screening_node_runs.created_at IS '创建时间';
COMMENT ON COLUMN screening_node_runs.updated_at IS '更新时间';
COMMENT ON COLUMN screening_node_runs.deleted_at IS '软删除时间';

-- 创建 audit_logs 表
CREATE TABLE audit_logs (
    "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "operator_type" VARCHAR(20) NOT NULL,
    "operator_id" UUID,
    "operator_name" VARCHAR(255),
    "operation_type" VARCHAR(20) NOT NULL,
    "resource_type" VARCHAR(50) NOT NULL,
    "resource_id" VARCHAR(255),
    "resource_name" VARCHAR(255),
    "method" VARCHAR(10) NOT NULL,
    "path" VARCHAR(500) NOT NULL,
    "query_params" TEXT,
    "request_body" TEXT,
    "user_agent" TEXT,
    "status_code" INTEGER NOT NULL,
    "status" VARCHAR(20) NOT NULL DEFAULT 'success',
    "response_body" TEXT,
    "error_message" TEXT,
    "ip" VARCHAR(45) NOT NULL,
    "country" VARCHAR(100),
    "province" VARCHAR(100),
    "city" VARCHAR(100),
    "isp" VARCHAR(100),
    "session_id" VARCHAR(255),
    "trace_id" VARCHAR(255),
    "business_data" TEXT,
    "changes" TEXT,
    "description" TEXT,
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL DEFAULT now(),
    "deleted_at" timestamptz NULL,
    "duration_ms" BIGINT
);

-- 创建索引
CREATE INDEX idx_audit_logs_operator_type_id ON audit_logs(operator_type, operator_id);
CREATE INDEX idx_audit_logs_operator_type_created_at ON audit_logs(operator_type, created_at);
CREATE INDEX idx_audit_logs_operation_type_created_at ON audit_logs(operation_type, created_at);
CREATE INDEX idx_audit_logs_resource_type_created_at ON audit_logs(resource_type, created_at);
CREATE INDEX idx_audit_logs_resource_type_id ON audit_logs(resource_type, resource_id);
CREATE INDEX idx_audit_logs_status_created_at ON audit_logs(status, created_at);
CREATE INDEX idx_audit_logs_status_code_created_at ON audit_logs(status_code, created_at);
CREATE INDEX idx_audit_logs_ip_created_at ON audit_logs(ip, created_at);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
CREATE INDEX idx_audit_logs_session_id ON audit_logs(session_id);
CREATE INDEX idx_audit_logs_trace_id ON audit_logs(trace_id);

-- 创建复合索引用于常见查询场景
CREATE INDEX idx_audit_logs_operator_operation_created_at ON audit_logs(operator_type, operation_type, created_at);
CREATE INDEX idx_audit_logs_resource_operation_created_at ON audit_logs(resource_type, operation_type, created_at);

-- 添加检查约束
ALTER TABLE audit_logs
ADD CONSTRAINT chk_audit_logs_operator_type CHECK (
    operator_type IN ('user', 'admin')
);

ALTER TABLE audit_logs
ADD CONSTRAINT chk_audit_logs_operation_type CHECK (
    operation_type IN ('create', 'update', 'delete', 'view', 'login', 'logout')
);

ALTER TABLE audit_logs
ADD CONSTRAINT chk_audit_logs_resource_type CHECK (
    resource_type IN ('user', 'admin', 'role', 'department', 'job_position', 'resume', 'screening', 'setting', 'attachment', 'conversation', 'message')
);

ALTER TABLE audit_logs
ADD CONSTRAINT chk_audit_logs_status CHECK (
    status IN ('success', 'failed')
);

ALTER TABLE audit_logs
ADD CONSTRAINT chk_audit_logs_status_code CHECK (
    status_code >= 100 AND status_code < 600
);

ALTER TABLE audit_logs
ADD CONSTRAINT chk_audit_logs_duration_ms CHECK (
    duration_ms >= 0
);

-- 添加注释
COMMENT ON TABLE audit_logs IS '操作审计日志表';
COMMENT ON COLUMN audit_logs.id IS '主键ID';
COMMENT ON COLUMN audit_logs.operator_type IS '操作者类型：user/admin';
COMMENT ON COLUMN audit_logs.operator_id IS '操作者ID';
COMMENT ON COLUMN audit_logs.operator_name IS '操作者名称';
COMMENT ON COLUMN audit_logs.operation_type IS '操作类型：create/update/delete/view/login/logout';
COMMENT ON COLUMN audit_logs.resource_type IS '资源类型';
COMMENT ON COLUMN audit_logs.resource_id IS '资源ID';
COMMENT ON COLUMN audit_logs.resource_name IS '资源名称';
COMMENT ON COLUMN audit_logs.method IS 'HTTP方法';
COMMENT ON COLUMN audit_logs.path IS '请求路径';
COMMENT ON COLUMN audit_logs.query_params IS '查询参数';
COMMENT ON COLUMN audit_logs.request_body IS '请求体（敏感信息已脱敏）';
COMMENT ON COLUMN audit_logs.user_agent IS '用户代理';
COMMENT ON COLUMN audit_logs.status_code IS 'HTTP状态码';
COMMENT ON COLUMN audit_logs.status IS '操作状态：success/failed';
COMMENT ON COLUMN audit_logs.response_body IS '响应体（敏感信息已脱敏）';
COMMENT ON COLUMN audit_logs.error_message IS '错误信息';
COMMENT ON COLUMN audit_logs.ip IS '客户端IP';
COMMENT ON COLUMN audit_logs.country IS '国家';
COMMENT ON COLUMN audit_logs.province IS '省份';
COMMENT ON COLUMN audit_logs.city IS '城市';
COMMENT ON COLUMN audit_logs.isp IS '运营商';
COMMENT ON COLUMN audit_logs.session_id IS '会话ID';
COMMENT ON COLUMN audit_logs.trace_id IS '链路追踪ID';
COMMENT ON COLUMN audit_logs.business_data IS '业务相关数据（JSON格式）';
COMMENT ON COLUMN audit_logs.changes IS '变更内容（JSON格式，记录before/after）';
COMMENT ON COLUMN audit_logs.description IS '操作描述';
COMMENT ON COLUMN audit_logs.created_at IS '创建时间';
COMMENT ON COLUMN audit_logs.updated_at IS '更新时间';
COMMENT ON COLUMN audit_logs.deleted_at IS '软删除时间';
COMMENT ON COLUMN audit_logs.duration_ms IS '请求处理时长（毫秒）';