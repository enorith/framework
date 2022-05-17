package http

import (
	"sync"
	"time"

	"github.com/enorith/http/contracts"
	"go.uber.org/zap"
)

type LoggingOption func(request contracts.RequestContract, statusCode int, startAt time.Time) zap.Field

var (
	loggingOptions []LoggingOption
	mu             sync.RWMutex
)

func WithLogginOption(option LoggingOption) {
	mu.Lock()
	defer mu.Unlock()
	loggingOptions = append(loggingOptions, option)
}
