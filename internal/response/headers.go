package response

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
)

func GetDefaultHeaders(contentLen int) headers.Headers {
	defaultHeader := headers.NewHeaders()
	defaultHeader.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	defaultHeader.Set("Connection", "close")
	defaultHeader.Set("Content-Type", "text/plain")
	return defaultHeader
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for key, val := range headers {
		_, err := w.Write([]byte(fmt.Sprintf("%s: %s%s", key, val, crlf)))
		if err != nil {
			return err
		}
	}
	_, err := w.Write([]byte(crlf))
	return err
}
