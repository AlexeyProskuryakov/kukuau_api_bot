package msngr

import (
	"msngr/db"
	s "msngr/structs"
	"errors"
)

var (
	DB_ERROR = errors.New("Ошибка на стороне базы данных, попробуйте позже.")
	DB_ERROR_RESULT = &s.MessageResult{Type:"chat", Body:DB_ERROR.Error()}

	MESSAGE_DATA_ERROR = errors.New("Ошибка в том, что нет данных в вашем сообщении. Попробуйте как-нибудь по-другому, пожалуйста. :(")
	MESSAGE_DATA_ERROR_RESULT = &s.MessageResult{Type:"chat", Body:MESSAGE_DATA_ERROR.Error()}

	GLOBAL_ERROR = errors.New("Ничего не понятно :(")
	GLOBAL_ERROR_RESULT = &s.MessageResult{Type:"chat", Body:GLOBAL_ERROR.Error()}
)

type Func func(in *s.InPkg) (*[]s.OutCommand, error)

type FuncTextBodyProcessor struct {
	Storage *db.MainDb
	F Func
	MessageRecipientIdentity string
	AnswerText *string
}


func (ftbp *FuncTextBodyProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	commands, err := ftbp.F(in)
	if err != nil{
		return s.ErrorMessageResult(err, commands)
	} else {
		if in.UserData != nil{
			err := ftbp.Storage.Users.StoreUserData(in.From, in.UserData)
			if err != nil{
				return DB_ERROR_RESULT
			}
		}
		if in.Message != nil && in.Message.Body != nil{
			mesageBody := in.Message.Body
			ftbp.Storage.Messages.StoreMessage(in.From, ftbp.MessageRecipientIdentity, *mesageBody, in.Message.ID)
			if ftbp.AnswerText != nil {
				answer := ftbp.AnswerText
				return &s.MessageResult{Body:*answer, Type:"chat", IsDeferred:false, Commands:commands}
			} else {
				return &s.MessageResult{IsDeferred:true, Commands:commands}
			}
		} else {
			return MESSAGE_DATA_ERROR_RESULT
		}
	}
	return GLOBAL_ERROR_RESULT
}


func NewSimpleTextBodyProcessor(storage *db.MainDb, commands *[]s.OutCommand, recipientId string, answerText *string) *FuncTextBodyProcessor{
	f := func (in *s.InPkg) (*[]s.OutCommand, error){
		return commands, nil
	}
	result := &FuncTextBodyProcessor{Storage:storage, F:f, MessageRecipientIdentity:recipientId, AnswerText:answerText}
	return result
}


type InformationProcessor struct {
	Information string
}

func (ip *InformationProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	result := s.MessageResult{Type:"chat", Body:ip.Information}
	return &result
}

func NewInformationProcessor(information string) *InformationProcessor{
	return &InformationProcessor{Information:information}
}
