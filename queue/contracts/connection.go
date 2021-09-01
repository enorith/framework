package contracts

type Connection interface {
	Consume(concurrency int) error
	Stop() error
	Dispatch(payload interface{}) error
}
