// Package notification mirrors the real-world case depx targets: a factory
// with many collaborators. Instead of a long positional constructor (where
// adding a dependency forces every call site, tests included, to pass another
// argument), depx generates a Deps struct + validated constructor + test
// builder. Run `make gen` to regenerate factory_deps.go.
package notification

import "time"

//go:generate depx -type factory -return Factory

// Factory is what the constructor returns.
type Factory interface {
	Build() string
}

// These stand in for the real collaborator types (db, auth funcs, services).
type (
	DB                 interface{ Query() string }
	CurrentProfileFunc func() string
	AccountUserService interface{ User() string }
	FlagChecker        interface{ Enabled() bool }
	Registry           interface{ Lookup(string) string }
	Enricher           interface{ Enrich(string) string }
)

// factory is the implementation struct. Its fields are the injected
// dependencies; depx mirrors them into an exported FactoryDeps.
type factory struct {
	db                 DB
	profileAuth        CurrentProfileFunc
	accountUserService AccountUserService
	flagChecker        FlagChecker
	templates          Registry `depx:"name=Templates"` // field/Deps names differ
	enricher           Enricher `depx:"optional"`       // may be nil

	createdAt time.Time `depx:"-"` // internal state, not injected
}

func (f *factory) Build() string {
	out := f.db.Query() + ":" + f.profileAuth()
	if f.enricher != nil {
		out = f.enricher.Enrich(out)
	}
	return out
}
