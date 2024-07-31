import { Namespace, Context } from "@ory/keto-namespace-types"

class Role implements Namespace {}

class Service implements Namespace {
  related: {
    callers: Role[]
  }

  permits = {
    call: (ctx: Context): boolean => this.related.callers.includes(ctx.subject),
  }
}
