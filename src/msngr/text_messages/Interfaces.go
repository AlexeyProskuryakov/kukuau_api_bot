package text_messages


type TextMessageSupplier interface {
	GenerateMessage() string
}