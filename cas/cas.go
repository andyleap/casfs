package cas

import (
	"net"
	"sync"
	"errors"

	"github.com/andyleap/casfs/cas/proto"
)

var (
	ErrLostConn = errors.New("Connection Lost")
)

type Client struct {
	conn       net.Conn
	requests   map[uint64]chan proto.Response
	requestsMu sync.Mutex
	nextID     uint64
}

func Dial(addr string) (*Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	c := &Client{
		conn:     conn,
		requests: map[uint64]chan proto.Response{},
	}
	go c.run()
	return c, nil
}

func (c *Client) run() {
	for {
		var resp proto.Resp
		err := resp.Deserialize(c.conn)
		c.requestsMu.Lock()
		if err != nil {
			for _, v := range c.requests {
				close(v)
			}
			c.requests = nil
			c.requestsMu.Unlock()
			return
		}
		v, ok := c.requests[resp.ID]
		if ok {
			v <- resp.Response
		}
		c.requestsMu.Unlock()
	}
}

func (c *Client) call(req proto.Request) chan proto.Response {
	c.requestsMu.Lock()
	defer c.requestsMu.Unlock()
	if c.requests == nil {
		return nil
	}
	c.nextID++
	pReq := proto.Req{
		ID:      c.nextID,
		Request: req,
	}
	respC := make(chan proto.Response)
	c.requests[pReq.ID] = respC
	err := pReq.Serialize(c.conn)
	if err != nil {
		return nil
	}
	return respC
}

type ReadError string

func (re ReadError) Error() string {
	return string(re)
}

func (c *Client) Read(hash []byte) (data []byte, err error) {
	req := proto.ReadReq{
		Hash: hash,
	}
	resp := <-c.call(req)
	readResp, ok := resp.(proto.ReadResp)
	if !ok {
		return nil, ErrLostConn
	}
	if len(readResp.Error) != 0 {
		return nil, ReadError(readResp.Error)
	}
	return readResp.Data, nil
}

type WriteError string

func (we WriteError) Error() string {
	return string(we)
}

func (c *Client) Write(data []byte) (hash []byte, err error) {
	req := proto.WriteReq{
		Data: data,
	}
	resp := <-c.call(req)
	writeResp, ok := resp.(proto.WriteResp)
	if !ok {
		return nil, ErrLostConn
	}
	if len(writeResp.Error) != 0 {
		return nil, WriteError(writeResp.Error)
	}
	return writeResp.Hash, nil
}

func (c *Client) Mark() (err error) {
	req := proto.MarkReq{}
	resp := <-c.call(req)
	_, ok := resp.(proto.MarkResp)
	if !ok {
		return ErrLostConn
	}
	return nil
}

func (c *Client) Sweep() (err error) {
	req := proto.SweepReq{}
	resp := <-c.call(req)
	_, ok := resp.(proto.SweepResp)
	if !ok {
		return ErrLostConn
	}
	return nil
}
