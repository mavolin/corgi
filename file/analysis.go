package file

type AnalysisCondition[T comparable] interface {
	Analysis[bool] | AnalysisWithReason[T]

	true() bool
	false() bool
	zero() T
}

// ConditionalAnalysis returns an analysis result that is dependent on the
// condition analysis:
//
// If the condition is true, result is returned.
// If the condition or the result is false, zero is returned.
// In any other case, namely if the condition is true and the result failed,
// or if the condition failed and the result is true or failed, a failed
// analysis is returned.
func ConditionalAnalysis[C, R comparable, CA AnalysisCondition[C]](condition CA, result AnalysisWithReason[R]) (final AnalysisWithReason[R]) {
	switch {
	case condition.true():
		return result
	case condition.false() || result.False():
		final.SetFalse()
	default:
		final.SetFailed()
	}
	return final
}

// SliceRef is a wrapper type around a slice, so that it can fulfil the
// comparable constraint, as needed for Analysis* types.
//
// A zero SliceRef is not valid, use NilSliceRef to get a valid zero value.
type SliceRef[T any] struct {
	s *[]T
}

func SliceRefFrom[T any](s []T) SliceRef[T] { return SliceRef[T]{s: &s} }

func NilSliceRef[T any]() SliceRef[T] {
	var s []T
	return SliceRef[T]{s: &s}
}

func (r SliceRef[T]) Get() []T { return *r.s }
func (r SliceRef[T]) Len() int { return len(*r.s) }

// ============================================================================
// Analysis
// ======================================================================================

// Analysis is a boolean analysis, i.e. one without reason.
type Analysis[T comparable] struct {
	result T
	ok     bool
}

func (a Analysis[T]) Successful() bool { return a.ok }
func (a Analysis[T]) Failed() bool     { return !a.Successful() }

// Result returns the result of the analysis.
// The analysis must be successful.
//
// Use ResultOr to get a fallback value if the analysis failed.
func (a Analysis[T]) Result() T {
	if a.Failed() {
		panic("Result called on a failed analysis")
	}
	return a.result
}

func (a Analysis[T]) ResultOr(fallback T) T {
	if a.Failed() {
		return fallback
	}
	return a.Result()
}

func (a *Analysis[T]) SetResult(v T) { a.result, a.ok = v, true }
func (a *Analysis[T]) SetFailed()    { a.result, a.ok = a.zero(), false }
func (a *Analysis[T]) SetZero()      { a.result, a.ok = a.zero(), true }

// NotZero indicates whether the analysis was successful and the result is not
// the zero value.
//
// Do not negate, the negation is probably not what you expect!
func (a Analysis[T]) NotZero() bool { return a.true() }

func (a Analysis[T]) Equal(v T) bool { return a.Successful() && a.result == v }
func (a Analysis[T]) true() bool     { return a.Successful() && a.result != a.zero() }
func (a Analysis[T]) false() bool    { return a.Successful() && a.result == a.zero() }

func (a Analysis[T]) zero() T {
	var zero T
	return zero
}

// ============================================================================
// Analysis With Reason
// ======================================================================================

// AnalysisWithReason is an analysis with a reason.
//
// The zero value is a failed analysis.
type AnalysisWithReason[T comparable] struct {
	a Analysis[T]
}

func (a AnalysisWithReason[T]) Successful() bool { return a.a.Successful() }
func (a AnalysisWithReason[T]) Failed() bool     { return a.a.Failed() }
func (a AnalysisWithReason[T]) True() bool       { return a.a.Successful() && a.a.result != a.zero() }
func (a AnalysisWithReason[T]) False() bool      { return a.a.Successful() && a.a.result == a.zero() }

func (a AnalysisWithReason[T]) Reason() T {
	if a.Failed() {
		panic("Reason called on a failed analysis")
	}
	return a.a.result
}

func (a *AnalysisWithReason[T]) SetReason(v T) { a.a.SetResult(v) }
func (a *AnalysisWithReason[T]) SetFailed()    { a.a.SetFailed() }
func (a *AnalysisWithReason[T]) SetFalse()     { a.a.SetZero() }

func (a AnalysisWithReason[T]) true() bool  { return a.True() }
func (a AnalysisWithReason[T]) false() bool { return a.False() }
func (a AnalysisWithReason[T]) zero() T     { return a.a.zero() }
