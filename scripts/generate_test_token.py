#!/usr/bin/env python3
"""
生成测试用的 JWT Token
用于绕过 Kratos 认证，直接测试游戏功能
"""

import jwt
import datetime
import sys

# JWT 配置（从 .env 文件中读取，这里使用默认值）
JWT_SECRET = "your-super-secret-jwt-key-change-in-production"
JWT_ALGORITHM = "HS256"

def generate_token(user_id, username="root", email="root@tsu-game.com"):
    """生成 JWT token"""
    
    # Token 过期时间（30分钟）
    expiration = datetime.datetime.utcnow() + datetime.timedelta(minutes=30)
    
    # JWT payload
    payload = {
        "user_id": user_id,
        "username": username,
        "email": email,
        "exp": expiration,
        "iat": datetime.datetime.utcnow()
    }
    
    # 生成 token
    token = jwt.encode(payload, JWT_SECRET, algorithm=JWT_ALGORITHM)
    
    return token

if __name__ == "__main__":
    if len(sys.argv) > 1:
        user_id = sys.argv[1]
    else:
        # 默认使用 root 用户 ID（需要从数据库查询）
        print("Usage: python generate_test_token.py <user_id>")
        print("Getting root user ID from database...")
        
        import psycopg2
        conn = psycopg2.connect(
            host="localhost",
            port=5432,
            database="tsu_db",
            user="postgres",
            password="postgres"
        )
        cursor = conn.cursor()
        cursor.execute("SELECT id FROM auth.users WHERE username = 'root' LIMIT 1")
        result = cursor.fetchone()
        
        if result:
            user_id = str(result[0])
            print(f"Found root user ID: {user_id}")
        else:
            print("Error: root user not found in database")
            sys.exit(1)
        
        cursor.close()
        conn.close()
    
    token = generate_token(user_id)
    
    print("\n" + "="*60)
    print("Generated JWT Token:")
    print("="*60)
    print(token)
    print("="*60)
    print(f"\nUser ID: {user_id}")
    print(f"Expires in: 30 minutes")
    print("\nUse this token in Authorization header:")
    print(f"Authorization: Bearer {token}")
    print("="*60)
