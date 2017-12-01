package gnode

type msgSysProcessStart struct{}
type msgSysProcessStop struct{}

type msgSysCallRequest struct {
	callID     uint32
	from       uint32
	to         uint32
	methodName string
	request    interface{}
}

type msgSysCallResponse struct {
	callID   uint32
	from     uint32
	to       uint32
	response interface{}
	err      error
}
