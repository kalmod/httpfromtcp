package server

import (
	"fmt"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"io"
)

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

type Handler func(w *response.Writer, req *request.Request)

// type Handler func(w io.Writer, req *request.Request) *HandlerError

func (h *HandlerError) GetStatusLine() []byte {
	return []byte(fmt.Sprintf("HTTP/1.1 %d %s\r\n", h.StatusCode, ""))

}

func WriteStatusLine_HandlerError(w io.Writer, handlerErr *HandlerError) error {
	_, err := w.Write(handlerErr.GetStatusLine())
	return err
}

func (he HandlerError) Write(w io.Writer) {
	response.WriteStatusLine(w, he.StatusCode)
	messageBytes := []byte(he.Message)
	headers := response.GetDefaultHeaders(len(messageBytes))
	response.WriteHeaders(w, headers)
	w.Write(messageBytes)
}
