package boolparam

func DoSomething(value string, enabled bool) {} // want "boolean parameters indicate"

func DoSomethingGood(value string) {}

func DoWithOptions(value string, opts map[string]interface{}) {}

type Handler struct{}

func (h Handler) Handle(ctx interface{}, force bool) {} // want "boolean parameters indicate"

func (h Handler) Process(ctx interface{}) {}

// MultipleBools has two bool parameters, should report each
func MultipleBools(a bool, b bool) {} // want "boolean parameters indicate" "boolean parameters indicate"
