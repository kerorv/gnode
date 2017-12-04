package gnode

type MsgProcessStart struct{}
type MsgProcessStop struct{}
type MsgProcessCoroutinePanic struct {
	Err error
}

type msgProcessCallRequest struct {
	callID     uint32
	from       uint32
	to         uint32
	methodName string
	request    []interface{}
}

type msgProcessCallResponse struct {
	callID   uint32
	from     uint32
	to       uint32
	response []interface{}
	err      error
}
