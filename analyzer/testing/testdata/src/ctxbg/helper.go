package ctxbg

import "context"

// Run does work. context.Background here is not this analyzer's business.
func Run() {
	ctx := context.Background()
	_ = ctx
}
