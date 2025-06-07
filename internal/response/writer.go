package response

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
)

type WriterState int

const (
	WriteToStatusLine WriterState = iota
	WriteToHeaders
	WriteToBody
	WriteFinished
)

type Writer struct {
	WriterState WriterState
	writer      io.Writer
}

type StatusLine struct {
	HttpVersion  string
	Code         StatusCode
	ReasonPhrase string
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		WriterState: WriteToStatusLine,
		writer:      w,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.WriterState != WriteToStatusLine {
		return fmt.Errorf("ReponseWriter not set to write to statusline > %v", w.WriterState)
	}
	defer func() { w.WriterState = WriteToHeaders }()
	_, err := w.writer.Write(getStatusLine(statusCode))
	return err
}

func (w *Writer) WriteHeaders(h headers.Headers) error {
	if w.WriterState != WriteToHeaders {
		return fmt.Errorf("ReponseWriter not set to write to Headers > %v", w.WriterState)
	}
	defer func() { w.WriterState = WriteToBody }()
	err := WriteHeaders(w.writer, h)
	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.WriterState != WriteToBody {
		return 0, fmt.Errorf("ReponseWriter not ready to write to body > %v", w.WriterState)
	}
	defer func() { w.WriterState = WriteFinished }()

	return w.writer.Write(p)

}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.WriterState != WriteToBody {
		return 0, fmt.Errorf("ResponseWriter not ready to write to body > %v", w.WriterState)
	}

	chunkSize := len(p)
	nTotal := 0

	n, err := fmt.Fprintf(w.writer, "%x\r\n", chunkSize)
	if err != nil {
		return nTotal, err
	}
	nTotal += n

	n, err = w.writer.Write(p)
	if err != nil {
		return nTotal, err
	}
	nTotal += n

	n, err = w.writer.Write([]byte("\r\n"))
	if err != nil {
		return nTotal, err
	}
	nTotal += n

	return nTotal, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.WriterState != WriteToBody {
		return 0, fmt.Errorf("ResponseWriter not ready to write to body > %v", w.WriterState)
	}
	defer func() { w.WriterState = WriteFinished }()
	chunkedEnd := []byte("0\r\n\r\n")
	return w.writer.Write(chunkedEnd)
}
