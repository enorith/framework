package contract

import (
	"context"
	"github.com/enorith/framework/contracts"
)

//RequestContract is interface of http request
type RequestContract interface {
	Context() context.Context
	Params() map[string]string
	Param(key string) string
	ParamInt64(key string) (int64, error)
	ParamUint64(key string) (uint64, error)
	ParamInt(key string) (int, error)
	SetParams(params map[string]string)
	Accepts() []byte
	ExceptsJson() bool
	IsXmlHttpRequest() bool
	GetMethod() string
	GetPathBytes() []byte
	GetUri() []byte
	Get(key string) []byte
	GetInt64(key string) (int64, error)
	GetUint64(key string) (uint64, error)
	GetString(key string) string
	GetInt(key string) (int, error)
	GetClientIp() string
	GetContent() []byte
	Unmarshal(to interface{}) error
	GetSignature() []byte
	Header(key string) []byte
	HeaderString(key string) string
	SetHeader(key string, value []byte) RequestContract
	SetHeaderString(key, value string) RequestContract
	Authorization() []byte
	BearerToken() ([]byte, error)
	User() contracts.User
	SetUser(u contracts.User)
}
