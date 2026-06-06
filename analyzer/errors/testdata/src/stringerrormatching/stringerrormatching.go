package stringerrormatching

import (
	"errors"
	"strings"
)

func BadContains(err error) bool {
	return strings.Contains(err.Error(), "timeout") // want "string matching on error message; use errors.Is or errors.As"
}

func BadHasPrefix(err error) bool {
	return strings.HasPrefix(err.Error(), "net") // want "string matching on error message; use errors.Is or errors.As"
}

func BadHasSuffix(err error) bool {
	return strings.HasSuffix(err.Error(), "closed") // want "string matching on error message; use errors.Is or errors.As"
}

func BadEqualFold(err error) bool {
	return strings.EqualFold(err.Error(), "ERROR") // want "string matching on error message; use errors.Is or errors.As"
}

func GoodUsesIs(err error) bool {
	return errors.Is(err, errors.New("timeout"))
}

func GoodUsesAs(err error) bool {
	var e interface{ Error() string }
	return errors.As(err, &e)
}

func GoodStringOps(s string) bool {
	return strings.Contains(s, "text")
}

func GoodMultipleArgs(err error, s string) bool {
	return strings.Contains(s, "text") // good - not calling err.Error()
}
