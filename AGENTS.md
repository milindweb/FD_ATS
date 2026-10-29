<!-- gitnexus:start -->
# GitNexus — Code Intelligence

This project is indexed by GitNexus as **FD_ATS** (1542 symbols, 3829 relationships, 109 execution flows).

> Index stale? Run `node .gitnexus/run.cjs analyze --index-only` from the project root — it auto-selects an available runner. No `.gitnexus/run.cjs` yet? Bootstrap with `npx`, `bunx`, or `pnpm dlx` — e.g. `bunx gitnexus@latest analyze` (npm 11 npx crash; #1939).

## Always Do

- **MUST run impact before editing.** Use `impact({target: "symbolName", direction: "upstream"})` or `node .gitnexus/run.cjs impact "symbolName" --direction upstream --repo .`; report callers, processes, and risk. Never substitute grep for graph analysis.
- **MUST analyze graph changes before committing.** Use `detect_changes({scope: "all"})` (MCP) or `node .gitnexus/run.cjs detect-changes --scope all --repo .` (CLI fallback). `partial: true` or `truncated: true` is not a clean check — a zero means unseen, not unaffected; re-run it. For regression review: `detect_changes({scope: "compare", base_ref: "main"})` or `node .gitnexus/run.cjs detect-changes --scope compare --base-ref "main" --repo .`.
- MUST warn on HIGH/CRITICAL `risk` pre-edit; never use `riskSharedAxes` to waive a HIGH/CRITICAL `risk` warning. Compare File/symbol: MCP File omits axes; Graph-RAG expands File.
- **MUST treat `risk: UNKNOWN` as unresolved, not as low.** An empty caller set is not evidence the symbol is unused — it can also mean the callers are not resolvable by the index (plain-object property access, dynamic dispatch, cross-language calls). `impact` pairs `UNKNOWN` with a `riskNote` saying so. Confirm with a text search before treating the symbol as safe to change or delete; do not proceed on the strength of a zero.
- **MUST use `query({search_query: "concept"})` for concepts/flows, `context({name: "symbolName"})` for a named symbol, or `impact` for blast radius, on read-only callers, dependencies, imports, or execution flow.** Graph first; text search only for empty/`UNKNOWN`/literals.
- For security review, `explain({target: "fileOrSymbol"})` lists taint findings (source→sink flows; needs `analyze --pdg`).

## Never Do

- NEVER edit a function, class, or method before MCP/CLI impact analysis.
- NEVER ignore HIGH or CRITICAL risk warnings from impact analysis, and never read `UNKNOWN` as an all-clear — it means the walk could not answer, which is the one verdict that requires confirming by other means.
- NEVER rename symbols with find-and-replace — use `rename` which understands the call graph.
- NEVER commit before MCP/CLI graph change analysis.

## Resources

| Resource | Use for |
| --- | --- |
| `gitnexus://repo/FD_ATS/context` | Codebase overview, check index freshness |
| `gitnexus://repo/FD_ATS/clusters` | All functional areas |
| `gitnexus://repo/FD_ATS/processes` | All execution flows |
| `gitnexus://repo/FD_ATS/process/{name}` | Step-by-step execution trace |

## CLI

| Task | Read this skill file |
| --- | --- |
| Understand architecture / "How does X work?" | `.claude/skills/gitnexus-exploring/SKILL.md` |
| Blast radius / "What breaks if I change X?" | `.claude/skills/gitnexus-impact-analysis/SKILL.md` |
| Trace bugs / "Why is X failing?" | `.claude/skills/gitnexus-debugging/SKILL.md` |
| Rename / extract / split / refactor | `.claude/skills/gitnexus-refactoring/SKILL.md` |
| Tools, resources, schema reference | `.claude/skills/gitnexus-guide/SKILL.md` |
| Index, status, clean, wiki CLI commands | `.claude/skills/gitnexus-cli/SKILL.md` |

<!-- gitnexus:end -->

## Superpowers + GitNexus (this repo)

GitNexus supplies code intelligence, Superpowers supplies the process. Use
them together, not as alternatives. Everything below is **on request of this
repo's owner** and applies to every agent working in it.

**Before changing behaviour**

1. `brainstorming` — open-ended / new feature design only. Skip for
   mechanical edits, small fixes, config and docs changes.
2. `writing-plans` or `gitnexus-plan` — multi-step work gets a plan file
   (`docs/plans/`); trivial changes do not.
3. `impact` / `context` / `query` — always on the symbol you are about to
   edit, before editing (see rules above). `risk: UNKNOWN` is not an
   all-clear.
4. `test-driven-development` for new behaviour, `systematic-debugging` for
   bugs (graph first, then reproduce). Batch fixes, run the suite once at the
   end.

**Before claiming done or committing**

- `verification-before-completion` — evidence from real command output, plus
  `detect_changes({scope: "all"})` for a clean graph check.
- `requesting-code-review` / `gitnexus-review` before merge.
- Never commit, push, branch, worktree or rewrite history unless explicitly
  asked.

**Pipelines**

- `/gitnexus-lfg` = plan → gate → work → review.
- CLI fallback for any GitNexus MCP tool: `node .gitnexus/run.cjs <command> --repo .`
