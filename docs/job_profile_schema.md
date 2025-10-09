
# 招聘系统岗位画像数据库设计 (PostgreSQL, UUID 主键)

本文档描述了岗位画像功能的数据库表设计。主键统一采用 `UUID`。

---

## 1. 部门表 `department`
存储公司部门信息。

```sql
CREATE TABLE department (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);
```

---

## 2. 岗位表 `job_position`
存储岗位的基本信息。

```sql
CREATE TABLE job_position (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,                -- 岗位名称
    department_id UUID NOT NULL REFERENCES department(id),
    location VARCHAR(200),                     -- 工作地点
    salary_min NUMERIC(12,2),                  -- 最低薪资
    salary_max NUMERIC(12,2),                  -- 最高薪资
    description TEXT,                          -- 岗位描述
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);
```

---

## 3. 岗位职责表 `job_responsibility`
存储岗位的多条职责。

```sql
CREATE TABLE job_responsibility (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id UUID NOT NULL REFERENCES job_position(id) ON DELETE CASCADE,
    responsibility TEXT NOT NULL,              -- 职责描述
    sort_order INT DEFAULT 0                   -- 排序
);
```

---

## 4. 技能表 `skill`
存储技能字典（避免岗位里写死字符串）。

```sql
CREATE TABLE skill (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,         -- 技能名称，如 Java, React
    created_at TIMESTAMP DEFAULT now()
);
```

---

## 5. 岗位技能表 `job_skill`
存储岗位所需技能（必备技能、加分技能 + 权重）。

```sql
CREATE TYPE skill_type AS ENUM ('required', 'bonus');

CREATE TABLE job_skill (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id UUID NOT NULL REFERENCES job_position(id) ON DELETE CASCADE,
    skill_id UUID NOT NULL REFERENCES skill(id),
    type skill_type NOT NULL,                  -- required / bonus
    weight INT CHECK (weight >= 0 AND weight <= 100),  -- 权重 %
    UNIQUE(job_id, skill_id, type)
);
```

---

## 6. 学历要求表 `job_education_requirement`
存储岗位对学历的要求和权重。

```sql
CREATE TABLE job_education_requirement (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id UUID NOT NULL REFERENCES job_position(id) ON DELETE CASCADE,
    min_degree VARCHAR(50) NOT NULL,           -- 本科/硕士/博士
    weight INT NOT NULL CHECK (weight >= 0 AND weight <= 100)
);
```

---

## 7. 工作经验要求表 `job_experience_requirement`
存储岗位对工作经验的要求和权重。

```sql
CREATE TABLE job_experience_requirement (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id UUID NOT NULL REFERENCES job_position(id) ON DELETE CASCADE,
    min_years INT DEFAULT 0,                   -- 最少工作年限
    ideal_years INT DEFAULT 0,                 -- 理想年限
    weight INT NOT NULL CHECK (weight >= 0 AND weight <= 100)
);
```

---

## 8. 行业/公司要求表 `job_industry_requirement`
存储岗位对候选人所在行业/公司要求及权重。

```sql
CREATE TABLE job_industry_requirement (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id UUID NOT NULL REFERENCES job_position(id) ON DELETE CASCADE,
    industry VARCHAR(100) NOT NULL,            -- 行业，如互联网/金融/制造业
    company_name VARCHAR(200),                 -- 可选，指定公司名称
    weight INT NOT NULL CHECK (weight >= 0 AND weight <= 100),
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);
```

---

## 设计要点

- **岗位画像主体是 `job_position`**，其余表都通过 `job_id` 关联。  
- **职责 / 技能 / 学历 / 经验 / 行业** 都拆到子表，方便扩展。  
- 权重字段设计成 `INT` (0-100)，校验逻辑由应用层保证合理性（如总和 <= 100%）。  
- 使用 `UUID` 作为主键，避免分布式下冲突。  

---

## 示例数据

```sql
-- 插入部门
INSERT INTO department (name) VALUES ('产品部');

-- 插入岗位
INSERT INTO job_position (name, department_id, location, salary_min, salary_max, description)
VALUES ('高级产品经理', 'uuid-dept-1', '上海', 15000, 25000, '负责产品规划与执行');

-- 插入岗位职责
INSERT INTO job_responsibility (job_id, responsibility, sort_order)
VALUES ('uuid-job-1', '负责产品需求分析', 1);

-- 插入技能
INSERT INTO skill (name) VALUES ('Java'), ('React'), ('Docker'), ('Kubernetes');

-- 插入岗位技能
INSERT INTO job_skill (job_id, skill_id, type, weight)
VALUES ('uuid-job-1', 'uuid-skill-java', 'required', 30);

-- 插入学历要求
INSERT INTO job_education_requirement (job_id, min_degree, weight)
VALUES ('uuid-job-1', '本科及以上', 30);

-- 插入经验要求
INSERT INTO job_experience_requirement (job_id, min_years, ideal_years, weight)
VALUES ('uuid-job-1', 3, 5, 40);

-- 插入行业/公司要求
INSERT INTO job_industry_requirement (job_id, industry, company_name, weight)
VALUES ('uuid-job-1', '互联网', '阿里巴巴', 20);
```
