package msngr

type RequestCommandProcessor interface {
	ProcessRequest(in InPkg) ([]Command, error)
}

type MessageCommandProcessor interface {
	ProcessMessage(in InPkg) (string, *[]Command, error)
}
