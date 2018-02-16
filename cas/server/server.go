package main

import (
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"sync"

	"github.com/andyleap/casfs/cas/proto"
)

type Server struct {
	l         net.Listener
	baseDir   string
	garbageMu sync.Mutex
	garbage   map[string]struct{}
}

func New(addr string, baseDir string) (*Server, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &Server{
		l:       l,
		baseDir: baseDir,
		garbage: map[string]struct{}{},
	}, nil
}

func (s *Server) Serve() error {
	for {
		conn, err := s.l.Accept()
		if err != nil {
			return err
		}
		go s.Handle(conn)
	}
}

func (s *Server) Handle(conn net.Conn) {
	for {
		var req proto.Req
		var resp proto.Resp
		err := req.Deserialize(conn)
		if err != nil {
			return
		}
		resp.ID = req.ID
		switch r := req.Request.(type) {
		case proto.ReadReq:
			hexHash := hex.EncodeToString(r.Hash)
			dataPath := filepath.Join(s.baseDir, hexHash[:2], hexHash[2:4], hexHash[4:])
			s.garbageMu.Lock()
			delete(s.garbage, dataPath)
			s.garbageMu.Unlock()
			data, err := ioutil.ReadFile(dataPath)
			errStr := ""
			if err != nil {
				errStr = err.Error()
			}
			resp.Response = proto.ReadResp{
				Data:  data,
				Error: errStr,
			}
		case proto.WriteReq:
			hash := sha256.Sum256(r.Data)
			hexHash := hex.EncodeToString(hash[:])
			dataPath := filepath.Join(s.baseDir, hexHash[:2], hexHash[2:4], hexHash[4:])
			s.garbageMu.Lock()
			delete(s.garbage, dataPath)
			s.garbageMu.Unlock()
			os.MkdirAll(filepath.Join(s.baseDir, hexHash[:2], hexHash[2:4]), 0777)
			err := ioutil.WriteFile(dataPath, r.Data, 0666)
			if err != nil {
				resp.Response = proto.WriteResp{
					Error: err.Error(),
				}
			} else {
				resp.Response = proto.WriteResp{
					Hash: hash[:],
				}
			}
		case proto.MarkReq:
			s.garbageMu.Lock()
			filepath.Walk(s.baseDir, func(path string, info os.FileInfo, err error) error {
				s.garbage[path] = struct{}{}
				return nil
			})
			s.garbageMu.Unlock()
			resp.Response = proto.MarkResp{}
		case proto.SweepReq:
			s.garbageMu.Lock()
			for k := range s.garbage {
				os.Remove(k)
				delete(s.garbage, k)
			}
			s.garbageMu.Unlock()
			resp.Response = proto.SweepResp{}
		}
		err = resp.Serialize(conn)
		if err != nil {
			return
		}
	}
}
