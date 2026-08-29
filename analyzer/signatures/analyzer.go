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
	// AbsenceSpelling errors on a pointer to a primitive, a pointer to a
	// collection, or a struct pointer returned alongside an ok.
	AbsenceSpelling = &analysis.Analyzer{
		Name:     "absence_spelling",
		Doc:      "error on a pointer to a primitive or collection, or a struct pointer returned alongside an ok",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runAbsenceSpelling,
	}

	// RecordValueWithBool warns when a record type is returned by value alongside an ok.
	RecordValueWithBool = &analysis.Analyzer{
		Name:     "record_value_with_bool",
		Doc:      "warn when a struct with identity is returned by value alongside an ok from what looks like a keyed lookup; records spell absence with a pointer",
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

// runAbsenceSpelling reports two things this analyzer can actually justify:
// a pointer to something that was never a record in the first place (a
// primitive, or a collection, which already spells absence as nil), and a
// struct pointer returned alongside an ok — nil already says "absent," so
// the ok is redundant when they agree and undefined when they do not. A
// struct returned by value alongside an ok, the comma-ok idiom, is not
// flagged here; see [RecordValueWithBool] for when that shape earns a
// pointer instead.
func runAbsenceSpelling(pass *analysis.Pass) (interface{}, error) {
	forEachSignature(pass, func(fn *ast.FuncDecl, results []types.Type) {
		var hasBool bool
		var structPointers []types.Type
		for _, result := range results {
			switch {
			case isBool(result):
				hasBool = true
			case isPointerToCollection(result):
				pass.Reportf(fn.Pos(), "a nil slice is an empty slice; return the collection itself, not a pointer to it")
			case isPointerToStruct(result):
				structPointers = append(structPointers, result)
			case isPointer(result):
				pass.Reportf(fn.Pos(), "%s is a pointer to a primitive; return the value, or (value, bool) if the zero value is ambiguous with absence", result.String())
			}
		}
		if !hasBool {
			return
		}
		for _, structPointer := range structPointers {
			pass.Reportf(fn.Pos(), "%s is a pointer alongside an ok; nil already spells absence, so the two agree redundantly or, when they disagree, undefined — return one or the other", structPointer.String())
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
		if !hasBool || len(records) == 0 {
			return
		}

		// A lookup keyed by a single identifying parameter is the pattern this
		// rule is after: a store-style accessor whose "not found" case should
		// be an unambiguous nil, not a same-shaped zero value indistinguishable
		// from a legitimately empty record. A function assembling a record from
		// several already-known inputs — a decoder, a parser — has no identity
		// to look up; its ok is the same comma-ok idiom as a map read, and
		// forcing a pointer there only adds an allocation for no benefit.
		if nonContextParamCount(pass, fn.Type.Params) > 1 {
			return
		}

		for _, record := range records {
			pass.Reportf(fn.Pos(), "%s is a record; return *%s and spell absence with nil, or add it to the value types if it has value semantics", record.String(), record.String())
		}
	})
	return nil, nil
}

// nonContextParamCount counts a signature's parameters, excluding
// context.Context: it carries no identity, so it should not count toward
// whether a call looks like a keyed lookup.
func nonContextParamCount(pass *analysis.Pass, params *ast.FieldList) int {
	if params == nil {
		return 0
	}
	count := 0
	for _, field := range params.List {
		if t := pass.TypesInfo.TypeOf(field.Type); t != nil && t.String() == "context.Context" {
			continue
		}
		n := len(field.Names)
		if n == 0 {
			n = 1
		}
		count += n
	}
	return count
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

// isPointerToStruct reports whether target is a pointer whose element is a
// struct — the shape that already has a record's identity, as opposed to a
// pointer to a primitive or a collection.
func isPointerToStruct(target types.Type) bool {
	pointer, ok := target.Underlying().(*types.Pointer)
	if !ok {
		return false
	}
	_, ok = pointer.Elem().Underlying().(*types.Struct)
	return ok
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
