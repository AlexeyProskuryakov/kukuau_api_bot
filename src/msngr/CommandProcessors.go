package msngr

type RequestCommandProcessor interface {
	ProcessRequest(in InPkg) ([]OutCommand, error)
}

type MessageCommandProcessor interface {
	ProcessMessage(in InPkg) (string, *[]OutCommand, error)
}
