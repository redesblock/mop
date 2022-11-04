package breaker

func NewBreakerWithCurrentTimeFn(o Options, currentTimeFn currentTimeFn) Interface {
	return newBreakerWithCurrentTimeFn(o, currentTimeFn)
}
