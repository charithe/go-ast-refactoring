package example

type Wibbler interface {
	Wibble(i int) int
}

type WibbleClient struct{}

func (wc WibbleClient) Wibble(i int) int {
	return i * 2
}

type WibbleClientWrapper struct {
	WibbleClient
}

func (wbc WibbleClientWrapper) Wobble(j int) int {
	return j * 5
}
