package msngr

type MessageResult struct {
	Commands *[]OutCommand
	Body string
	Error error
	IsDeferred bool
}

type RequestResult struct {
	Commands *[]OutCommand
	Error error
}

type RequestCommandProcessor interface {
	ProcessRequest(in InPkg) RequestResult
}

type MessageCommandProcessor interface {
	ProcessMessage(in InPkg) MessageResult
}
