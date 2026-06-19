package depx_test

import (
	"errors"
	"testing"

	"github.com/zkrebbekx/depx"

	. "github.com/smartystreets/goconvey/convey"
)

type Logger interface{ Log(string) }

type Deps struct {
	DB       *int           // pointer dep
	Auth     func() string  // func dep
	Log      Logger         // interface dep
	Cache    map[string]int // map dep
	Enricher Logger         `depx:"optional"` // optional: nil allowed
	Internal Logger         `depx:"-"`        // excluded
	Retries  int            // non-nilable: ignored
	hidden   Logger         //nolint:unused // unexported: skipped
}

func full() Deps {
	x := 1
	return Deps{
		DB:    &x,
		Auth:  func() string { return "" },
		Log:   nil, // set below in tests as needed
		Cache: map[string]int{},
	}
}

func TestValidateAllRequiredSet(t *testing.T) {
	Convey("Given a Deps with every required field set", t, func() {
		d := full()
		d.Log = nopLogger{}

		Convey("When validated", func() {
			err := depx.Validate(d)

			Convey("Then it passes", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestValidateNamesMissing(t *testing.T) {
	Convey("Given a Deps missing two required fields", t, func() {
		d := full()
		d.DB = nil  // missing
		d.Log = nil // missing

		Convey("When validated", func() {
			err := depx.Validate(d)

			Convey("Then it returns a MissingError naming them in order", func() {
				So(err, ShouldNotBeNil)
				var me *depx.MissingError
				So(errors.As(err, &me), ShouldBeTrue)
				So(me.Fields, ShouldResemble, []string{"DB", "Log"})
				So(me.Type, ShouldEqual, "Deps")
				So(err.Error(), ShouldContainSubstring, "missing required dependencies: DB, Log")
			})
		})
	})
}

func TestOptionalAndExcludedSkipped(t *testing.T) {
	Convey("Given a Deps whose optional and excluded fields are nil", t, func() {
		d := full()
		d.Log = nopLogger{}
		d.Enricher = nil // optional
		d.Internal = nil // excluded

		Convey("When validated", func() {
			err := depx.Validate(d)

			Convey("Then those fields do not count as missing", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestNonNilableIgnored(t *testing.T) {
	Convey("Given a required non-nilable field left at zero", t, func() {
		d := full()
		d.Log = nopLogger{}
		d.Retries = 0 // zero int, but not nil-able

		Convey("When validated", func() {
			err := depx.Validate(d)

			Convey("Then it is not reported (depx only checks nil-able deps)", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestValidateAcceptsPointer(t *testing.T) {
	Convey("Given a pointer to a complete Deps", t, func() {
		d := full()
		d.Log = nopLogger{}

		Convey("When validated by pointer", func() {
			err := depx.Validate(&d)

			Convey("Then it dereferences and passes", func() {
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("Given a typed nil pointer", t, func() {
		Convey("When validated", func() {
			err := depx.Validate((*Deps)(nil))

			Convey("Then it errors clearly rather than panicking", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "nil")
			})
		})
	})
}

func TestValidateRejectsNonStruct(t *testing.T) {
	Convey("Given a non-struct value", t, func() {
		Convey("When validated", func() {
			err := depx.Validate(42)

			Convey("Then it reports the misuse", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "requires a struct")
			})
		})
	})
}

func TestMustValidatePanics(t *testing.T) {
	Convey("Given an incomplete Deps", t, func() {
		d := full()
		d.DB = nil

		Convey("When MustValidate is called", func() {
			Convey("Then it panics", func() {
				So(func() { depx.MustValidate(d) }, ShouldPanic)
			})
		})

		Convey("When MustValidate is called on a complete Deps", func() {
			d.DB = func() *int { x := 1; return &x }()
			d.Log = nopLogger{}

			Convey("Then it does not panic", func() {
				So(func() { depx.MustValidate(d) }, ShouldNotPanic)
			})
		})
	})
}

type nopLogger struct{}

func (nopLogger) Log(string) {}
