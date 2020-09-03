package errors

import (
	"github.com/enorith/framework/exception"
	"github.com/enorith/framework/http/content"
	"github.com/enorith/framework/http/contract"
	"github.com/enorith/framework/kernel"
)

type ErrorHandler interface {
	HandleError(e interface{}, r contract.RequestContract) contract.ResponseContract
}

type StandardErrorHandler struct {
	App *kernel.Application
}

func (h *StandardErrorHandler) HandleError(e interface{}, r contract.RequestContract) contract.ResponseContract {
	return h.BaseHandle(e, r)
}

func (h *StandardErrorHandler) BaseHandle(e interface{}, r contract.RequestContract) contract.ResponseContract {
	var ex exception.Exception
	var code = 500
	var headers map[string]string = nil
	if t, ok := e.(string); ok {
		ex = exception.NewException(t, code)
	} else if t, ok := e.(exception.HttpException); ok {
		ex = t
		headers = t.Headers()
	} else if t, ok := e.(exception.Exception); ok {
		ex = t
	} else if t, ok := e.(error); ok {
		ex = exception.NewExceptionFromError(t, code)
	} else {
		ex = exception.NewException("undefined exception", code)
	}

	if t, ok := e.(contract.WithStatusCode); ok {
		code = t.StatusCode()
	}

	if r.ExceptsJson() {
		return content.JsonErrorResponseFormatter(ex, code, h.App.Debug(), headers)
	} else {
		//tmp := fmt.Sprintf("%s/errors/%d.html", h.App.Structure().BasePath, code)
		//fmt.Println(tmp)
		//if e, _ := file.PathExists(tmp); e {
		//	return content.FileResponse(tmp, 200, ex)
		//}
		return content.HtmlErrorResponseFormatter(ex, code, h.App.Debug(), headers)
	}
}
