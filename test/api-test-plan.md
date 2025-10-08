# Admin 服务接口全面验证计划

## 📋 测试概览

**服务地址**: http://localhost:80  
**测试账号**: root  
**测试密码**: password  
**API 版本**: v1  
**认证方式**: Bearer Token (通过 Nginx + Oathkeeper)  

---

## 🎯 测试目标

1. **功能完整性**: 验证所有接口功能正常工作
2. **权限验证**: 确认认证和授权机制正确
3. **数据一致性**: 验证 CRUD 操作的数据完整性
4. **错误处理**: 测试各类边界条件和异常情况
5. **业务逻辑**: 验证关联关系和复杂业务流程

---

## 📊 接口清单（共 100+ 个接口）

### 1. 认证接口 (4个) - 公开访问
| 方法 | 路径 | 功能 | 优先级 |
|-----|------|------|--------|
| POST | `/api/v1/auth/register` | 用户注册 | P0 |
| POST | `/api/v1/auth/login` | 用户登录 | P0 |
| POST | `/api/v1/auth/logout` | 用户登出 | P0 |
| GET | `/api/v1/auth/users/:user_id` | 获取用户信息 | P1 |

### 2. 用户管理 (7个) - 需要认证
| 方法 | 路径 | 功能 | 优先级 |
|-----|------|------|--------|
| GET | `/api/v1/admin/users/me` | 获取当前用户信息 | P0 |
| GET | `/api/v1/admin/users` | 获取用户列表 | P0 |
| GET | `/api/v1/admin/users/:id` | 获取用户详情 | P1 |
| PUT | `/api/v1/admin/users/:id` | 更新用户信息 | P1 |
| POST | `/api/v1/admin/users/:id/ban` | 封禁用户 | P1 |
| POST | `/api/v1/admin/users/:id/unban` | 解封用户 | P1 |

### 3. 角色管理 (6个)
| 方法 | 路径 | 功能 | 优先级 |
|-----|------|------|--------|
| GET | `/api/v1/admin/roles` | 获取角色列表 | P0 |
| POST | `/api/v1/admin/roles` | 创建角色 | P1 |
| PUT | `/api/v1/admin/roles/:id` | 更新角色 | P1 |
| DELETE | `/api/v1/admin/roles/:id` | 删除角色 | P1 |
| GET | `/api/v1/admin/roles/:id/permissions` | 获取角色权限 | P1 |
| POST | `/api/v1/admin/roles/:id/permissions` | 分配权限给角色 | P1 |

### 4. 权限管理 (2个)
| 方法 | 路径 | 功能 | 优先级 |
|-----|------|------|--------|
| GET | `/api/v1/admin/permissions` | 获取权限列表 | P0 |
| GET | `/api/v1/admin/permission-groups` | 获取权限组列表 | P1 |

### 5. 用户-角色关联 (3个)
| 方法 | 路径 | 功能 | 优先级 |
|-----|------|------|--------|
| GET | `/api/v1/admin/users/:user_id/roles` | 获取用户角色 | P1 |
| POST | `/api/v1/admin/users/:user_id/roles` | 分配角色给用户 | P1 |
| DELETE | `/api/v1/admin/users/:user_id/roles` | 撤销用户角色 | P1 |

### 6. 用户-权限管理 (1个)
| 方法 | 路径 | 功能 | 优先级 |
|-----|------|------|--------|
| GET | `/api/v1/admin/users/:user_id/permissions` | 获取用户权限 | P1 |

### 7. 基础游戏配置 (24个) - 8类配置 × 3操作
每类配置包含：GET (列表), POST (创建), GET/:id (详情), PUT/:id (更新), DELETE/:id (删除)

| 配置类型 | 路径前缀 | 数量 | 优先级 |
|---------|---------|------|--------|
| 职业 | `/api/v1/admin/classes` | 3 | P0 |
| 技能分类 | `/api/v1/admin/skill-categories` | 3 | P0 |
| 动作分类 | `/api/v1/admin/action-categories` | 3 | P1 |
| 伤害类型 | `/api/v1/admin/damage-types` | 3 | P1 |
| 英雄属性类型 | `/api/v1/admin/hero-attribute-types` | 3 | P0 |
| 标签 | `/api/v1/admin/tags` | 3 | P1 |
| 标签关系 | `/api/v1/admin/tag-relations` | 3 | P2 |
| 动作标记 | `/api/v1/admin/action-flags` | 3 | P2 |

### 8. 元数据定义 (12个) - 4类定义 × 3操作
只读配置，每类包含：GET (分页), GET/all (全部), GET/:id (详情)

| 定义类型 | 路径前缀 | 数量 | 优先级 |
|---------|---------|------|--------|
| 效果类型定义 | `/api/v1/admin/metadata/effect-type-definitions` | 3 | P1 |
| 公式变量 | `/api/v1/admin/metadata/formula-variables` | 3 | P1 |
| 范围配置规则 | `/api/v1/admin/metadata/range-config-rules` | 3 | P1 |
| 动作类型定义 | `/api/v1/admin/metadata/action-type-definitions` | 3 | P1 |

### 9. 技能系统 (10个)
| 方法 | 路径 | 功能 | 优先级 |
|-----|------|------|--------|
| GET | `/api/v1/admin/skills` | 获取技能列表 | P0 |
| POST | `/api/v1/admin/skills` | 创建技能 | P0 |
| GET | `/api/v1/admin/skills/:id` | 获取技能详情 | P0 |
| PUT | `/api/v1/admin/skills/:id` | 更新技能 | P0 |
| DELETE | `/api/v1/admin/skills/:id` | 删除技能 | P1 |
| GET | `/api/v1/admin/skills/:id/level-configs` | 获取技能等级配置 | P0 |
| POST | `/api/v1/admin/skills/:id/level-configs` | 创建技能等级配置 | P0 |
| GET | `/api/v1/admin/skills/:id/level-configs/:config_id` | 获取等级配置详情 | P1 |
| PUT | `/api/v1/admin/skills/:id/level-configs/:config_id` | 更新等级配置 | P1 |
| DELETE | `/api/v1/admin/skills/:id/level-configs/:config_id` | 删除等级配置 | P1 |

### 10. 效果系统 (18个)
**Effects (5个)**
| 方法 | 路径 | 功能 | 优先级 |
|-----|------|------|--------|
| GET | `/api/v1/admin/effects` | 获取效果列表 | P0 |
| POST | `/api/v1/admin/effects` | 创建效果 | P0 |
| GET | `/api/v1/admin/effects/:id` | 获取效果详情 | P1 |
| PUT | `/api/v1/admin/effects/:id` | 更新效果 | P1 |
| DELETE | `/api/v1/admin/effects/:id` | 删除效果 | P1 |

**Buffs (5个)**
| 方法 | 路径 | 功能 | 优先级 |
|-----|------|------|--------|
| GET | `/api/v1/admin/buffs` | 获取Buff列表 | P0 |
| POST | `/api/v1/admin/buffs` | 创建Buff | P0 |
| GET | `/api/v1/admin/buffs/:id` | 获取Buff详情 | P1 |
| PUT | `/api/v1/admin/buffs/:id` | 更新Buff | P1 |
| DELETE | `/api/v1/admin/buffs/:id` | 删除Buff | P1 |

**Buff-Effects 关联 (4个)**
| 方法 | 路径 | 功能 | 优先级 |
|-----|------|------|--------|
| GET | `/api/v1/admin/buffs/:buff_id/effects` | 获取Buff关联的效果 | P1 |
| POST | `/api/v1/admin/buffs/:buff_id/effects` | 添加效果到Buff | P1 |
| POST | `/api/v1/admin/buffs/:buff_id/effects/batch` | 批量设置Buff效果 | P2 |
| DELETE | `/api/v1/admin/buffs/:buff_id/effects/:effect_id` | 移除Buff效果 | P2 |

### 11. 动作系统 (13个)
**Actions (5个)**
| 方法 | 路径 | 功能 | 优先级 |
|-----|------|------|--------|
| GET | `/api/v1/admin/actions` | 获取动作列表 | P0 |
| POST | `/api/v1/admin/actions` | 创建动作 | P0 |
| GET | `/api/v1/admin/actions/:id` | 获取动作详情 | P1 |
| PUT | `/api/v1/admin/actions/:id` | 更新动作 | P1 |
| DELETE | `/api/v1/admin/actions/:id` | 删除动作 | P1 |

**Action-Effects 关联 (4个)**
| 方法 | 路径 | 功能 | 优先级 |
|-----|------|------|--------|
| GET | `/api/v1/admin/actions/:action_id/effects` | 获取动作效果 | P1 |
| POST | `/api/v1/admin/actions/:action_id/effects` | 添加效果到动作 | P1 |
| POST | `/api/v1/admin/actions/:action_id/effects/batch` | 批量设置动作效果 | P2 |
| DELETE | `/api/v1/admin/actions/:action_id/effects/:effect_id` | 移除动作效果 | P2 |

**Skill-Unlock-Actions 关联 (4个)**
| 方法 | 路径 | 功能 | 优先级 |
|-----|------|------|--------|
| GET | `/api/v1/admin/skills/:skill_id/unlock-actions` | 获取技能解锁的动作 | P1 |
| POST | `/api/v1/admin/skills/:skill_id/unlock-actions` | 添加解锁动作 | P1 |
| POST | `/api/v1/admin/skills/:skill_id/unlock-actions/batch` | 批量设置解锁动作 | P2 |
| DELETE | `/api/v1/admin/skills/:skill_id/unlock-actions/:action_id` | 移除解锁动作 | P2 |

### 12. 系统接口 (2个)
| 方法 | 路径 | 功能 | 优先级 |
|-----|------|------|--------|
| GET | `/health` | 健康检查 | P0 |
| GET | `/swagger/*` | Swagger 文档 | P0 |

---

## 🧪 测试策略

### 阶段 1: 基础功能验证 (P0 优先级)
**目标**: 确保核心功能可用

1. **系统可用性**
   - [ ] 健康检查接口
   - [ ] Swagger 文档可访问

2. **认证流程**
   - [ ] 用户登录成功
   - [ ] Token 获取和验证
   - [ ] 认证失败场景

3. **核心 CRUD**
   - [ ] 用户管理基础操作
   - [ ] 角色权限查询
   - [ ] 游戏配置基础查询

### 阶段 2: 权限系统验证 (P1 优先级)
**目标**: 验证 RBAC 权限体系

1. **角色管理**
   - [ ] 角色 CRUD 完整流程
   - [ ] 角色-权限关联
   - [ ] 权限继承验证

2. **用户权限**
   - [ ] 用户-角色关联
   - [ ] 权限检查机制
   - [ ] 权限组功能

### 阶段 3: 游戏配置验证 (P1-P2 优先级)
**目标**: 验证游戏配置完整性

1. **基础配置 CRUD**
   - [ ] 职业系统
   - [ ] 技能分类
   - [ ] 英雄属性类型
   - [ ] 伤害类型
   - [ ] 动作分类
   - [ ] 标签系统
   - [ ] 标签关系
   - [ ] 动作标记

2. **元数据定义**
   - [ ] 效果类型定义查询
   - [ ] 公式变量查询
   - [ ] 范围配置规则查询
   - [ ] 动作类型定义查询

### 阶段 4: 技能系统验证
**目标**: 验证技能系统完整性

1. **技能基础**
   - [ ] 技能 CRUD
   - [ ] 技能分类关联
   - [ ] 技能属性验证

2. **技能等级配置**
   - [ ] 等级配置 CRUD
   - [ ] 等级配置与技能关联
   - [ ] 多等级配置验证

3. **技能解锁动作**
   - [ ] 解锁动作关联
   - [ ] 批量设置解锁动作
   - [ ] 解锁动作移除

### 阶段 5: 效果系统验证
**目标**: 验证效果和 Buff 系统

1. **效果管理**
   - [ ] 效果 CRUD
   - [ ] 效果类型关联
   - [ ] 效果公式验证

2. **Buff 管理**
   - [ ] Buff CRUD
   - [ ] Buff 属性验证

3. **Buff-Effects 关联**
   - [ ] 单个效果添加
   - [ ] 批量效果设置
   - [ ] 效果移除
   - [ ] 关联关系验证

### 阶段 6: 动作系统验证
**目标**: 验证动作系统完整性

1. **动作管理**
   - [ ] 动作 CRUD
   - [ ] 动作类型关联
   - [ ] 动作分类关联

2. **Action-Effects 关联**
   - [ ] 单个效果添加
   - [ ] 批量效果设置
   - [ ] 效果移除
   - [ ] 关联关系验证

### 阶段 7: 边界条件和错误处理
**目标**: 验证系统健壮性

1. **输入验证**
   - [ ] 必填字段校验
   - [ ] 数据类型校验
   - [ ] 数据范围校验
   - [ ] 特殊字符处理

2. **业务规则**
   - [ ] 唯一性约束
   - [ ] 外键关联检查
   - [ ] 状态转换规则
   - [ ] 软删除验证

3. **错误响应**
   - [ ] 404 - 资源不存在
   - [ ] 400 - 请求参数错误
   - [ ] 401 - 未认证
   - [ ] 403 - 无权限
   - [ ] 409 - 资源冲突
   - [ ] 500 - 服务器错误

4. **并发场景**
   - [ ] 并发创建相同资源
   - [ ] 并发更新同一资源
   - [ ] 并发删除和更新

---

## 🛠️ 测试工具和方法

### 方法 1: 使用 Swagger UI (推荐快速测试)
```bash
# 访问 Swagger UI
open http://localhost:80/swagger/index.html

# 步骤：
# 1. 先调用 POST /api/v1/auth/login 获取 token
# 2. 点击右上角 "Authorize" 按钮
# 3. 输入: Bearer {token}
# 4. 依次测试各个接口
```

### 方法 2: 使用 curl 命令行
```bash
# 1. 登录获取 token
TOKEN=$(curl -X POST http://localhost:80/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"root","password":"password"}' | jq -r '.data.token')

# 2. 使用 token 访问受保护接口
curl -X GET http://localhost:80/api/v1/admin/users/me \
  -H "Authorization: Bearer $TOKEN"
```

### 方法 3: 使用 Postman/Insomnia
1. 导入 Swagger 文档: http://localhost:80/swagger/doc.json
2. 配置环境变量
3. 创建测试集合
4. 自动化测试流程

### 方法 4: 编写自动化测试脚本
```bash
# 使用项目中的测试框架
cd /Users/lonyon/working/军信东方/tsu项目/tsu-server-self/tsu-self
# 创建集成测试脚本
```

---

## 📝 测试用例模板

### 示例: 技能 CRUD 测试

#### TC-001: 创建技能
```yaml
测试目的: 验证技能创建功能
前置条件:
  - 用户已登录
  - 存在有效的技能分类
请求:
  方法: POST
  路径: /api/v1/admin/skills
  Headers:
    Authorization: Bearer {token}
  Body:
    name: "火球术"
    name_en: "Fireball"
    skill_category_id: 1
    description: "发射一枚火球"
    is_enabled: true
预期结果:
  - 状态码: 201
  - 返回创建的技能对象
  - 技能 ID > 0
验证点:
  - [ ] 技能已保存到数据库
  - [ ] 关联关系正确
  - [ ] 返回数据格式正确
```

#### TC-002: 查询技能列表
```yaml
测试目的: 验证技能列表查询
请求:
  方法: GET
  路径: /api/v1/admin/skills?page=1&page_size=10
预期结果:
  - 状态码: 200
  - 返回分页数据
  - 包含总数和数据列表
验证点:
  - [ ] 分页参数生效
  - [ ] 数据格式正确
  - [ ] 包含关联数据
```

#### TC-003: 更新技能
```yaml
测试目的: 验证技能更新功能
请求:
  方法: PUT
  路径: /api/v1/admin/skills/{id}
  Body:
    name: "高级火球术"
    description: "更强大的火球"
预期结果:
  - 状态码: 200
  - 返回更新后的数据
验证点:
  - [ ] 数据已更新
  - [ ] 未修改字段保持不变
  - [ ] updated_at 时间更新
```

#### TC-004: 删除技能
```yaml
测试目的: 验证技能删除功能
请求:
  方法: DELETE
  路径: /api/v1/admin/skills/{id}
预期结果:
  - 状态码: 200/204
  - 软删除: deleted_at 有值
验证点:
  - [ ] 技能不再出现在列表中
  - [ ] 关联数据处理正确
  - [ ] 可以恢复（如果支持）
```

---

## 🔍 关键测试场景

### 场景 1: 完整技能配置流程
```
1. 创建技能分类
2. 创建技能
3. 为技能创建多个等级配置
4. 创建效果
5. 创建动作
6. 关联效果到动作
7. 关联动作到技能（解锁动作）
8. 查询完整技能树
9. 更新配置
10. 删除测试数据
```

### 场景 2: RBAC 权限验证流程
```
1. 创建自定义角色
2. 为角色分配权限
3. 创建测试用户
4. 为用户分配角色
5. 使用该用户登录
6. 验证权限生效
7. 测试无权限操作被拒绝
8. 清理测试数据
```

### 场景 3: 关联数据完整性
```
1. 创建主实体（如 Buff）
2. 创建关联实体（如 Effects）
3. 建立关联关系
4. 验证关联查询
5. 删除关联（不删除主体）
6. 验证关联已解除
7. 删除主实体
8. 验证关联数据处理
```

### 场景 4: 批量操作测试
```
1. 创建 Buff
2. 创建多个 Effects
3. 使用批量接口设置 Buff-Effects
4. 验证所有关联已建立
5. 使用批量接口覆盖设置
6. 验证旧关联已清除，新关联已建立
```

---

## 📊 测试数据准备

### 基础测试数据
```sql
-- 职业数据
INSERT INTO game_config.classes (name, name_en, description) VALUES
  ('战士', 'Warrior', '近战物理职业'),
  ('法师', 'Mage', '远程魔法职业'),
  ('牧师', 'Priest', '治疗辅助职业');

-- 技能分类
INSERT INTO game_config.skill_categories (name, name_en, description) VALUES
  ('攻击', 'Attack', '攻击类技能'),
  ('防御', 'Defense', '防御类技能'),
  ('辅助', 'Support', '辅助类技能');

-- 伤害类型
INSERT INTO game_config.damage_types (name, name_en, color) VALUES
  ('物理', 'Physical', '#FF0000'),
  ('魔法', 'Magic', '#0000FF'),
  ('治疗', 'Healing', '#00FF00');
```

### 测试用户数据
```json
{
  "test_admin": {
    "username": "test_admin",
    "password": "Test123!",
    "email": "admin@test.com",
    "role": "admin"
  },
  "test_user": {
    "username": "test_user",
    "password": "Test123!",
    "email": "user@test.com",
    "role": "user"
  }
}
```

---

## 🚦 测试执行检查清单

### 测试前准备
- [ ] 确认服务已启动 (`docker-compose ps`)
- [ ] 确认数据库迁移完成
- [ ] 确认 Nginx/Oathkeeper 运行正常
- [ ] 准备测试账号 (root/password)
- [ ] 备份数据库（如果在生产环境）

### 执行过程
- [ ] 按优先级执行测试用例
- [ ] 记录每个接口的测试结果
- [ ] 对失败的用例记录详细日志
- [ ] 截图保存关键测试场景
- [ ] 验证 Swagger 文档与实际接口一致

### 测试后
- [ ] 清理测试数据
- [ ] 生成测试报告
- [ ] 提交 Bug 报告（如有）
- [ ] 更新接口文档（如有变更）
- [ ] 归档测试结果

---

## 📈 测试报告模板

### 测试摘要
```
测试日期: YYYY-MM-DD
测试人员: XXX
测试环境: Local/Dev/Test
服务版本: v1.0.0

总接口数: 100+
已测试: XX
通过: XX
失败: XX
阻塞: XX
通过率: XX%
```

### 问题清单
| ID | 接口 | 问题描述 | 严重级别 | 状态 |
|----|------|---------|---------|------|
| BUG-001 | POST /api/v1/admin/skills | 创建技能时分类ID验证缺失 | 中 | Open |
| BUG-002 | DELETE /api/v1/admin/roles/:id | 删除关联用户的角色未返回错误 | 高 | Open |

### 改进建议
1. 统一错误响应格式
2. 添加更详细的验证消息
3. 优化批量操作性能
4. 完善 Swagger 文档注释

---

## 🔧 常见问题处理

### 1. Token 过期
```bash
# 重新登录获取新 token
curl -X POST http://localhost:80/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"root","password":"password"}'
```

### 2. 403 Forbidden
- 检查用户是否有相应权限
- 确认角色配置正确
- 验证 Oathkeeper 规则

### 3. 404 Not Found
- 检查 URL 路径是否正确
- 确认 Nginx 代理配置
- 验证服务是否正常运行

### 4. 500 Internal Server Error
- 查看服务日志: `docker logs tsu_admin`
- 检查数据库连接
- 验证数据完整性约束

---

## 📚 参考资料

- [Swagger UI](http://localhost:80/swagger/index.html)
- [项目认证指南](../docs/AUTHENTICATION_GUIDE.md)
- [权限测试文档](../docs/PERMISSION_TESTING.md)
- [技能系统规范](../configs/技能配置规范.md)
- [API 架构规则](../CLAUDE.md)

---

## 附录：自动化测试脚本示例

见下一个文件：`admin-api-test.sh`
