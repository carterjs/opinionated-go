package fireandforget

func task() {}

func example() {
	go task() // want "fire-and-forget goroutine"

	ch := make(chan struct{})
	go func() {
		<-ch
	}()
	close(ch)
}
