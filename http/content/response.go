package content

import (
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/enorith/framework/exception"
	"github.com/enorith/framework/http/contracts"
)

var (
	ContentTypeJson = "application/json; charset=utf-8"
	ContentTypeHtml = "text/html; charset=utf-8"
	DefaultHeader   = func() map[string]string {
		return map[string]string{}
	}
	TextHeader = func() map[string]string {
		return map[string]string{
			"Content-Type": ContentTypeHtml,
		}
	}
	JsonHeader = func() map[string]string {
		return map[string]string{
			"Content-Type": ContentTypeJson,
		}
	}
	HtmlHeader = func() map[string]string {
		return map[string]string{
			"Content-Type": ContentTypeHtml,
		}
	}
)

//Response http response
type Response struct {
	content    []byte
	headers    map[string]string
	statusCode int
	handled    bool
}

func (r *Response) SetStatusCode(code int) {
	r.statusCode = code
}

func (r *Response) Handled() bool {
	return r.handled
}

func (r *Response) SetHeader(key string, value string) contracts.ResponseContract {
	r.headers[key] = value
	return r
}

//Content response body
func (r *Response) Content() []byte {
	return r.content
}

//Headers response headers
func (r *Response) Headers() map[string]string {
	return r.headers
}

//WithStatusCode status code
func (r *Response) StatusCode() int {
	return r.statusCode
}

type ErrorResponse struct {
	*Response
	e exception.Exception
}

func (e *ErrorResponse) E() exception.Exception {
	return e.e
}

type TemplateResponse struct {
	*Response
	template     *template.Template
	templateData interface{}
}

func (t *TemplateResponse) TemplateData() interface{} {
	return t.templateData
}

func (t *TemplateResponse) SetTemplateData(templateData interface{}) *TemplateResponse {
	t.templateData = templateData
	return t
}

func (t *TemplateResponse) Template() *template.Template {
	return t.template
}

type File struct {
	*Response
	path string
}

func (f *File) Path() string {
	return f.path
}

func NewResponse(content []byte, headers map[string]string, code int) *Response {
	// copy headers when new a response
	hs := make(map[string]string)
	for k, v := range headers {
		hs[k] = v
	}

	return &Response{
		content:    content,
		headers:    hs,
		statusCode: code,
	}
}

func HtmlResponse(content string, code int) *Response {
	return NewResponse([]byte(content), HtmlHeader(), code)
}

func ErrResponse(e exception.Exception, code int, headers map[string]string) *ErrorResponse {
	return &ErrorResponse{
		NewResponse([]byte(e.Error()), headers, code),
		e,
	}
}

func ErrResponseFromError(e error, code int, headers map[string]string) *ErrorResponse {
	var ex exception.Exception
	if es, ok := e.(contracts.WithStatusCode); ok {
		ex = exception.NewHttpExceptionFromError(e, es.StatusCode(), 0, headers)
		headers = nil
	} else {
		ex = exception.NewExceptionFromError(e, 500)
	}

	return ErrResponse(ex, code, headers)
}

func HttpErrorResponse(message string, statusCode int, code int, headers map[string]string) *ErrorResponse {
	return ErrResponse(exception.NewHttpException(message, statusCode, code, headers), statusCode, headers)
}

func NotFoundResponse(message string) *ErrorResponse {
	return ErrResponse(exception.NewHttpException(message, 404, 404, nil), 404, nil)
}

func ErrResponseFromOrigin(resp *Response) *ErrorResponse {
	return &ErrorResponse{
		resp,
		nil,
	}
}

func TextResponse(content string, code int) *Response {

	return NewResponse([]byte(content), TextHeader(), code)
}

func JsonResponse(data interface{}, code int, headers map[string]string) contracts.ResponseContract {
	var j []byte
	var err error
	if b, ok := data.([]byte); ok {
		j = b
	} else if t, ok := data.(json.Marshaler); ok {
		j, err = t.MarshalJSON()
		if err != nil {
			return ErrResponseFromError(err, 500, nil)
		}
	} else {
		j, err = json.Marshal(data)
		if err != nil {
			return ErrResponseFromError(err, 500, nil)
		}
	}
	if headers == nil {
		headers = map[string]string{}
	}

	headers["Content-Type"] = "application/json; charset=utf-8"

	return NewResponse(j, headers, code)
}

func TempResponse(t *template.Template, code int, data interface{}) *TemplateResponse {
	return &TemplateResponse{
		NewResponse(nil, HtmlHeader(), code),
		t,
		data,
	}
}

func NewHandledResponse(code ...int) *Response {

	var c int
	if len(code) < 1 {
		c = 200
	} else {
		c = code[0]
	}
	hs := make(map[string]string)

	return &Response{
		handled:    true,
		statusCode: c,
		headers:    hs,
	}
}

type JsonMessage int

func (m JsonMessage) MarshalJSON() ([]byte, error) {
	code := int(m)
	return json.Marshal(map[string]interface{}{
		"message": http.StatusText(code),
		"code":    code,
	})
}
