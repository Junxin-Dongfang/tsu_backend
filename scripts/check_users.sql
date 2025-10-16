-- 检查数据库用户
SELECT rolname, rolcanlogin FROM pg_roles WHERE rolname LIKE 'tsu%';

-- 重新设置用户密码（确保与 .env 文件中的密码一致）
ALTER USER tsu_auth_user WITH PASSWORD 'tsu_auth_password';
ALTER USER tsu_game_user WITH PASSWORD 'tsu_game_password';
ALTER USER tsu_admin_user WITH PASSWORD 'tsu_admin_password';

-- 确认修改
SELECT rolname, rolcanlogin FROM pg_roles WHERE rolname LIKE 'tsu%';
