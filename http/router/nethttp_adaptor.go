package router

import (
	"net/http"

	"github.com/enorith/framework/http/content"
	"github.com/enorith/framework/http/contract"
)

func NetHttpHandlerFromHttp(request *content.NetHttpRequest, h http.Handler) contract.ResponseContract {
	r, ow := request.Origin(), request.OriginWriter()

	h.ServeHTTP(ow, r)

	return content.NewHandledResponse()
}
