package testnaming

import "testing"

func TestParseDocument(t *testing.T) {}

func TestStore_Get(t *testing.T) {}

func TestParseDocumentEdgeCases(t *testing.T) {} // want "names no exported function in this package"

func TestStore_Get_Errors(t *testing.T) {} // want "carries more than one underscore"

func TestParseInternal(t *testing.T) {} // want "names no exported function in this package"

func TestStore_Missing(t *testing.T) {} // want "names no exported method in this package"

func helper(t *testing.T) {}
