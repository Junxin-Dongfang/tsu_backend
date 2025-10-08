// Keto Namespace Configuration
// Ory Keto v0.14+ Namespace Definition
// Reference: https://www.ory.sh/docs/keto/modeling/create-permission-model

import { Namespace, SubjectSet, Context } from '@ory/keto-namespace-types'

// User (Subject type)
class User implements Namespace {}

// Role namespace
// Defines role membership
class Role implements Namespace {
  related: {
    // Users who are members of this role
    member: User[]
  }
}

// Permission namespace
// Defines permission grants with role inheritance
class Permission implements Namespace {
  related: {
    // Users who directly hold this permission (bypass roles)
    holder: User[]
    // Roles that are granted this permission (as SubjectSet to enable traversal)
    granted: SubjectSet<Role, 'member'>[]
  }

  // Permission check: user has permission if:
  // 1. User is a direct holder, OR
  // 2. User is a member of a role that is granted this permission
  permits = {
    view: (ctx: Context): boolean =>
      this.related.holder.includes(ctx.subject) ||
      this.related.granted.includes(ctx.subject)
  }
}
