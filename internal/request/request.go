package request

import (
	"bytes"
	"errors"
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"strings"
)

const (
	crlf       = "\r\n"
	bufferSize = 8
)

type ParserState int

const (
	requestState_initialized ParserState = iota
	requestState_parsingHeaders
	requestState_done
)

type Request struct {
	RequestLine RequestLine
	State       ParserState
	Headers     headers.Headers
}

// GET /coffee HTTP/1.1
type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	// rawbytes, err := io.ReadAll(reader)
	input_buffer := make([]byte, bufferSize)
	readToIndex := 0
	request := &Request{State: requestState_initialized, Headers: headers.NewHeaders()}
	for request.State != requestState_done { // I keep adding data into my buffer
		if readToIndex >= len(input_buffer) {
			tmp := make([]byte, len(input_buffer)*2)
			copy(tmp, input_buffer)
			input_buffer = tmp
		}
		numBytesRead, err := reader.Read(input_buffer[readToIndex:])
		if err != nil {
			if errors.Is(io.EOF, err) {
				request.State = requestState_done
				break
			}
			return nil, err

		}
		readToIndex += numBytesRead

		// in the parser -> parseRequestLine, once I get to crlf I ingest the data.
		numBytesParsed, err := request.parse(input_buffer[:readToIndex])
		if err != nil {
			return nil, err
		}
		// Now that I finally ingested the data and parsed it, I purge it.
		// I just copy starting from what has been read to the end of the slice.
		// remove that value from my starting index
		copy(input_buffer, input_buffer[numBytesParsed:])
		readToIndex -= numBytesParsed

	}

	return request, nil
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		// return nil, 0, fmt.Errorf("could not find CRLF in request-line")
		return nil, 0, nil // no \r\n found yet
	}
	requestLineText := string(data[:idx])
	requestLine, err := requestLineFromString(requestLineText)
	if err != nil {
		return nil, 0, err
	}

	// +2 to skip the \r\n
	return requestLine, idx + 2, nil
}

func requestLineFromString(str string) (*RequestLine, error) {
	parts := strings.Split(str, " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("poorly formatted request-line: %s", str)
	}
	method := parts[0]
	for _, c := range method {
		if c < 'A' || c > 'Z' {
			return nil, fmt.Errorf("invalid method: %s", method)
		}
	}

	requestTarget := parts[1]

	versionParts := strings.Split(parts[2], "/")
	if len(versionParts) != 2 {
		return nil, fmt.Errorf("malformed start-line: %s", str)
	}

	httpPart := versionParts[0]
	if httpPart != "HTTP" {
		return nil, fmt.Errorf("unrecognized HTTP-version: %s", str)
	}
	version := versionParts[1]
	if version != "1.1" {
		return nil, fmt.Errorf("unrecognized HTTP-version: %s", str)
	}

	return &RequestLine{
		Method: method, RequestTarget: requestTarget, HttpVersion: versionParts[1],
	}, nil
}

// accepts next slice of data to go into our request struct
func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.State != requestState_done {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}
		totalBytesParsed += n
		if n == 0 {
			break
		}
	}
	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.State {
	case requestState_initialized:
		requestLine, bytesConsumed, err := parseRequestLine(data)
		if err != nil {
			return 0, fmt.Errorf("could not parse request: %s", err)
		}
		if bytesConsumed == 0 {
			// more data needed
			return 0, nil
		}
		r.RequestLine = *requestLine
		r.State = requestState_parsingHeaders
		return bytesConsumed, nil
	case requestState_parsingHeaders:
		bytesConsumed, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, fmt.Errorf("could not parse headers: %s", err)
		}
		if bytesConsumed == 0 {
			return 0, nil
		}
		if done {
			r.State = requestState_done
		}
		return bytesConsumed, nil
	case requestState_done:
		return 0, fmt.Errorf("error: trying to read data in a done state")
	default:
		return 0, fmt.Errorf("error: unkown state")
	}
}

func (r *Request) Print() {

	fmt.Printf("Request line: \n- Method: %s\n- Target: %s\n- Version: %s\n",
		r.RequestLine.Method,
		r.RequestLine.RequestTarget,
		r.RequestLine.HttpVersion)

	fmt.Println("Headers:")
	for key, val := range r.Headers {
		fmt.Printf(" - %s: %s\n", key, val)
	}
}
