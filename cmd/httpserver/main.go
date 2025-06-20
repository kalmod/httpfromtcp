package main

import (
	"crypto/sha256"
	"fmt"
	"httpfromtcp/internal/headers"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

const port = 42069

func main() {
	server, err := server.Serve(port, test_handler)
	if err != nil {
		log.Fatalf("error starting server: %v", err)
	}

	defer server.Close()

	log.Println("Server started on port: ", port)

	// make a signal that waits until a syscal signal is sent to the channel.
	// like ctrl+c
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped.")
}

// func test_handler01(w io.Writer, req *request.Request) *server.HandlerError {
// 	he := &server.HandlerError{}
// 	switch req.RequestLine.RequestTarget {
// 	case "/yourproblem":
// 		he.StatusCode = response.StatusCodeBadRequest
// 		he.Message = "Your problem is not my problem"
// 		return he
// 	case "/myproblem":
// 		he.StatusCode = response.StatusCodeInternalServerError
// 		he.Message = "Woopsie, my bad"
// 		return he
// 	}
// 	w.Write([]byte("All good, frfr\n"))
// 	return nil
// }

func test_handler(w *response.Writer, req *request.Request) {
	if req.RequestLine.RequestTarget == "/yourproblem" {
		handler400(w, req)
		return
	}
	if req.RequestLine.RequestTarget == "/myproblem" {
		handler500(w, req)
		return
	}
	if req.RequestLine.RequestTarget == "/video" {
		handlerVideo(w, req)
		return
	}
	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
		handlerHTTPBIN(w, req)
		return
	}
	handler200(w, req)
	return
}

func handlerVideo(w *response.Writer, req *request.Request) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(err)
		handler500(w, req)
		return
	}

	filePath := filepath.Join(homeDir, "Code", "bootdev", "httpfromtcp", "assets", "vim.mp4")

	videoFile, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println(err)
		handler500(w, req)
		return
	}
	w.WriteStatusLine(response.StatusCodeSuccess)
	h := response.GetDefaultHeaders(len(videoFile))
	h.Override("Content-Type", "video/mp4")
	w.WriteHeaders(h)
	w.WriteBody(videoFile)
	return
}

func handlerHTTPBIN(w *response.Writer, req *request.Request) {
	target := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")
	routed_target := "https://httpbin.org/" + target

	fmt.Println("Proxying to >> ", routed_target)
	http_resp, resp_err := http.Get(routed_target)
	if resp_err != nil {
		handler500(w, req)
		return
	}
	defer http_resp.Body.Close()

	w.WriteStatusLine(response.StatusCodeSuccess)
	h := response.GetDefaultHeaders(0)
	h.Override("Transfer-Encoding", "chunked")
	h.Remove("Content-Length") // chaning length and setting encoding
	h.Set("Trailer", "X-Content-SHA256")
	h.Set("Trailer", "X-Content-Length")

	w.WriteHeaders(h)

	const maxChunkSize = 1024
	buf := make([]byte, maxChunkSize)

	fmt.Println("Starting to read from: ", routed_target)
	totalResponseBody := []byte{}
	for {
		n, err := http_resp.Body.Read(buf)
		fmt.Println("\t", n, "- bytes, read from httpbin")
		if n > 0 {
			totalResponseBody = append(totalResponseBody, buf[:n]...)
			_, err = w.WriteChunkedBody(buf)
			if err != nil {
				fmt.Println("WriteChunkedBody ERROR: ", err)
				break
			}

		}
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("Body Read ERROR: ", err)
			break
		}

	}
	_, err := w.WriteChunkedBodyDone()
	if err != nil {
		fmt.Println("Error writing chunked body done:", err)
	}
	trailers := headers.NewHeaders()
	shaSum := sha256.Sum256(totalResponseBody)
	trailers.Override("X-Content-SHA256", fmt.Sprintf("%x", shaSum[:]))
	trailers.Override("X-Content-Length", fmt.Sprintf("%d", len(totalResponseBody)))
	err = w.WriteTrailers(trailers)
	if err != nil {
		fmt.Println("\tError writing trailers: ", err)
	}
	fmt.Println("Finished reading from: ", routed_target)
	return

}

func handler400(w *response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.StatusCodeBadRequest)
	body := []byte(`<html>
<head>
<title>400 Bad Request</title>
</head>
<body>
<h1>Bad Request</h1>
<p>Your request honestly kinda sucked.</p>
</body>
</html>`)
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
	return
}

func handler500(w *response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.StatusCodeInternalServerError)
	body := []byte(`<html>
<head>
<title>500 Internal Server Error</title>
</head>
<body>
<h1>Internal Server Error</h1>
<p>Okay, you know what? This one is on me.</p>
</body>
</html>`)
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
	return
}

func handler200(w *response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.StatusCodeSuccess)
	body := []byte(`<html>
<head>
<title>200 OK</title>
</head>
<body>
<h1>Success!</h1>
<p>Your request was an absolute banger.</p>
</body>
</html>`)
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
	return
}
