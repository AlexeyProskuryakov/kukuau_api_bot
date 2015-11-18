package structs
import (
	"fmt"
	"time"
	"log"
)

type FieldAttribute struct {
	Label     string  `json:"label"`
	Required  bool    `json:"required"`
	Regex     *string `json:"regex,omitempty"`
	URL       *string `json:"url,omitempty"`
	EmptyText *string `json:"empty_text,omitempty"`
}

type InForm struct {
	Title  string    `json:"title,omitempty"`
	Text   string    `json:"text,omitempty"`
	Type   string    `json:"type,omitempty"`
	Name   string    `json:"name,omitempty"`
	Label  string    `json:"label,omitempty"`
	URL    string    `json:"url,omitempty"`
	Fields []InField `json:"fields,omitempty"`
}

type InField struct {
	Name string `json:"name"`
	Type string `json:"type,omitempty"`
	Data InFieldData `json:"data,omitempty"`
}

func (i InField) String() string {
	return fmt.Sprintf("\nName:%s\nType:%s\nData:%+v\n", i.Name, i.Type, i.Data)
}
type InFieldData struct {
	Value string `json:"value"`
	Text  string `json:"text"`
}
type InCommand struct {
	Title  string `json:"title,omitempty"`
	Action string `json:"action"`
	Form   InForm `json:"form"`
}
type InMessage struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Thread   string       `json:"thread"`
	Body     *string      `json:"body"`
	Commands *[]InCommand `json:"commands"`
}

type InRequest struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Query struct {
			  Title  string `json:"title,omtempty"`
			  Action string `json:"action"`
			  Form   InForm `json:"form"`
		  } `json:"query"`
}

type InUserData struct {
	Phone string `json:"phone"`
}

type InPkg struct {
	From     string      `json:"from"`
	UserData *InUserData `json:"userdata,omitempty"`
	Message  *InMessage  `json:"message"`
	Request  *InRequest  `json:"request"`
}

type OutField struct {
	Name       string `json:"name"`
	Type       string `json:"type,omitempty"`
	Data       *struct {
	} `json:"data,omitempty"`
	Attributes FieldAttribute `json:"attrs"`
}

type OutForm struct {
	Title  string     `json:"title,omitempty"`
	Text   string     `json:"text,omitempty"`
	Type   string     `json:"type,omitempty"`
	Name   string     `json:"name,omitempty"`
	Label  string     `json:"label,omitempty"`
	URL    string     `json:"url,omitempty"`
	Fields []OutField `json:"fields,omitempty"`
}

type OutCommand struct {
	Title    string   `json:"title"`
	Action   string   `json:"action"`
	Position int      `json:"position"`
	Fixed    bool     `json:"fixed"`
	Repeated bool     `json:"repeated"`
	Form     *OutForm `json:"form,omitempty"`
}

type OutMessage struct {
	ID       string        `json:"id"`
	Type     string        `json:"type,omitempty"`
	Thread   string        `json:"thread,omitempty"`
	Body     string        `json:"body"`
	Commands *[]OutCommand `json:"commands,omitempty"`
}

type OutRequest struct {
	ID    string `json:"id,omitempty"`
	Type  string `json:"type,omitempty"`
	Query struct {
			  Title  string       `json:"title,omitempty"`
			  Action string       `json:"action"`
			  Text   string       `json:"text,omitempty"`
			  Form   *OutForm     `json:"form,omitempty"`
			  Result []OutCommand `json:"result,omitempty"`
		  } `json:"query"`
}

type OutPkg struct {
	To      string      `json:"to"`
	Message *OutMessage `json:"message,omitempty"`
	Request *OutRequest `json:"request,omitempty"`
}

type CheckFunc func() (string, bool)

type BotContext struct {
	Name             string
	Check            CheckFunc
	Request_commands map[string]RequestCommandProcessor
	Message_commands map[string]MessageCommandProcessor
	Commands         map[string]*[]OutCommand
	Settings         map[string]interface{}
}

type MessageResult struct {
	Commands   *[]OutCommand
	Body       string
	Error      error
	IsDeferred bool
	Type       string
}

type RequestResult struct {
	Commands *[]OutCommand
	Error    error
	Type     string
}

type RequestCommandProcessor interface {
	ProcessRequest(in *InPkg) *RequestResult
}

type MessageCommandProcessor interface {
	ProcessMessage(in *InPkg) *MessageResult
}

func ExceptionMessageResult(err error) *MessageResult {
	return &MessageResult{Body:fmt.Sprintf("Ошибка! %v \n Попробуйте еще раз позже.", err), Type:"error"}
}
//todo
func ErrorMessageResult(err error, commands *[]OutCommand) *MessageResult {
	result := MessageResult{Body:fmt.Sprintf("Ошибка! %v", err), Type:"chat"}
	if commands != nil {
		result.Commands = commands
	}
	return &result
}

func ExceptionRequestResult(err error, commands *[]OutCommand) *RequestResult {
	return &RequestResult{Error:err, Commands:commands}
}

func StartAfter(check CheckFunc, what func()) {
	for {
		if message, ok := check(); ok {
			break
		}else {
			log.Printf("wait %v", message)
			time.Sleep(5 * time.Second)
		}
	}
	go what()

}
