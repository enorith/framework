package framework

import "github.com/enorith/framework/http"

//StandardAppStructure for application bootstrap
type StandardAppStructure struct {
	//application base path
	BasePath string
	//http middleware group
	MiddlewareGroup map[string][]http.RequestMiddleware
	//http middleware
	Middleware []http.RequestMiddleware
}
