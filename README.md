# depx

Constructors with many dependencies, without the pain. `depx` turns a long
positional constructor into a **deps struct** — so adding a dependency no longer
forces every call site (tests included) to pass another argument — and gives you
back the safety you lose: a **fail-fast check** that every required dependency is
wired, naming any that aren't.

Library is zero-dependency. Go 1.24+.

```go
import "github.com/zkrebbekx/depx"
```

## The problem

```go
func NewFactory(
	db *sqlx.DB,
	profileAuth auth.CurrentProfileFunc,
	// ...7 more...
) Factory { ... }
```

Add a tenth dependency and **every** call site breaks — including the dozens of
tests that pass `nil` for things they don't care about. Switch to a struct and
that churn disappears (Go zero-values omitted fields), but you lose the compile
error that told you a real wiring site forgot the new dependency.

`depx` gives you the struct *and* the safety.

## Runtime: validate required dependencies

```go
type FactoryDeps struct {
	DB          *sqlx.DB
	ProfileAuth auth.CurrentProfileFunc
	Enricher    Enricher `depx:"optional"` // nil allowed
}

func NewFactory(d FactoryDeps) (Factory, error) {
	if err := depx.Validate(d); err != nil {
		return nil, err // "depx: FactoryDeps is missing required dependencies: DB, ProfileAuth"
	}
	return &factory{db: d.DB, profileAuth: d.ProfileAuth, enricher: d.Enricher}, nil
}
```

- Checks only **nil-able** fields (pointers, interfaces, funcs, maps, channels,
  slices) — what a missing dependency actually looks like.
- `depx:"optional"` and `depx:"-"` fields, and unexported fields, are skipped.
- Adding a field is automatically covered — there's no `||` chain to forget.
- `depx.MustValidate(d)` panics instead, for process start-up.

The error is a typed `*depx.MissingError` (`.Type`, `.Fields`) so callers can
inspect it.

## Generator: stop writing the boilerplate

Annotate the implementation struct and let `depx` emit the rest:

```go
//go:generate depx -type factory -return Factory

type factory struct {
	db          DB
	profileAuth CurrentProfileFunc
	templates   Registry `depx:"name=Templates"` // override the exported name
	enricher    Enricher `depx:"optional"`
	createdAt   time.Time `depx:"-"`             // internal state, not injected
}
```

```sh
go install github.com/zkrebbekx/depx/cmd/depx@latest
go generate ./...   # writes factory_deps.go
```

It generates:

- `FactoryDeps` — the exported deps struct, mirroring the injected fields
  (carrying `depx:"optional"` through, and any imports the field types need).
- `NewFactory(FactoryDeps) (Factory, error)` — validates with `depx.Validate`,
  then wires each field. Return type comes from `-return` (default `*factory`).
- `NewTestFactoryDeps(overrides ...func(*FactoryDeps)) FactoryDeps` — a test
  builder so a test sets only the dependencies it cares about.

Field tags:

| tag | effect |
|---|---|
| _none_ | required dependency |
| `depx:"optional"` | may be nil |
| `depx:"-"` | not a dependency (internal state); excluded |
| `depx:"name=DB"` | override the exported `Deps` field name |

## What you get when a dependency is added

Add a new required field and regenerate:

- **Compile:** existing tests are untouched — no positional `nil` churn.
- **Runtime:** any test that builds a real value gets
  `missing required dependencies: Mailer`, fixed in **one** place (your test-deps
  helper) instead of scattered `nil`s — and never a nil-pointer panic far from
  the cause.

## Develop

```sh
make test   # library tests
make lint
cd cmd/depx && make test
cd examples && make gen && make test
```

## License

MIT
