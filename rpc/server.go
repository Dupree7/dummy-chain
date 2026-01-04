package rpc

import (
	"dummy-chain/storage"
	"io"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
)

type Server struct {
	server *rpc.Server
}

func NewServer(storage *storage.BadgerDb, memPool *MemoryPool) (*Server, error) {
	newServer := rpc.NewServer()
	newService := NewService(storage, memPool)
	if errRegister := newServer.RegisterName("chain", newService); errRegister != nil {
		return nil, errRegister
	}
	return &Server{
		server: newServer,
	}, nil
}

func (s *Server) Start() error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(405)
			return
		}

		conn := &HttpConn{in: r.Body, out: w}
		codec := jsonrpc.NewServerCodec(conn)

		s.server.ServeCodec(codec)
	})
	return http.ListenAndServe(":12345", nil)
}

type HttpConn struct {
	in  io.Reader
	out io.Writer
}

func (c *HttpConn) Read(p []byte) (n int, err error)  { return c.in.Read(p) }
func (c *HttpConn) Write(p []byte) (n int, err error) { return c.out.Write(p) }
func (c *HttpConn) Close() error                      { return nil }
