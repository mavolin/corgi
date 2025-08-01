package file

// Analysis is the generic result of an analysis.
type Analysis[T comparable] struct {
	Result T
	Failed bool
}

func Result[T comparable](v T) (a Analysis[T]) {
	a.Set(v)
	return a
}

func ResultIf[T comparable](v T, cond bool) (a Analysis[T]) {
	a.SetIf(v, cond)
	return a
}

func FailedAnalysis[T comparable]() (a Analysis[T]) {
	a.SetFailed()
	return a
}

// ConditionalAnalysis returns an analysis result that is dependent on the
// condition analysis:
//
// If the condition is not zero, result is returned.
// If the condition or the result is zero, zero is returned.
// In any other case, namely if the condition is not zero and the result failed,
// or if the condition failed and the result is not zero or failed, a failed
// analysis is returned.
func ConditionalAnalysis[C, R comparable](condition Analysis[C], result Analysis[R]) (final Analysis[R]) {
	var cZero C
	var rZero R
	switch {
	case condition.NotZero():
		return result
	case condition.Equal(cZero) || result.Equal(rZero):
		return Result(rZero)
	default:
		return FailedAnalysis[R]()
	}
}

func (a Analysis[T]) GetOr(fallback T) T {
	if a.Failed {
		return fallback
	}
	return a.Result
}

func (a *Analysis[T]) Set(v T) {
	a.Result, a.Failed = v, false
}

// SetIf sets the result to v if cond is true, otherwise it marks the analysis
// as failed.
func (a *Analysis[T]) SetIf(v T, cond bool) {
	if cond {
		a.Set(v)
	} else {
		a.SetFailed()
	}
}

func (a *Analysis[T]) SetFailed() {
	var zero T
	a.Result, a.Failed = zero, true
}

func (a *Analysis[T]) SetZero() {
	var zero T
	a.Result, a.Failed = zero, false
}

// NotZero indicates whether the analysis was successful and the result is not
// the zero value.
//
// Do not negate, the negation is probably not what you expect!
func (a Analysis[T]) NotZero() bool {
	var zero T
	return !a.Failed && a.Result != zero
}

func (a Analysis[T]) Equal(v T) bool {
	return !a.Failed && a.Result == v
}
