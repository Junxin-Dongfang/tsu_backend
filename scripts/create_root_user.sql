-- 创建 root 用户（用于测试，不使用 Kratos）
-- 注意：这只是创建数据库记录，实际认证需要通过 Kratos 或游戏系统的认证接口
INSERT INTO auth.users (id, username, email, nickname, is_banned, login_count, created_at, updated_at) 
VALUES (uuid_generate_v4(), 'root', 'root@tsu-game.com', 'Administrator', false, 0, NOW(), NOW()) 
ON CONFLICT (username) DO NOTHING;

-- 显示创建的用户
SELECT id, username, email, nickname, created_at FROM auth.users WHERE username = 'root';
