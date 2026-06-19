// Package depx validates that a dependency struct has every required
// dependency wired, so a constructor with many collaborators can fail fast
// with a clear message instead of panicking on a nil far from the cause.
//
// It pairs with the "deps struct" constructor pattern: instead of a long
// positional parameter list (where adding a dependency forces every call
// site — including tests — to pass another argument), collaborators are
// fields of an exported struct. [Validate] then checks that every required
// field is non-nil, naming any that are missing.
//
//	type FactoryDeps struct {
//		DB          *sqlx.DB
//		ProfileAuth auth.CurrentProfileFunc
//		Enricher    Enricher `depx:"optional"` // nil is allowed
//	}
//
//	func NewFactory(d FactoryDeps) (Factory, error) {
//		if err := depx.Validate(d); err != nil {
//			return nil, err // "depx: missing required dependencies: DB, ProfileAuth"
//		}
//		return &factory{db: d.DB, profileAuth: d.ProfileAuth, ...}, nil
//	}
//
// Only nil-able fields (pointers, interfaces, funcs, maps, channels, slices)
// are checked — those are what a missing dependency looks like. Other kinds
// are ignored. Fields tagged `depx:"optional"` or `depx:"-"`, and unexported
// fields, are skipped.
//
// A go:generate tool that emits the Deps struct, the wired constructor, and
// a test-deps builder from an annotated implementation struct is available as
// github.com/zkrebbekx/depx/cmd/depx.
package depx

import (
	"fmt"
	"reflect"
	"strings"
)

// MissingError reports the required dependency fields that were nil. It is
// returned by [Validate].
type MissingError struct {
	// Type is the name of the validated struct type.
	Type string
	// Fields are the names of the missing required fields, in declaration
	// order.
	Fields []string
}

func (e *MissingError) Error() string {
	return fmt.Sprintf("depx: %s is missing required dependencies: %s",
		e.Type, strings.Join(e.Fields, ", "))
}

// Validate reports whether every required dependency in d is set. d must be a
// struct or a non-nil pointer to one. A required field is an exported,
// nil-able field not tagged `depx:"optional"` or `depx:"-"`; if any such
// field is nil, Validate returns a [*MissingError] naming them all.
func Validate[T any](d T) error {
	v := reflect.ValueOf(d)
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return fmt.Errorf("depx: Validate called on a nil %T", d)
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("depx: Validate requires a struct, got %s", v.Kind())
	}

	t := v.Type()
	var missing []string
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		if skipField(f.Tag.Get("depx")) {
			continue
		}
		fv := v.Field(i)
		if isNilable(fv.Kind()) && fv.IsNil() {
			missing = append(missing, f.Name)
		}
	}
	if len(missing) > 0 {
		return &MissingError{Type: t.Name(), Fields: missing}
	}
	return nil
}

// MustValidate is like [Validate] but panics on a missing dependency. Use it
// at process start, where a misconfiguration should stop the program.
func MustValidate[T any](d T) {
	if err := Validate(d); err != nil {
		panic(err)
	}
}

// skipField reports whether a `depx` struct-tag value marks a field as not
// required (optional, or excluded).
func skipField(tag string) bool {
	for _, opt := range strings.Split(tag, ",") {
		switch strings.TrimSpace(opt) {
		case "optional", "-":
			return true
		}
	}
	return false
}

func isNilable(k reflect.Kind) bool {
	switch k {
	case reflect.Pointer, reflect.Interface, reflect.Func,
		reflect.Map, reflect.Chan, reflect.Slice:
		return true
	default:
		return false
	}
}
