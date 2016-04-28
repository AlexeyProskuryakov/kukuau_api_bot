package msngr

import (
	"msngr/db"
	s "msngr/structs"
	"errors"
	"log"
)

var (
	DB_ERROR = errors.New("Ошибка на стороне базы данных, попробуйте позже.")
	DB_ERROR_RESULT = &s.MessageResult{Type:"chat", Body:DB_ERROR.Error()}

	MESSAGE_DATA_ERROR = errors.New("Ошибка в том, что нет данных в вашем сообщении. Попробуйте как-нибудь по-другому, пожалуйста. :(")
	MESSAGE_DATA_ERROR_RESULT = &s.MessageResult{Type:"chat", Body:MESSAGE_DATA_ERROR.Error()}

	GLOBAL_ERROR = errors.New("Ничего не понятно :(")
	GLOBAL_ERROR_RESULT = &s.MessageResult{Type:"chat", Body:GLOBAL_ERROR.Error()}
)

type CommandsGenerator func(in *s.InPkg) (*[]s.OutCommand, error)

type FuncTextBodyProcessor struct {
	Storage                  *db.MainDb
	F                        CommandsGenerator
	MessageRecipientIdentity string
	AnswerText               *string
}

func (ftbp *FuncTextBodyProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	commands, err := ftbp.F(in)
	if err != nil {
		return s.ErrorMessageResult(err, commands)
	} else {
		if in.UserData != nil {
			err := ftbp.Storage.Users.StoreUserData(in.From, in.UserData)
			if err != nil {
				return DB_ERROR_RESULT
			}
		}
		if in.Message != nil && in.Message.Body != nil {
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

func NewSimpleTextBodyProcessor(storage *db.MainDb, commands *[]s.OutCommand, recipientId string, answerText *string) *FuncTextBodyProcessor {
	f := func(in *s.InPkg) (*[]s.OutCommand, error) {
		return commands, nil
	}
	result := &FuncTextBodyProcessor{Storage:storage, F:f, MessageRecipientIdentity:recipientId, AnswerText:answerText}
	return result
}

func NewFuncTextBodyProcessor(storage *db.MainDb, function CommandsGenerator, recipientId string, answerText *string) *FuncTextBodyProcessor {
	result := &FuncTextBodyProcessor{Storage:storage, F:function, MessageRecipientIdentity:recipientId, AnswerText:answerText}
	return result
}

type InformationProcessor struct {
	Information string
	F           CommandsGenerator
}

func (ip *InformationProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	cmds, err := ip.F(in)
	if err != nil {
		log.Printf("Inforamtion processor: ERROR %v", err)
		return &s.MessageResult{Type:"chat", Body:ip.Information}
	}
	result := s.MessageResult{Type:"chat", Body:ip.Information, Commands:cmds}
	return &result
}

func NewSimpleInformationProcessor(information string) *InformationProcessor {
	return &InformationProcessor{
		Information:information,
		F:func(in *s.InPkg) (*[]s.OutCommand, error) {
			return &[]s.OutCommand{}, nil
		},
	}
}

func NewInformationProcessor(information string, cg CommandsGenerator) *InformationProcessor {
	return &InformationProcessor{Information:information, F:cg}
}


type InformationProcessorUpdatable struct {
	ConfigStore *db.ConfigurationStorage
	Key string
	F           CommandsGenerator
}
func (ip *InformationProcessorUpdatable) ProcessMessage(in *s.InPkg) *s.MessageResult {
	cmds, err := ip.F(in)
	if err != nil {
		log.Printf("Inforamtion processor: ERROR at forming commands %v", err)
		return &s.MessageResult{Type:"chat", Body:err.Error()}
	}
	information, err := ip.ConfigStore.GetInformation(ip.Key)
	log.Printf("Information for %v is %s",ip.Key, *information)
	if err != nil {
		log.Printf("Inforamtion processor: ERROR at getting information %v", err)
		return &s.MessageResult{Type:"chat", Body:err.Error()}
	}
	result := s.MessageResult{Type:"chat", Body:*information, Commands:cmds}
	return &result
}

func NewUpdatableInformationProcessor(confStore *db.ConfigurationStorage, cg CommandsGenerator, key string) *InformationProcessorUpdatable{
	return &InformationProcessorUpdatable{ConfigStore:confStore, Key:key, F:cg}

}