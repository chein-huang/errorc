package errorc

import (
	"fmt"
	"io"
	"runtime"

	"github.com/pkg/errors"
)

type F = map[string]interface{}

type ErrorWithFields struct {
	error
	fields F
	stack  *stack
}

func (e *ErrorWithFields) Is(err error) bool {
	return errors.Is(e.error, err)
}

func (e *ErrorWithFields) Unwrap() error {
	return e.error
}

func (e *ErrorWithFields) LogFields() map[string]interface{} {
	return e.fields
}

func Cause(err error) error {
	if err == nil {
		return nil
	}

	return MustBase(err).Unwrap()
}

func MustBase(err error) *ErrorWithFields {
	return mustBase(err, 1)
}

func Newf(format string, a ...interface{}) error {
	return &ErrorWithFields{
		error:  fmt.Errorf(format, a...),
		stack:  callers(1),
		fields: F{},
	}
}

func Wrap(err error) error {
	fe := base(err)
	if fe == nil {
		f := &ErrorWithFields{
			error:  err,
			stack:  callers(1),
			fields: F{},
		}
		return f
	}
	return err
}

func base(err error) *ErrorWithFields {
	if err == nil {
		return nil
	}

	unwrapErr := err
	for unwrapErr != nil {
		if e, ok := unwrapErr.(*ErrorWithFields); ok {
			return e
		}
		unwrapErr = errors.Unwrap(unwrapErr)
	}
	return nil
}

func mustBase(err error, skip int) *ErrorWithFields {
	fe := base(err)
	if fe == nil {
		f := &ErrorWithFields{
			error:  err,
			stack:  callers(1 + skip),
			fields: F{},
		}
		return f
	}
	return fe
}

func AddField(err error, key string, value interface{}) *ErrorWithFields {
	if err == nil {
		return nil
	}

	f := mustBase(err, 1)
	return f.AddField(key, value)
}

func (e *ErrorWithFields) AddField(key string, value interface{}) *ErrorWithFields {
	if e.fields == nil {
		e.fields = F{key: value}
	} else {
		e.fields[key] = value
	}

	return e
}

func AddFieldf(err error, key, valueFormat string, args ...interface{}) *ErrorWithFields {
	if err == nil {
		return nil
	}

	return AddField(err, key, fmt.Sprintf(valueFormat, args...))
}

func (e *ErrorWithFields) AddFieldf(key, valueFormat string, args ...interface{}) *ErrorWithFields {
	return e.AddField(key, fmt.Sprintf(valueFormat, args...))
}

func AddFields(err error, fields F) *ErrorWithFields {
	if err == nil {
		return nil
	}

	f := mustBase(err, 1)
	for key, value := range fields {
		f.AddField(key, value)
	}
	return f
}

type stack []uintptr

func (s *stack) Format(st fmt.State, verb rune) {
	switch verb {
	case 'v':
		switch {
		case st.Flag('+'):
			for _, pc := range *s {
				f := errors.Frame(pc)
				fmt.Fprintf(st, "\n%+v", f)
			}
		}
	}
}

func (s *stack) StackTrace() errors.StackTrace {
	f := make([]errors.Frame, len(*s))
	for i := 0; i < len(f); i++ {
		f[i] = errors.Frame((*s)[i])
	}
	return f
}

func callers(skip int) *stack {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(2+skip, pcs[:])
	var st stack = pcs[0:n]
	return &st
}

func (f *ErrorWithFields) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%+v", f.error)
			for key, value := range f.fields {
				fmt.Fprintf(s, ". %+v: %+v", key, value)
			}
			f.stack.Format(s, verb)
			return
		}
		fallthrough
	case 's':
		io.WriteString(s, f.Error())
		for key, value := range f.fields {
			fmt.Fprintf(s, ". %v: %v", key, value)
		}
	case 'q':
		fmt.Fprintf(s, "%q", f.Error())
		for key, value := range f.fields {
			fmt.Fprintf(s, ". %q: %q", key, value)
		}
	}
}
