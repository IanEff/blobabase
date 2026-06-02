# blobabase

Long-running service. Module `github.com/ifurst/blobabase`. Structured `slog` logging, context-driven graceful shutdown.

## Definition of done
- `make ci` is green: fmt → vet → test (`-race`) → build. Don't call work done on a red bar.

## Go house rules
- Errors: wrap with `%w`, compare with `errors.Is` / `errors.As`, combine with `errors.Join`. Sentinels as package-level `var ErrFoo = errors.New(...)`.
- Never return a typed-nil pointer as an `error` — return literal `nil`.
- Interfaces are small and defined by the consumer, not shipped beside the implementation. Accept interfaces, return structs.
- `context.Context` is the first parameter, never a struct field. Thread it through; no `context.Background()` deep in a call chain.
- Anything touching goroutines or channels must pass `go test -race`. Use `testing/synctest` (`synctest.Test`) for deterministic time/concurrency tests.
- No performance claim without a `testing.B` and a `benchstat` before/after. `go build -gcflags=-m` for escape-to-heap questions.
- stdlib-first: `any` over `interface{}`, builtin `min`/`max`/`clear`, `log/slog`, `slices`/`maps` over hand-rolled loops.
- Don't guess a signature — read it via `go doc` or the gopls `go_package_api` tool.

## gopls (MCP)
- Run `go_diagnostics` after each edit, not once at the end — catch errors incrementally.
- `go_package_api` to learn an unfamiliar package's surface; `go_symbol_references` / `go_rename_symbol` for refactors instead of textual find-replace.

## Service shape
- Operational output goes through the default `slog` JSON handler — no `fmt.Println`.
- Shutdown is driven by `signal.NotifyContext`; new long-running work selects on `ctx.Done()` and exits cleanly.
