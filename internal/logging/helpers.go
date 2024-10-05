package logging

import (
	"bytes"
	"io"
	"net/http"
)

func ReadRequestBodyWithoutClosing(r *http.Request) []byte {
	var bodyBuffer bytes.Buffer
	io.Copy(&bodyBuffer, r.Body)
	r.Body.Close()
	r.Body = io.NopCloser(io.LimitReader(bytes.NewReader(bodyBuffer.Bytes()), 1024))
	return bodyBuffer.Bytes()
}
