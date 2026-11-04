---
name: taste
description: >-
  House rules for refactoring, distilled from past review feedback on refactoring requests.
  Use whenever asked to refactor, simplify, clean up, clean code, rename, extract, inline,
  move, decouple, collapse, or remove indirection in existing code, or to propose refactoring
  opportunities on a branch, diff, or file. Covers scope, indirection, placement, naming,
  comments, signatures, and verification.
---

# Taste

## Workflow

1. **Pin the scope.** Restate in one line what changes and what stays. A refactor changes
   structure, not behavior, flags, public interfaces, or unrelated code. Only the named
   function, file, or diff is in scope unless the request says otherwise.
2. **Read the neighborhood first.** Check how the codebase already does it (sibling files,
   other models, recent scripts, the module being migrated from) and match it. Before deleting
   anything that looks dead, check callers and history for why it was added.
3. **Propose when the request is open** ("propose refactoring", "look for opportunities",
   "wdyt?"). Numbered list, highest impact first, one line each with `file:line`. For renames
   give 2-3 alternatives. Say "current approach is fine" when it is. Wait for the pick.
4. **Execute the smallest change** that satisfies the request.
5. **Sweep.** After a rename, move, or signature change, update every caller, test, doc, and
   prompt that references it.
6. **Verify.** Typecheck, lint, tests. Fix real failures. Never raise a threshold, loosen a
   type, or add a type extension to make a check pass.

## Scope

- Don't touch what wasn't asked: other call sites, shared submodules, behavior.
- Minimize the diff. No reformatting, reordering, retyping, or improving lines outside the ask.
- Every changed line must answer "why this?". If the answer is "seemed cleaner", revert it.
- Never silently drop or delete while simplifying. Keep it or ask.
- Preserve edge cases the old code handled.

## Indirection

- Inline single-use wrappers. Call the underlying function directly.
- Inline trivial wrappers even when reused: a one-line getter, field fallback, or config read.
  Keep a function only when it holds real logic reused in several places. When the inlined
  expression is not self-explanatory, put it in a named intermediate variable.
- No re-exports, aliases, duplicate constants, forwarding wrappers, or type extensions to
  bridge a change. Old and new never live together.
- Collapse split conditionals into one. Derive booleans as variables in place.
- Don't design for cases not yet needed. No optional param, flag, or abstraction for one use.
- Prefer framework features and well-known libraries over hand-rolled code. Prefer the
  codebase's existing convention over a new one.

## Placement

- Extract when a block is reused or a function is long. Prefer an extracted function over a
  new parameter. Use intermediate variables over nested expressions.
- Fewer files. Fold into an existing file before creating one. Merge modules that are the
  same concept. Put shared test helpers in one place.
- Code lives in the module that owns the concept.
- Shared gating or validation goes at the shared call site, not repeated in each caller.
- Side effects and control flow stay visible in the caller. Pure or derived logic moves down
  into the callee.
- One responsibility per function: authentication apart from authorization, status update
  apart from level update, persisting state apart from emitting events.
- No leaky abstractions. An update function persists what it says it updates.
- Models store minimal state. Derived mappings live in controllers, not in the schema.

## Naming

- The name states exactly what the function does, at the precision of its domain. A function
  that only writes state is `update…`, not `sync…`. A generic assert gets the noun it checks.
- Don't repeat the module name inside the module.
- No vague or hedging verbs (`ensure`, `maybe`, `sync`, `handle` for a read) and no uncommon
  prefixes (`new…`, `$set`).
- Generic parameter names when the type already carries the meaning (`error`, `userId`).
- Units in the name when the value has a unit.
- Remove dead names, statuses, fields, and config options found along the way.

## Comments

- Keep a comment only if it adds a non-obvious why. The how is the code.
- Delete comments that narrate the code, restate the function name, say where else something
  is enforced, or describe how other code depends on this one.
- No comments referencing temporary flags or rollout state.
- Don't duplicate a comment across layers. File-header comments stay short.
- A business or economic reason for a non-obvious constant is welcome, in `/** */` form, and
  must be precise.

## Signatures

- Few parameters. Never pass the same data twice or duplicate a field between params.
- Build internal details (references, counterparties, ids) inside the function. Don't ask
  callers for them.
- When migrating, keep the previous interface so call sites don't change.
- Don't loosen types to compile. Tighten when the loose type has no consumer.
- Destructure. Full `if` blocks with braces.
