package boolparam

func DoSomething(value string, enabled bool) {} // want "boolean parameters indicate"

func DoSomethingGood(value string) {}

func DoWithOptions(value string, opts map[string]interface{}) {}
