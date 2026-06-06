package comments_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/carterjs/opinionated-go/analyzer/comments"
)

func TestExportedCommentFormat(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), comments.ExportedCommentFormat, "exportedcomment")
}
