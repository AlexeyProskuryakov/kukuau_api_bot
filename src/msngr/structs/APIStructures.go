package structs

import (
	"fmt"
	"time"
	"log"
)

type FieldAttribute struct {
	Label     string  `json:"label",bson:"label"`
	Required  bool    `json:"required",bson:"required"`
	Regex     *string `json:"regex,omitempty",bson:"regex,omitempty"`
	URL       *string `json:"url,omitempty",bson:"url,omitempty"`
	EmptyText *string `json:"empty_text,omitempty",bson:"empty_text,omitempty"`
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

func (i_f InForm) GetValue(fieldName string) (string, bool) {
	for _, f := range i_f.Fields {
		if f.Name == fieldName {
			return f.Data.Value, true
		}
	}
	return "", false
}

func (i_f InForm) GetText(fieldName string) (string, bool) {
	for _, f := range i_f.Fields {
		if f.Name == fieldName {
			return f.Data.Text, true
		}
	}
	return "", false
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

type MessageError struct {
	Code      int `json:"code"`
	Type      string `json:"type"`
	Condition string `json:"condition"`
}

type InMessage struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Thread   string       `json:"thread"`
	Body     *string      `json:"body"`
	Commands *[]InCommand `json:"commands"`
	Error    *MessageError `json:"error"`
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
	Name  string `json:"name"`
	Email string `json:"e-mail"`
}

type InPkg struct {
	From     string      `json:"from"`
	UserData *InUserData `json:"userdata,omitempty"`
	Message  *InMessage  `json:"message"`
	Request  *InRequest  `json:"request"`
}

type FieldItemContent struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Icon     string `json:"icon"`
}

type FieldItem struct {
	Value   string `json:"value"`
	Content FieldItemContent `json:"content"`
}

type OutField struct {
	Name       string `json:"name",bson:"name"`
	Type       string `json:"type,omitempty",bson:"type,omitempty"`
	Data       *struct {
	} `json:"data,omitempty",bson:"data,omitempty"`
	Attributes FieldAttribute `json:"attrs",bson:"attrs"`
	Items      []FieldItem `json:"items"`
}

type OutForm struct {
	Title  string     `json:"title",bson:"title"`
	Text   string     `json:"text",bson:"text"`
	Type   string     `json:"type,omitempty",bson:"type,omitempty"`
	Name   string     `json:"name",bson:"name"`
	Label  string     `json:"label,omitempty",bson:"label,omitempty"`
	URL    string     `json:"url,omitempty",bson:"url,omitempty"`
	Fields []OutField `json:"fields,omitempty",bson:"fields,omitempty"`
}

type OutCommand struct {
	Title    string   `json:"title",bson:"title"`
	Action   string   `json:"action",bson:"action"`
	Position int      `json:"position",bson:"position"`
	Fixed    bool     `json:"fixed",bson:"fixed"`
	Repeated bool     `json:"repeated",bson:"repeated"`
	Form     *OutForm `json:"form,omitempty",bson:"form,omitempty"`
}

func (oc OutCommand) String() string {
	return fmt.Sprintf("Command to send:\n\t%v[%v], position:%v, fixed? %v, repeated? %v, \n\t\tform: %+v;", oc.Title, oc.Action, oc.Position, oc.Fixed, oc.Repeated, oc.Form)
}

type OutMessage struct {
	ID       string        `json:"id"`
	Type     string        `json:"type,omitempty"`
	Thread   string        `json:"thread,omitempty"`
	Body     string        `json:"body"`
	Error    *MessageError  `json:"error,omitempty"`
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

func (o OutPkg) String() string {
	return fmt.Sprintf("OUT{TO:%s, MESSAGE:%+v, REQUEST:%+v}", o.To, o.Message, o.Request)
}

type CheckFunc func() (string, bool)

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

type AutocompleteDictItem struct {
	Key      string `json:"key"`
	Title    string `json:"title"`
	SubTitle string `json:"subtitle"`
}
