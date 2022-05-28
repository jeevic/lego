package define

const (
	DefaultBreaker BreakerType = iota
	GoogleBreaker
)

type BreakerType int
