package std

import (
	"os"

	"github.com/enorith/framework/queue/contracts"
)

type Worker struct {
	connection contracts.Connection
}

func (w *Worker) Run(concurrency int, done chan os.Signal) {
	w.connection.Consume(concurrency, done)
}
