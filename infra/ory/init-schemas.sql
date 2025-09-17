-- 创建 kratos 和 keto 的 schema
CREATE SCHEMA IF NOT EXISTS kratos;
CREATE SCHEMA IF NOT EXISTS keto;

-- 给 ory_user 授予这些 schema 的权限
GRANT ALL PRIVILEGES ON SCHEMA kratos TO ory_user;
GRANT ALL PRIVILEGES ON SCHEMA keto TO ory_user;

-- 设置默认权限，确保在这些 schema 中创建的表也有正确的权限
ALTER DEFAULT PRIVILEGES IN SCHEMA kratos GRANT ALL ON TABLES TO ory_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA keto GRANT ALL ON TABLES TO ory_user;