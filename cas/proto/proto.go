package proto

//go:generate gencode go -package proto -schema proto.schema

type Request interface {
	request()
}

type Response interface {
	response()
}

func (ReadReq) request()  {}
func (WriteReq) request() {}
func (MarkReq) request() {}
func (SweepReq) request() {}

func (ReadResp) response()  {}
func (WriteResp) response() {}
func (MarkResp) response() {}
func (SweepResp) response() {}
