package errorc_test

import (
	"fmt"
	"testing"

	"github.com/chein-huang/errorc"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestError(t *testing.T) {
	err := Call()
	fmt.Printf("%+v\n", err)
	fmt.Println(errorc.Wrap(err).LogFields())

	a, ok := errorc.GetErrorWithFields(err).Unwrap().(IA)
	if !ok {
		t.Error("not IA")
		return
	}
	fmt.Println(a.FuncA())
}

func TestIs(t *testing.T) {
	assert := require.New(t)

	err := &Error{S: "ss", I: 1}
	wErr := errorc.AddField(err, "BB", "CC")
	assert.True(errors.Is(wErr, err))
}

func Call() error {
	return errorc.AddField(&Error{S: "ss", I: 1}, "BB", "CC")
}

type IA interface {
	FuncA() string
}

type Error struct {
	S string
	I int
}

func (e *Error) Error() string {
	return fmt.Sprintf("%v-%v", e.S, e.I)
}

func (e *Error) FuncA() string {
	return e.S
}
