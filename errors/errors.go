// Package errors provides drop-in replacement for github.com/pkg/errors
// which provides simple error handling primitives.
//
// The traditional error handling idiom in Go is roughly akin to
//
//	if err != nil {
//	        return err
//	}
//
// which when applied recursively up the call stack results in error reports
// without context or debugging information. The errors package allows
// programmers to add context to the failure path in their code in a way
// that does not destroy the original value of the error.
//
// # Adding context to an error
//
// The errors.Wrap function returns a new error that adds context to the
// original error by recording a stack trace at the point Wrap is called,
// together with the supplied message. For example
//
//	_, err := ioutil.ReadAll(r)
//	if err != nil {
//	        return errors.Wrap(err, "read failed")
//	}
//
// If additional control is required, the errors.WithStack and
// errors.WithMessage functions destructure errors.Wrap into its component
// operations: annotating an error with a stack trace and with a message,
// respectively.
//
// # Retrieving the cause of an error
//
// Using errors.Wrap constructs a stack of errors, adding context to the
// preceding error. Depending on the nature of the error it may be necessary
// to reverse the operation of errors.Wrap to retrieve the original error
// for inspection. Any error value which implements this interface
//
//	type causer interface {
//	        Cause() error
//	}
//
// can be inspected by errors.Cause. errors.Cause will recursively retrieve
// the topmost error that does not implement causer, which is assumed to be
// the original cause. For example:
//
//	switch err := errors.Cause(err).(type) {
//	case *MyError:
//	        // handle specifically
//	default:
//	        // unknown error
//	}
//
// Although the causer interface is not exported by this package, it is
// considered a part of its stable public interface.
//
// # Formatted printing of errors
//
// All error values returned from this package implement fmt.Formatter and can
// be formatted by the fmt package. The following verbs are supported:
//
//	%s    print the error. If the error has a Cause it will be
//	      printed recursively.
//	%v    see %s
//	%+v   extended format. Each Frame of the error's StackTrace will
//	      be printed in detail.
//
// # Retrieving the stack trace of an error or wrapper
//
// New, Errorf, Wrap, and Wrapf record a stack trace at the point they are
// invoked. This information can be retrieved with the following interface:
//
//	type stackTracer interface {
//	        StackTrace() errors.StackTrace
//	}
//
// The returned errors.StackTrace type is defined as
//
//	type StackTrace []Frame
//
// The Frame type represents a call site in the stack trace. Frame supports
// the fmt.Formatter interface that can be used for printing information about
// the stack trace of this error. For example:
//
//	if err, ok := err.(stackTracer); ok {
//	        for _, f := range err.StackTrace() {
//	                fmt.Printf("%+s:%d\n", f, f)
//	        }
//	}
//
// Although the stackTracer interface is not exported by this package, it is
// considered a part of its stable public interface.
//
// See the documentation for Frame.Format for more details.

package errors

import (
	stderrors "errors"
	"fmt"

	"emperror.dev/errors"
	pkgerrors "github.com/pkg/errors"
)

var (
	stackDepth = 1
)

func ResetStackDepth(d int) {
	stackDepth = d
}

type Sentinel = errors.Sentinel

// New returns an error with the supplied message.
// New also records the stack trace at the point it was called.
func New(message string) error {
	return errors.WithStackDepthIf(stderrors.New(message), stackDepth)
}

// Errorf formats according to a format specifier and returns the string
// as a value that satisfies error.
// Errorf also records the stack trace at the point it was called.
func Errorf(format string, args ...interface{}) error {
	return errors.WithStackDepthIf(fmt.Errorf(format, args...), stackDepth)
}
func ErrorfWithRelativeStackDepth(depth int, format string, args ...interface{}) error {
	return errors.WithStackDepthIf(fmt.Errorf(format, args...), stackDepth+depth)
}

// WithStack annotates err with a stack trace at the point WithStack was called.
// If err is nil, WithStack returns nil.
func WithStack(err error) error {
	return errors.WithStackDepthIf(err, stackDepth)
}

// Wrap returns an error annotating err with a stack trace
// at the point Wrap is called.
// If err is nil, Wrap returns nil.
func Wrap(err error) error {
	return errors.WithStackDepthIf(err, stackDepth)
}

func WrapWithRelativeStackDepth(err error, depth int) error {
	return errors.WithStackDepthIf(err, stackDepth+depth)
}

type WithMsgfIf interface {
	WithErrorf(format string, args ...interface{}) error
}

// Wrapf returns an error annotating err with a stack trace
// at the point Wrapf is called, and the format specifier.
// If err is nil, Wrapf returns nil.
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	werr, ok := err.(WithMsgfIf)
	if ok {
		return werr.WithErrorf(format, args...)
	}

	return errors.WithStackDepthIf(pkgerrors.WithMessagef(err, format, args...), stackDepth)
}

func WrapfWithRelativeStackDepth(err error, depth int, format string, args ...any) error {
	return errors.WithStackDepthIf(pkgerrors.WithMessagef(err, format, args...), stackDepth+depth)
}

// Cause returns the underlying cause of the error, if possible.
// An error value has a cause if it implements the following
// interface:
//
//	type causer interface {
//	       Cause() error
//	}
//
// If the error does not implement Cause, the original error will
// be returned. If the error is nil, nil will be returned without further
// investigation.
func Cause(err error) error {
	return errors.Cause(err)
}

func StackTrace(err error) errors.StackTrace {
	var st ErrorWithStackTrace

	if As(err, &st) {
		return st.StackTrace()
	}

	return nil
}

func Frames(err error) []errors.Frame {
	var st ErrorWithStackTrace

	if As(err, &st) {
		return st.StackTrace()
	}

	return nil
}

type ErrorWithStackTrace interface {
	StackTrace() errors.StackTrace
}
