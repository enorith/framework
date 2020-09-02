package exception

import (
	"errors"
	goErrors "github.com/go-errors/errors"
)

type ErrorParser func(err error, code int, skipTrace int) *StandardException
type TraceParser func(err error, skip int) []Trace

func GoErrorTraceParser(err error, skip int) []Trace {

	e := goErrors.New(err)
	var ts []Trace
	for k, f := range e.StackFrames() {
		if k < skip {
			continue
		}
		ts = append(ts, &StandardTrace{
			file: f.File,
			line: f.LineNumber,
			name: f.Name,
		})
	}
	return ts
}

func GoErrorParser(err error, code int, skipTrace int) *StandardException {
	e := goErrors.New(err)
	return &StandardException{
		message: e.Error(),
		code:    code,
		traces:  GoErrorTraceParser(e, skipTrace),
	}
}

type Trace interface {
	File() string
	Line() int
	Name() string
}

type Exception interface {
	error
	Code() int
	File() string
	Line() int
	Traces() []Trace
}

type HttpException interface {
	Exception
	StatusCode() int
	Headers() map[string]string
}

type StandardTrace struct {
	file string
	line int
	name string
}

func (t *StandardTrace) File() string {
	return t.file
}

func (t *StandardTrace) Line() int {
	return t.line
}

func (t *StandardTrace) Name() string {
	return t.name
}

type StandardException struct {
	message string
	code    int
	traces  []Trace
}

func (e *StandardException) Error() string {
	return e.message
}

func (e *StandardException) Code() int {
	return e.code
}

func (e *StandardException) File() string {
	if len(e.Traces()) > 0 {
		return e.Traces()[0].File()
	}

	return ""
}

func (e *StandardException) Line() int {
	if len(e.Traces()) > 0 {
		return e.Traces()[0].Line()
	}

	return 0
}

func (e *StandardException) Traces() []Trace {
	return e.traces
}

type StandardHttpException struct {
	*StandardException
	statusCode int
	headers    map[string]string
}

func (s *StandardHttpException) Headers() map[string]string {
	return s.headers
}

func (s *StandardHttpException) StatusCode() int {
	return s.statusCode
}

func NewHttpException(message string, statusCode int, code int, headers map[string]string) *StandardHttpException {
	return &StandardHttpException{
		&StandardException{
			message: message,
			code:    code,
			traces:  GoErrorTraceParser(errors.New(message), 2),
		},
		statusCode,
		headers,
	}
}
func NewHttpExceptionFromError(err error, statusCode int, code int, headers map[string]string) *StandardHttpException {
	return &StandardHttpException{
		GoErrorParser(err, code, 3),
		statusCode,
		headers,
	}
}

func NewException(message string, code int) *StandardException {
	return &StandardException{
		message: message,
		code:    code,
		traces:  GoErrorTraceParser(errors.New(message), 2),
	}
}

func NewExceptionFromError(err error, code int) *StandardException {
	return GoErrorParser(err, code, 3)
}
