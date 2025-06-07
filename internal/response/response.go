package response

import (
	"fmt"
	"io"
)

const (
	crlf = "\r\n"
)

type StatusCode int

const (
	StatusCodeSuccess             StatusCode = 200
	StatusCodeBadRequest          StatusCode = 400
	StatusCodeInternalServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	_, err := w.Write(getStatusLine(statusCode))
	return err
}

func getStatusLine(statusCode StatusCode) []byte {
	responsePhrase := ""
	switch statusCode {
	case StatusCodeSuccess:
		responsePhrase += "OK"
	case StatusCodeBadRequest:
		responsePhrase += "Bad Request"
	case StatusCodeInternalServerError:
		responsePhrase += "Internal Server Error"
	}

	return []byte(fmt.Sprintf("HTTP/1.1 %d %s%s", statusCode, responsePhrase, crlf))
}
