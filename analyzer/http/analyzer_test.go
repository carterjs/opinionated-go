package http_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/carterjs/opinionated-go/analyzer/http"
)

func TestNoHTTPFramework(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), http.NoHTTPFramework, "httpframework")
}
