package notification

import (
	"errors"
	"testing"

	"github.com/zkrebbekx/depx"

	. "github.com/smartystreets/goconvey/convey"
)

type stubDB struct{}

func (stubDB) Query() string { return "q" }

type stubAUS struct{}

func (stubAUS) User() string { return "u" }

type stubFlags struct{}

func (stubFlags) Enabled() bool { return true }

type stubRegistry struct{}

func (stubRegistry) Lookup(string) string { return "" }

// allDeps fills every required dependency. A test that doesn't care about a
// dependency simply omits it.
func allDeps() FactoryDeps {
	return NewTestFactoryDeps(func(d *FactoryDeps) {
		d.Db = stubDB{}
		d.ProfileAuth = func() string { return "p" }
		d.AccountUserService = stubAUS{}
		d.FlagChecker = stubFlags{}
		d.Templates = stubRegistry{}
		// Enricher omitted — it is optional.
	})
}

func TestNewFactoryWired(t *testing.T) {
	Convey("Given a fully wired FactoryDeps", t, func() {
		f, err := NewFactory(allDeps())

		Convey("Then construction succeeds and the factory works", func() {
			So(err, ShouldBeNil)
			So(f.Build(), ShouldEqual, "q:p")
		})
	})
}

func TestNewFactoryMissingRequired(t *testing.T) {
	Convey("Given deps with a required dependency left nil", t, func() {
		d := allDeps()
		d.ProfileAuth = nil

		Convey("When constructing", func() {
			f, err := NewFactory(d)

			Convey("Then it fails fast, naming the missing dependency", func() {
				So(f, ShouldBeNil)
				So(err, ShouldNotBeNil)
				var me *depx.MissingError
				So(errors.As(err, &me), ShouldBeTrue)
				So(me.Fields, ShouldContain, "ProfileAuth")
			})
		})
	})
}

func TestOptionalDependencyMayBeNil(t *testing.T) {
	Convey("Given deps with the optional Enricher nil", t, func() {
		f, err := NewFactory(allDeps()) // Enricher never set

		Convey("Then construction still succeeds", func() {
			So(err, ShouldBeNil)
			So(f, ShouldNotBeNil)
		})
	})
}

func TestTestBuilderSetsOnlyWhatMatters(t *testing.T) {
	Convey("Given a test that overrides one dependency over a full base", t, func() {
		d := NewTestFactoryDeps(
			func(d *FactoryDeps) { *d = allDeps() },
			func(d *FactoryDeps) { d.ProfileAuth = func() string { return "override" } },
		)

		Convey("When building", func() {
			f, err := NewFactory(d)

			Convey("Then the override takes effect", func() {
				So(err, ShouldBeNil)
				So(f.Build(), ShouldEqual, "q:override")
			})
		})
	})
}
