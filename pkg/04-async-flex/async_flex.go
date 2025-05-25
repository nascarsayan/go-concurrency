// https://go.dev/play/p/FxSsaTToIHe
package async_flex

type In struct{}
type Out struct{}

func Asynchronous(x In) <-chan Out {
	ch := make(chan Out)
	go func() {
		ch <- Synchronous(x)
	}()
	return ch
}

func Synchronous(x In) Out {
	c := Asynchronous(x)
	return <-c
}
