package http

import (
	"go/ast"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// frameworkPrefixes lists third-party HTTP router and framework import paths.
// The standard library net/http handles routing since Go 1.22.
var frameworkPrefixes = []string{
	"github.com/gin-gonic/gin",
	"github.com/labstack/echo",
	"github.com/gofiber/fiber",
	"github.com/go-chi/chi",
	"github.com/gorilla/mux",
	"github.com/julienschmidt/httprouter",
	"github.com/beego/beego",
	"github.com/astaxie/beego",
	"github.com/valyala/fasthttp",
	"github.com/emicklei/go-restful",
}

// NoHTTPFramework warns on third-party HTTP framework imports.
var NoHTTPFramework = &analysis.Analyzer{
	Name:     "no_http_framework",
	Doc:      "warn on third-party HTTP framework imports; use net/http",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      runNoHTTPFramework,
}

func runNoHTTPFramework(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspect.Preorder([]ast.Node{(*ast.ImportSpec)(nil)}, func(node ast.Node) {
		spec := node.(*ast.ImportSpec)
		path, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			return
		}

		if isFrameworkImport(path) {
			pass.Reportf(spec.Pos(), "third-party HTTP framework %q; use the net/http standard library", path)
		}
	})
	return nil, nil
}

func isFrameworkImport(path string) bool {
	for _, prefix := range frameworkPrefixes {
		if path == prefix || strings.HasPrefix(path, prefix+"/") {
			return true
		}
	}
	return false
}
