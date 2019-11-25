package errutil

import (
	"fmt"
	"strings"

	"golang.org/x/xerrors"
)

// Merge is a convenience method for making a Slice of errors and calling the Merge method.
func Merge(errs ...error) error {
	s := Slice(errs)
	return s.Merge()
}

// Slice is a slice of errors that implements the error interface itself.
type Slice []error

// Push extends a Slice with an error if the error is non-nil.
//
// If a Slice is passed to Push, the result is flattened.
func (s *Slice) Push(err error) {
	if s2 := new(Multierr); xerrors.As(err, s2) {
		*s = append(*s, s2.s...)
	} else if err != nil {
		*s = append(*s, err)
	}
}

// Merge first removes any nil errors from the Slice.
// If the resulting length of the Slice is zero, it returns nil.
// If there is only one error, it returns that error as is.
// If there are multiple errors, it returns a Multierr
// containing all the errors.
func (s *Slice) Merge() error {
	// Making a copy in case we need to flatten a nested Slice
	errsFiltered := make(Slice, 0, len(*s))
	for _, err := range *s {
		errsFiltered.Push(err)
	}
	*s = errsFiltered
	if len(errsFiltered) < 1 {
		return nil
	}
	if len(errsFiltered) == 1 {
		return (errsFiltered)[0]
	}
	return Multierr{errsFiltered}
}

// Multierr wraps multiple errors.
type Multierr struct {
	s Slice
}

var _ error = Multierr{}

// Slice returns the underlying slice of errors.
func (m Multierr) Slice() Slice {
	return m.s
}

// Strings returns the strings from the underlying errors.
func (m Multierr) Strings() []string {
	return errorsToStrings(m.s)
}

// Error implements the error interface.
func (m Multierr) Error() string {
	a := m.Strings()
	if len(a) == 0 {
		return "<empty error slice>"
	}
	plural := "s"
	if len(a) == 1 {
		plural = ""
	}
	return fmt.Sprintf("%d error%s: %s", len(a), plural, strings.Join(a, "; "))
}

func errorsToStrings(s []error) []string {
	a := make([]string, 0, len(s))
	for _, err := range s {
		if s != nil {
			a = append(a, err.Error())
		}
	}
	return a
}

var _ fmt.Formatter = Multierr{}

// Format implements fmt.Formatter.
func (m Multierr) Format(state fmt.State, verb rune) {
	switch verb {
	case 's', 'q', 'v':
		if verb == 'v' && state.Flag('+') {
			a := m.Strings()
			if len(a) == 0 {
				fmt.Fprint(state, "<empty error slice>")
				return
			}

			plural := "s"
			if len(a) == 1 {
				plural = ""
			}
			fmt.Fprintf(state, "%d error%s:\n", len(a), plural)
			for i, err := range a {
				fmt.Fprintf(state, "\terror %d: %s\n", i+1, err)
			}
		} else {
			msg := m.Error()
			if verb == 'q' {
				msg = fmt.Sprintf("%q", msg)
			}
			fmt.Fprint(state, msg)
		}
	}
}
