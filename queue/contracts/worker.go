package contracts

import "os"

type Worker interface {
	Run(concurrency int, done chan os.Signal)
}
