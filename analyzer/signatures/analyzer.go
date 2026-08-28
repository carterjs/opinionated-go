// Package signatures reports how a signature spells absence and groups its parameters.
package signatures

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// valueStructs are struct types that behave as values, so they spell absence with a bool rather than a pointer.
var valueStructs = map[string]bool{
	"time.Time":                   true,
	"net/netip.Addr":              true,
	"net/netip.AddrPort":          true,
	"net/netip.Prefix":            true,
	"github.com/google/uuid.UUID": true,
}

var (
	// AbsenceSpelling errors when a signature spells absence twice, or spells a collection as absent.
	AbsenceSpelling = &analysis.Analyzer{
		Name:     "absence_spelling",
		Doc:      "error on a signature that returns both an ok and an error, both a pointer and an ok, or a collection that can be absent",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runAbsenceSpelling,
	}

	// RecordValueWithBool warns when a record type is returned by value alongside an ok.
	RecordValueWithBool = &analysis.Analyzer{
		Name:     "record_value_with_bool",
		Doc:      "warn when a struct with identity is returned by value alongside an ok; records spell absence with a pointer",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runRecordValueWithBool,
	}

	// AdjacentSameTypeParameters warns when an exported signature has two adjacent parameters of the same type.
	AdjacentSameTypeParameters = &analysis.Analyzer{
		Name:     "adjacent_same_type_parameters",
		Doc:      "warn on adjacent same-typed parameters in an exported signature; callers will swap them",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runAdjacentSameTypeParameters,
	}
)

func runAbsenceSpelling(pass *analysis.Pass) (interface{}, error) {
	forEachSignature(pass, func(fn *ast.FuncDecl, results []types.Type) {
		var values, pointers, collections int
		var hasBool, hasError bool
		for _, result := range results {
			switch {
			case isError(result):
				hasError = true
			case isBool(result):
				hasBool = true
			case isPointerToCollection(result):
				pass.Reportf(fn.Pos(), "a nil slice is an empty slice; return the collection itself, not a pointer to it")
				pointers++
				values++
			case isPointer(result):
				pointers++
				values++
			case isCollection(result):
				collections++
				values++
			default:
				values++
			}
		}

		// A lone (bool, error) answers with its bool, as regexp.MatchString does; only a bool beside a value is an ok.
		if hasBool && hasError && values > 0 {
			pass.Reportf(fn.Pos(), "a signature returns ok or error, never both; a legitimate miss returns ok, a miss the caller cannot pass returns a sentinel error")
		}
		if hasBool && pointers > 0 {
			pass.Reportf(fn.Pos(), "a record spells absence with nil; returning both a pointer and an ok leaves (nil, true) undefined")
		}
		if hasBool && collections > 0 {
			pass.Reportf(fn.Pos(), "a collection has no absent state; a nil slice and an empty slice are the same to every caller")
		}
	})
	return nil, nil
}

func runRecordValueWithBool(pass *analysis.Pass) (interface{}, error) {
	forEachSignature(pass, func(fn *ast.FuncDecl, results []types.Type) {
		var hasBool bool
		var records []types.Type
		for _, result := range results {
			switch {
			case isBool(result):
				hasBool = true
			case isRecordValue(result):
				records = append(records, result)
			}
		}
		if !hasBool {
			return
		}
		for _, record := range records {
			pass.Reportf(fn.Pos(), "%s is a record; return *%s and spell absence with nil, or add it to the value types if it has value semantics", record.String(), record.String())
		}
	})
	return nil, nil
}

func runAdjacentSameTypeParameters(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspect.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(node ast.Node) {
		fn := node.(*ast.FuncDecl)
		if !isExported(fn.Name.Name) || fn.Type.Params == nil {
			return
		}

		var previous types.Type
		for _, field := range fn.Type.Params.List {
			current := pass.TypesInfo.TypeOf(field.Type)
			if current == nil {
				previous = nil
				continue
			}
			// A single field naming several parameters is already adjacent: `src, dst string`.
			if len(field.Names) > 1 || (previous != nil && types.Identical(previous, current)) {
				pass.Reportf(field.Pos(), "adjacent parameters of type %s are swappable at the call site; group them into a request or options struct", current.String())
			}
			previous = current
		}
	})
	return nil, nil
}

// forEachSignature calls visit with the flattened result types of every function and method in the package.
func forEachSignature(pass *analysis.Pass, visit func(fn *ast.FuncDecl, results []types.Type)) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspect.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(node ast.Node) {
		fn := node.(*ast.FuncDecl)
		if fn.Type.Results == nil {
			return
		}

		var results []types.Type
		for _, field := range fn.Type.Results.List {
			resultType := pass.TypesInfo.TypeOf(field.Type)
			if resultType == nil {
				continue
			}
			count := len(field.Names)
			if count == 0 {
				count = 1
			}
			for range count {
				results = append(results, resultType)
			}
		}
		if len(results) > 1 {
			visit(fn, results)
		}
	})
}

func isBool(target types.Type) bool {
	basic, ok := target.(*types.Basic)
	return ok && basic.Kind() == types.Bool
}

func isError(target types.Type) bool {
	return target.String() == "error"
}

func isPointer(target types.Type) bool {
	_, ok := target.Underlying().(*types.Pointer)
	return ok
}

func isCollection(target types.Type) bool {
	switch target.Underlying().(type) {
	case *types.Slice, *types.Map:
		return true
	}
	return false
}

func isPointerToCollection(target types.Type) bool {
	pointer, ok := target.Underlying().(*types.Pointer)
	if !ok {
		return false
	}
	return isCollection(pointer.Elem())
}

// isRecordValue reports whether target is a named struct passed by value, which owes its callers a pointer instead.
func isRecordValue(target types.Type) bool {
	named, ok := target.(*types.Named)
	if !ok {
		return false
	}
	if _, ok := named.Underlying().(*types.Struct); !ok {
		return false
	}
	pkg := named.Obj().Pkg()
	if pkg == nil {
		return false
	}
	return !valueStructs[pkg.Path()+"."+named.Obj().Name()]
}

func isExported(name string) bool {
	return len(name) > 0 && name[0] >= 'A' && name[0] <= 'Z'
}
