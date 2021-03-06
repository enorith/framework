package contracts

import (
	"context"
	"github.com/enorith/framework/contracts"
	"mime/multipart"
)

//RequestContract is interface of http request
type RequestContract interface {
	Context() context.Context
	Params() map[string][]byte
	Param(key string) string
	ParamInt64(key string) (int64, error)
	ParamUint64(key string) (uint64, error)
	ParamInt(key string) (int, error)
	SetParams(params map[string][]byte)
	SetParamsSlice(paramsSlice [][]byte)
	ParamsSlice() [][]byte
	ParamBytes(key string) []byte
	Accepts() []byte
	ExceptsJson() bool
	RequestWithJson() bool
	IsXmlHttpRequest() bool
	GetMethod() string
	GetPathBytes() []byte
	GetUri() []byte
	Get(key string) []byte
	File(key string) (UploadFile, error)
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

type UploadFile interface {
	Save(dist string) error
	Open() (multipart.File, error)
	Close() error
	Filename() string
}
