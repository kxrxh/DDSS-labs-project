package server

import (
	"encoding/json"
	"log"

	"github.com/kxrxh/ddss/internal/db"
	"github.com/panjf2000/gnet/v2"
)

type Server struct {
	gnet.BuiltinEventEngine
	eng    gnet.Engine
	dgraph *db.DgraphClient
	addr   string
}

func NewServer(addr string, dgraph *db.DgraphClient) *Server {
	return &Server{
		addr:   addr,
		dgraph: dgraph,
	}
}

func (s *Server) OnBoot(eng gnet.Engine) gnet.Action {
	s.eng = eng
	log.Printf("Server is listening on %s\n", s.addr)
	return gnet.None
}

func (s *Server) OnTraffic(c gnet.Conn) gnet.Action {
	buf, _ := c.Next(-1)
	if len(buf) == 0 {
		return gnet.None
	}

	var req Request
	if err := json.Unmarshal(buf, &req); err != nil {
		response := Response{
			Success: false,
			Error:   "Invalid request format",
		}
		s.sendResponse(c, response)
		return gnet.None
	}

	response := s.handleRequest(req)
	s.sendResponse(c, response)
	return gnet.None
}

func (s *Server) sendResponse(c gnet.Conn, response Response) {
	data, _ := json.Marshal(response)
	data = append(data, '\n')
	c.AsyncWrite(data, nil)
}

func (s *Server) GetAddress() string {
	return s.addr
}
