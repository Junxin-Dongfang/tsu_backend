// Keto Relation Definition - TypeScript 风格语法
// 参考: https://www.ory.sh/docs/keto/modeling/create-permission-model

// 定义命名空间
import { Namespace, Context, SubjectSet } from "@ory/keto-namespace-types"

// 角色命名空间
class Role implements Namespace {
  // 角色的成员关系
  related: {
    member: User[]
  }
}

// 权限命名空间
class Permission implements Namespace {
  // 权限的授予关系
  related: {
    // 直接授予用户的权限
    holder: User[]
    // 通过角色授予的权限
    granted: SubjectSet<Role, "member">[]
  }

  // 权限检查: 用户是否拥有此权限
  // 1. 用户是直接 holder
  // 2. 或者用户是某个角色的 member,且该角色被 granted 此权限
  permits = {
    view: (ctx: Context): boolean =>
      this.related.holder.includes(ctx.subject) ||
      this.related.granted.includes(ctx.subject)
  }
}

// 用户类(Subject)
class User implements Namespace {}
