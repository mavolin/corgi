package safe

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsSafeURLScheme(t *testing.T) {
	t.Parallel()

	successCases := []string{
		"http", "https", "mailto", "tel",
		"hTTp", "httPs", "maiLTO", "TEL",
	}
	failureCases := []string{
		"javascript", "data", "file",
		"javaSCRIPT", "DATA", "FILE",
	}

	runIsSafeSchemeSuite(t, successCases, failureCases, IsSafeURLScheme)
}

func TestIsSafeResourceURLScheme(t *testing.T) {
	successCases := []string{"https", "httPs", "wss", "WSS"}
	developSuccessCases := []string{"http", "htTTp", "ws", "WS"}
	failureCases := []string{
		"javascript", "data", "file",
		"javaSCRIPT", "DATA", "FILE",
	}

	t.Run("development mode", func(t *testing.T) {
		successCases := append(successCases, developSuccessCases...)
		developmentMode(func() {
			runIsSafeSchemeSuite(t, successCases, failureCases, IsSafeResourceURLScheme)
		})
	})

	t.Run("production mode", func(t *testing.T) {
		failureCases := append(failureCases, developSuccessCases...)
		runIsSafeSchemeSuite(t, successCases, failureCases, IsSafeResourceURLScheme)
	})
}

func runIsSafeSchemeSuite(t *testing.T, successCases, failureCases []string, cmpF func(string) bool) {
	t.Helper()

	t.Run("success", func(t *testing.T) {
		for _, c := range successCases {
			scheme := strings.SplitN(c, ":", 2)[0]

			t.Run(scheme, func(t *testing.T) {
				t.Parallel()
				assert.True(t, cmpF(c))
			})
		}
	})

	t.Run("failure", func(t *testing.T) {
		t.Parallel()
		for _, c := range failureCases {
			scheme := strings.SplitN(c, ":", 2)[0]

			t.Run(scheme, func(t *testing.T) {
				t.Parallel()
				assert.False(t, cmpF(c))
			})
		}
	})
}

func TestTrustedURLAttr(t *testing.T) {
	t.Parallel()

	expect := "foo"
	actual := TrustedURLAttr(expect)
	assert.Equal(t, expect, actual.Escaped())
}

func TestTrustedURLListAttr(t *testing.T) {
	t.Parallel()

	expect := "foo"
	actual := TrustedURLListAttr(expect)
	assert.Equal(t, expect, actual.Escaped())
}

func TestTrustedResourceURLAttr(t *testing.T) {
	t.Parallel()

	expect := "foo"
	actual := TrustedResourceURLAttr(expect)
	assert.Equal(t, expect, actual.Escaped())
}
