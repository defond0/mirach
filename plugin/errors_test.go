// +build unit

package plugin

import (
	"errors"
	"reflect"
	"testing"
)

func TestExceptionOrError_known(t *testing.T) {
	exceptions := []string{"known error"}
	err := ExceptionOrError(errors.New("known error"), exceptions)
	eType := reflect.TypeOf(err).String()
	expected := "plugin.Exception"
	if eType != expected {
		t.Errorf("error was of type: %s, expected: %s", eType, expected)
	}
}

func TestExceptionOrError_unknown(t *testing.T) {
	exceptions := []string{"known error"}
	err := ExceptionOrError(errors.New("unknown error"), exceptions)
	eType := reflect.TypeOf(err).String()
	expected := "*errors.errorString"
	if eType != expected {
		t.Errorf("error was of type: %s, expected: %s", eType, expected)
	}
}
