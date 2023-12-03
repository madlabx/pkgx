package httpx

type DropWriter struct {
}

func (dw *DropWriter) Write(p []byte) (int, error) {
	return len(p), nil
}
