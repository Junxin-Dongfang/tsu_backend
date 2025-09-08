-- 启用所需扩展
-- 注意：需要具备在当前数据库创建扩展的权限

CREATE EXTENSION IF NOT EXISTS pgcrypto; -- 提供 gen_random_uuid()

-- 如需使用 uuid_generate_v4()，可启用 uuid-ossp 扩展
-- CREATE EXTENSION IF NOT EXISTS "uuid-ossp";



