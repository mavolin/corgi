package file

import (
	"testing"

	"github.com/mavolin/corgi/v2/internal/should"
)

func TestConditionalAnalysis(t *testing.T) {
	tests := []struct {
		name       string
		condition  Analysis[bool]
		result     AnalysisWithReason[string]
		wantReason string
		wantFailed bool
	}{
		{
			name:      "condition true, result successful with non-zero value",
			condition: Analysis[bool]{result: true, ok: true},
			result: func() AnalysisWithReason[string] {
				a := AnalysisWithReason[string]{}
				a.a.SetResult("success")
				return a
			}(),
			wantReason: "success",
		},
		{
			name:       "condition true, result successful with zero value",
			condition:  Analysis[bool]{result: true, ok: true},
			result:     func() AnalysisWithReason[string] { a := AnalysisWithReason[string]{}; a.a.SetZero(); return a }(),
			wantReason: "",
		},
		{
			name:       "condition true, result failed",
			condition:  Analysis[bool]{result: true, ok: true},
			result:     AnalysisWithReason[string]{},
			wantFailed: true,
		},
		{
			name:      "condition false, result successful with non-zero value",
			condition: Analysis[bool]{result: false, ok: true},
			result: func() AnalysisWithReason[string] {
				a := AnalysisWithReason[string]{}
				a.a.SetResult("success")
				return a
			}(),
			wantReason: "",
		},
		{
			name:       "condition false, result successful with zero value",
			condition:  Analysis[bool]{result: false, ok: true},
			result:     func() AnalysisWithReason[string] { a := AnalysisWithReason[string]{}; a.a.SetZero(); return a }(),
			wantReason: "",
		},
		{
			name:       "condition false, result failed",
			condition:  Analysis[bool]{result: false, ok: true},
			result:     AnalysisWithReason[string]{},
			wantReason: "",
		},
		{
			name:      "condition failed, result successful with non-zero value",
			condition: Analysis[bool]{result: false, ok: false},
			result: func() AnalysisWithReason[string] {
				a := AnalysisWithReason[string]{}
				a.a.SetResult("success")
				return a
			}(),
			wantFailed: true,
		},
		{
			name:       "condition failed, result successful with zero value",
			condition:  Analysis[bool]{result: false, ok: false},
			result:     func() AnalysisWithReason[string] { a := AnalysisWithReason[string]{}; a.a.SetZero(); return a }(),
			wantReason: "",
		},
		{
			name:       "condition failed, result failed",
			condition:  Analysis[bool]{result: false, ok: false},
			result:     AnalysisWithReason[string]{},
			wantFailed: true,
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			got := ConditionalAnalysis(c.condition, c.result)

			should.Equal(t, got.Failed(), c.wantFailed)
			if got.Successful() {
				should.Equal(t, got.Reason(), c.wantReason)
			}
		})
	}
}

// ============================================================================
// Analysis
// ======================================================================================

func TestAnalysis_Result(t *testing.T) {
	tests := []struct {
		name   string
		in     Analysis[int]
		panic  bool
		result int
	}{
		{name: "nonzero", in: Analysis[int]{result: 42, ok: true}, result: 42},
		{name: "zero", in: Analysis[int]{result: 0, ok: true}},
		{name: "failed", in: Analysis[int]{result: 0, ok: false}, panic: true},
	}
	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					should.Equal(t, c.panic, true)
				} else {
					should.Equal(t, c.panic, false)
				}
			}()
			res := c.in.Result()
			should.Equal(t, res, c.result)
		})
	}
}

func TestAnalysis_ResultOr(t *testing.T) {
	t.Parallel()

	t.Run("successful", func(t *testing.T) {
		t.Parallel()

		a := Analysis[int]{result: 0, ok: true}
		should.Equal(t, a.ResultOr(99), 0)
	})

	t.Run("failed", func(t *testing.T) {
		t.Parallel()

		a := Analysis[int]{result: 0, ok: false}
		should.Equal(t, a.ResultOr(99), 99)
	})
}

func TestAnalysis_SetResult(t *testing.T) {
	t.Parallel()

	var a Analysis[int]

	a.SetResult(42)
	should.Equal(t, a.result, 42)
	should.Equal(t, a.ok, true)
}

func TestAnalysis_SetFailed(t *testing.T) {
	t.Parallel()

	var a Analysis[int]

	a.SetFailed()
	should.Equal(t, a.result, 0)
	should.Equal(t, a.ok, false)
}

func TestAnalysis_SetZero(t *testing.T) {
	t.Parallel()

	var a Analysis[int]

	a.SetZero()
	should.Equal(t, a.result, 0)
	should.Equal(t, a.ok, true)
}

func TestAnalysis_NotZero(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input Analysis[int]
		want  bool
	}{
		{name: "zero", input: Analysis[int]{result: 0, ok: true}, want: false},
		{name: "nonzero", input: Analysis[int]{result: 1, ok: true}, want: true},
		{name: "failed", input: Analysis[int]{result: 0, ok: false}, want: false},
	}
	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			should.Equal(t, c.input.NotZero(), c.want)
		})
	}
}

func TestAnalysis_Equal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input Analysis[int]
		cmp   int
		want  bool
	}{
		{name: "equal-success", input: Analysis[int]{result: 42, ok: true}, cmp: 42, want: true},
		{name: "notequal-success", input: Analysis[int]{result: 42, ok: true}, cmp: 99, want: false},
		{name: "failed", input: Analysis[int]{result: 42, ok: false}, cmp: 42, want: false},
	}
	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			should.Equal(t, c.input.Equal(c.cmp), c.want)
		})
	}
}

func TestAnalysis_true(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   Analysis[int]
		want bool
	}{
		{name: "nonzero", in: Analysis[int]{result: 1, ok: true}, want: true},
		{name: "zero", in: Analysis[int]{result: 0, ok: true}, want: false},
		{name: "failed", in: Analysis[int]{result: 0, ok: false}, want: false},
	}
	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			should.Equal(t, c.in.true(), c.want)
		})
	}
}

func TestAnalysis_false(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   Analysis[int]
		want bool
	}{
		{name: "zero", in: Analysis[int]{result: 0, ok: true}, want: true},
		{name: "nonzero", in: Analysis[int]{result: 1, ok: true}, want: false},
		{name: "failed", in: Analysis[int]{result: 0, ok: false}, want: false},
	}
	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			should.Equal(t, c.in.false(), c.want)
		})
	}
}

func TestAnalysis_zero(t *testing.T) {
	t.Parallel()

	t.Run("int", func(t *testing.T) {
		t.Parallel()

		var a Analysis[int]
		should.Equal(t, a.zero(), 0)
	})

	t.Run("string", func(t *testing.T) {
		t.Parallel()

		var a Analysis[string]
		should.Equal(t, a.zero(), "")
	})
}

// ============================================================================
// Analysis With Reason
// ======================================================================================

func TestAnalysisWithReason_Reason(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		in     AnalysisWithReason[int]
		panic  bool
		reason int
	}{
		{name: "nonzero", in: AnalysisWithReason[int]{a: Analysis[int]{result: 42, ok: true}}, reason: 42},
		{name: "zero", in: AnalysisWithReason[int]{a: Analysis[int]{result: 0, ok: true}}},
		{name: "failed", in: AnalysisWithReason[int]{a: Analysis[int]{result: 0, ok: false}}, panic: true},
	}
	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			defer func() {
				if r := recover(); r != nil {
					should.Equal(t, c.panic, true)
				} else {
					should.Equal(t, c.panic, false)
				}
			}()
			should.Equal(t, c.in.Reason(), c.reason)
		})
	}
}

func TestAnalysisWithReason_SetReason(t *testing.T) {
	t.Parallel()

	a := AnalysisWithReason[int]{}
	a.SetReason(42)
	should.Equal(t, a.Reason(), 42)
	should.Equal(t, a.Successful(), true)

	a.SetReason(99)
	should.Equal(t, a.Reason(), 99)
}

func TestAnalysisWithReason_SetFailed(t *testing.T) {
	t.Parallel()

	a := AnalysisWithReason[int]{}

	a.SetFailed()
	should.Equal(t, a.Successful(), false)
}

func TestAnalysisWithReason_SetFalse(t *testing.T) {
	t.Parallel()

	a := AnalysisWithReason[int]{}
	a.SetFalse()
	should.Equal(t, a.Reason(), 0)
	should.True(t, a.Successful())
}

func TestAnalysisWithReason_NotZero(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   AnalysisWithReason[int]
		want bool
	}{
		{name: "nonzero", in: AnalysisWithReason[int]{a: Analysis[int]{result: 42, ok: true}}, want: true},
		{name: "zero", in: AnalysisWithReason[int]{a: Analysis[int]{result: 0, ok: true}}},
		{name: "failed", in: AnalysisWithReason[int]{a: Analysis[int]{result: 0, ok: false}}},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			should.Equal(t, c.in.True(), c.want)
		})
	}
}

func TestAnalysisWithReason_Successful(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   AnalysisWithReason[int]
		want bool
	}{
		{name: "nonzero", in: AnalysisWithReason[int]{a: Analysis[int]{result: 42, ok: true}}, want: true},
		{name: "zero", in: AnalysisWithReason[int]{a: Analysis[int]{result: 0, ok: true}}, want: true},
		{name: "failed", in: AnalysisWithReason[int]{a: Analysis[int]{result: 0, ok: false}}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			should.Equal(t, tc.in.Successful(), tc.want)
		})
	}
}

func TestAnalysisWithReason_Failed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   AnalysisWithReason[int]
		want bool
	}{
		{name: "nonzero", in: AnalysisWithReason[int]{a: Analysis[int]{result: 42, ok: true}}},
		{name: "zero", in: AnalysisWithReason[int]{a: Analysis[int]{result: 0, ok: true}}},
		{name: "failed", in: AnalysisWithReason[int]{a: Analysis[int]{result: 0, ok: false}}, want: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			should.Equal(t, tc.in.Failed(), tc.want)
		})
	}
}

func TestAnalysisWithReason_True(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   AnalysisWithReason[int]
		want bool
	}{
		{name: "nonzero", in: AnalysisWithReason[int]{a: Analysis[int]{result: 42, ok: true}}, want: true},
		{name: "zero", in: AnalysisWithReason[int]{a: Analysis[int]{result: 0, ok: true}}},
		{name: "failed", in: AnalysisWithReason[int]{a: Analysis[int]{result: 0, ok: false}}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			should.Equal(t, tc.in.True(), tc.want)
		})
	}
}

func TestAnalysisWithReason_False(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   AnalysisWithReason[int]
		want bool
	}{
		{name: "nonzero", in: AnalysisWithReason[int]{a: Analysis[int]{result: 42, ok: true}}},
		{name: "zero", in: AnalysisWithReason[int]{a: Analysis[int]{result: 0, ok: true}}, want: true},
		{name: "failed", in: AnalysisWithReason[int]{a: Analysis[int]{result: 0, ok: false}}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			should.Equal(t, tc.in.False(), tc.want)
		})
	}
}

func TestAnalysisWithReason_true(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   AnalysisWithReason[int]
		want bool
	}{
		{name: "nonzero", in: AnalysisWithReason[int]{a: Analysis[int]{result: 42, ok: true}}, want: true},
		{name: "zero", in: AnalysisWithReason[int]{a: Analysis[int]{result: 0, ok: true}}},
		{name: "failed", in: AnalysisWithReason[int]{a: Analysis[int]{result: 0, ok: false}}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			should.Equal(t, tc.in.true(), tc.want)
		})
	}
}

func TestAnalysisWithReason_false(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   AnalysisWithReason[int]
		want bool
	}{
		{name: "nonzero", in: AnalysisWithReason[int]{a: Analysis[int]{result: 42, ok: true}}},
		{name: "zero", in: AnalysisWithReason[int]{a: Analysis[int]{result: 0, ok: true}}, want: true},
		{name: "failed", in: AnalysisWithReason[int]{a: Analysis[int]{result: 0, ok: false}}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			should.Equal(t, tc.in.false(), tc.want)
		})
	}
}
