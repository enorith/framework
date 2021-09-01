package contracts

type Job interface {
	Handle() error
}
