package server

import (
	"fmt"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"log"
	"net"
	"sync/atomic"
)

type ServerState int

const (
	serverState_Listening ServerState = iota
	serverState_Closed
	serverState_Error
)

type Server struct {
	handler  Handler
	listener net.Listener
	closed   atomic.Bool
}

// Creates a net.Listener and returns a new Server isntance.
// Listener runs on a go routine
func Serve(port int, handlerFunc Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	srv := &Server{
		handler:  handlerFunc,
		listener: listener,
	}
	go srv.listen()
	return srv, nil
}

// Closes the listener and the server.
func (s *Server) Close() error {
	s.closed.Store(true)
	if s.listener != nil {
		s.listener.Close()
	}
	return nil
}

// Loops to Accept new connections as they come in
// Handles each new request in a go routine.
// atomic.Bool is used to track if a server is closed.
func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			log.Printf("Server::listen::error > %v", err.Error())
			return
		}
		go s.handle(conn)
	}
}

// Handles a single connection by writing the following response and closing the connection.
// func (s *Server) handle(conn net.Conn) {
// 	defer conn.Close()
// 	req, err := request.RequestFromReader(conn)
// 	if err != nil {
// 		hErr := &HandlerError{
// 			StatusCode: response.StatusCodeBadRequest,
// 			Message:    err.Error(),
// 		}
// 		hErr.Write(conn)
// 		return
// 	}
// 	buf := bytes.NewBuffer([]byte{})
// 	hErr := s.handler(buf, req)
// 	if hErr != nil {
// 		hErr.Write(conn)
// 		return
// 	}
//
// 	b := buf.Bytes()
// 	response.WriteStatusLine(conn, response.StatusCodeSuccess)
// 	headers := response.GetDefaultHeaders(len(b))
// 	response.WriteHeaders(conn, headers)
// 	conn.Write(b)
// 	return
// }
//

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	req, err := request.RequestFromReader(conn)
	w := response.NewWriter(conn)
	if err != nil {
		w.WriteStatusLine(response.StatusCodeInternalServerError)
		w.WriteHeaders(response.GetDefaultHeaders(0))
		w.WriteBody([]byte(fmt.Sprintf("Error parsing request: %v", err)))
		return
	}
	s.handler(w, req)
	return
}
