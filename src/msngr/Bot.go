package msngr

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"errors"
	"strings"

	s "msngr/structs"
	u "msngr/utils"
	"msngr/configuration"
	db "msngr/db"
)

var DEBUG bool
var TEST bool
//var textProvider = tm.NewTextMessageSupplier()


type BotContext struct {
	Name              string
	Check             s.CheckFunc
	RequestProcessors map[string]s.RequestCommandProcessor
	MessageProcessors map[string]s.MessageCommandProcessor
	Commands          map[string]*[]s.OutCommand
	CommandsStorage   configuration.ConfigStorage
	Settings          map[string]interface{}
}

func (bc BotContext) String() string {
	check, ok := bc.Check()
	var available_cmds string
	for name, cmds := range bc.Commands {
		cmds_represent := []string{}
		for _, cmd := range *cmds {
			cmds_represent = append(cmds_represent, cmd.String())
		}
		available_cmds += fmt.Sprintf("\n\tFor state: [%v] next commands: \n\t%v", name, strings.Join(cmds_represent, "\n\t"))
	}
	return fmt.Sprintf("\nBot context for %v\nChecked?: %v (%v)\nRequestCommands: %+v\nMessageCommands: %+v\nAvailable commands: %+v\nSettings: %+v\n", bc.Name, ok, check, bc.RequestProcessors, bc.MessageProcessors, available_cmds, bc.Settings)
}

func FormInPackage(r *http.Request) (*s.InPkg, error) {
	var in s.InPkg
	if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		return nil, errors.New("No header `Content-Type` or his value is not `application/json`")
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("BOT: error at reading: %q \n", err)
	}
	if DEBUG {
		log.Printf("BOT: RECEIVED: \n%s\n", string(body))
	}

	err = json.Unmarshal(body, &in)
	if err != nil {
		log.Printf("BOT: error at unmarshal: %q \n %s", err, body)
	}
	return &in, err
}

func PutOutPackage(w http.ResponseWriter, out *s.OutPkg, isError bool, isDeferred bool) {
	jsoned_out, err := json.Marshal(out)
	if err != nil {
		log.Println("set out package: ", jsoned_out, err)
	}
	w.Header().Set("Content-Type", "application/json")
	if DEBUG {
		log.Printf("BOT RESPONSED: \n%s\n", string(jsoned_out))
	}
	if isError {
		w.WriteHeader(http.StatusInternalServerError)
	} else if isDeferred {
		w.WriteHeader(http.StatusNoContent)
		return
	}else {
		w.WriteHeader(http.StatusOK)
	}
	fmt.Fprintf(w, "%s", string(jsoned_out))
}

func process_request_pkg(buff *s.OutPkg, in *s.InPkg, context *BotContext) (*s.OutPkg, error) {
	if in.Request.Type == "error" {
		log.Printf("error because type of request is error:\n %+v", in.Request)
		return buff, errors.New("error because request type is error")
	}
	action := in.Request.Query.Action
	buff.Request = &s.OutRequest{ID: u.GenId(), Type: "result"}
	buff.Request.Query.Action = action
	buff.Request.Type = "result"

	if commandProcessor, ok := context.RequestProcessors[action]; ok {
		requestResult := commandProcessor.ProcessRequest(in)
		if requestResult.Error != nil {
			err := requestResult.Error
			return buff, err
		}else {
			//normal our request forming
			buff.Request.Query.Result = *requestResult.Commands
			if requestResult.Type != "" {
				buff.Request.Type = requestResult.Type
			}
		}
	} else {
		err := errors.New("Команда не поддерживается.")
		return buff, err
	}
	return buff, nil
}

func process_message(commandProcessor s.MessageCommandProcessor, buff *s.OutPkg, in *s.InPkg) (*s.OutPkg, bool, error) {
	messageResult := commandProcessor.ProcessMessage(in)
	buff.Message = &s.OutMessage{
		Thread: in.Message.Thread,
		ID: u.GenId(),
		Type:"chat",
	}
	//normal buff message forming
	if messageResult.Type != "" {
		buff.Message.Type = messageResult.Type
	}
	buff.Message.Body = messageResult.Body
	buff.Message.Commands = messageResult.Commands

	log.Printf("message result\ntype: %+v \nbody:%+v\ncommands:%+v\ndeffered?: %+v", messageResult.Type, buff.Message.Body, buff.Message.Commands, messageResult.IsDeferred)

	return buff, messageResult.IsDeferred, messageResult.Error
}

const (
	ERR_BAD_FORMAT = "bad-format"

	ERR_BAD_REQUEST = "bad-request"

	ERR_CONFLICT = "conflict"

	ERR_FEATURE_NOT_IMPLEMENTED = "feature-not-implemented"

	ERR_FORBIDDEN = "forbidden"
	ERR_GONE = "gone"
	ERR_INTERNAL_SERVER_ERROR = "internal-server-error"

	ERR_ITEM_NOT_FOUND = "item-not-found"

	ERR_JID_MALFORMED = "jid-malformed"

	ERR_NOT_ACCEPTABLE = "not-acceptable"

	ERR_NOT_ALLOWED = "not-allowed"

	ERR_NOT_AUTHORIZED = "not-authorized"

	ERR_PAYMENT_REQUIRED = "payment-required"

	ERR_RECIPIENT_UNAVAILABLE = "recipient-unavailabled"

	ERR_REDIRECT = "redirect"

	ERR_REGISTRATION_REQUIRED = "registration-required"

	ERR_REMOTE_SERVER_NOT_FOUND = "remote-server-not-found"

	ERR_REMOTE_SERVER_TIMEOUT = "remote-server-timeout"

	ERR_RESOURCE_CONSTRAINT = "resource-constraint"

	ERR_SERVICE_UNAVAILABLE = "service-unavailable"

	ERR_SUBSCRIPTION_REQUIRED = "subscription-required"

	ERR_UNEXPECTED_REQUEST = "unexpected-request"

	ERR_UNEXPECTED_REQUEST_CANCEL = "unexpected-request_cancel"
)

var ERRORS_MAP = map[string]func() (int, string){
	ERR_BAD_FORMAT: func() (int, string) {
		return 406, "modify"
	},
	ERR_BAD_REQUEST: func() (int, string) {
		return 400, "modify"
	},
	ERR_CONFLICT: func() (int, string) {
		return 409, "cancel_conflict"
	},
	ERR_FEATURE_NOT_IMPLEMENTED: func() (int, string) {
		return 501, "cancel"
	},
	ERR_FORBIDDEN: func() (int, string) {
		return 403, "auth_forbidden"
	},
	ERR_GONE: func() (int, string) {
		return 302, "modify_gone"
	},
	ERR_INTERNAL_SERVER_ERROR: func() (int, string) {
		return 500, "wait"
	},
	ERR_ITEM_NOT_FOUND: func() (int, string) {
		return 404, "cancel"
	},

	ERR_JID_MALFORMED: func() (int, string) {
		return 400, "modify"
	},
	ERR_NOT_ACCEPTABLE: func() (int, string) {
		return 406, "modify"
	},
	ERR_NOT_ALLOWED: func() (int, string) {
		return 405, "cancel"
	},
	ERR_NOT_AUTHORIZED: func() (int, string) {
		return 401, "auth"
	},
	ERR_PAYMENT_REQUIRED: func() (int, string) {
		return 402, "auth"
	},
	ERR_RECIPIENT_UNAVAILABLE: func() (int, string) {
		return 404, "wait"
	},
	ERR_REDIRECT: func() (int, string) {
		return 302, "modify_redirect"
	},
	ERR_REGISTRATION_REQUIRED: func() (int, string) {
		return 407, "auth"
	},
	ERR_REMOTE_SERVER_NOT_FOUND: func() (int, string) {
		return 404, "cancel"
	},
	ERR_REMOTE_SERVER_TIMEOUT: func() (int, string) {
		return 504, "wait"
	},
	ERR_RESOURCE_CONSTRAINT: func() (int, string) {
		return 500, "wait"
	},
	ERR_SERVICE_UNAVAILABLE: func() (int, string) {
		return 503, "cancel"
	},
	ERR_SUBSCRIPTION_REQUIRED: func() (int, string) {
		return 407, "auth"
	},
	ERR_UNEXPECTED_REQUEST: func() (int, string) {
		return 400, "wait"
	},
	ERR_UNEXPECTED_REQUEST_CANCEL:func() (int, string) {
		return 401, "cancel"
	},
}

func process_message_pkg(buff *s.OutPkg, in *s.InPkg, context *BotContext) (*s.OutPkg, bool, error) {
	var err error
	var isDeferred bool

	if in.Message.Type == "error" {
		log.Printf("error because type of message is error:\n %+v", in.Message.Error)
		return buff, false, errors.New(fmt.Sprintf("Error because type of message id: %+v is error", in.Message.ID))
	}

	in_commands := in.Message.Commands
	for _, command := range *in_commands {
		action := command.Action
		if commandProcessor, ok := context.MessageProcessors[action]; ok {
			log.Printf("BOT found [%v] action", action)
			buff, isDeferred, err = process_message(commandProcessor, buff, in)
		} else {
			err = errors.New("Команда не поддерживается.")
			buff.Message = &s.OutMessage{
				Thread:in.Message.Thread,
				ID :u.GenId(),
				Type:"chat",
				Body: err.Error(),
			}
		}
	}
	return buff, isDeferred, err
}

type controllerHandler func(w http.ResponseWriter, r *http.Request)

func FormBotController(context *BotContext, db *db.MainDb) controllerHandler {

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "I can not work with non POST methods", 405)
			return
		}

		out := &s.OutPkg{}
		var in *s.InPkg
		var isError, isDeferred bool
		var global_error, request_error, message_error error

		check := context.Check
		if check != nil {
			if detail, ok := check(); !ok {
				out.Message = &s.OutMessage{Type: "chat", Thread: "0", ID: u.GenId(), Body: fmt.Sprintln(detail)}
				PutOutPackage(w, out, true, false)
				return
			}
		}

		in, global_error = FormInPackage(r)
		if in != nil {
			out.To = in.From
			if in.Request != nil {
				out, request_error = process_request_pkg(out, in, context)
			}
			if in.Message != nil {
				storedMessage, _ := db.Messages.GetMessageByMessageId(in.Message.ID)
				log.Printf("BOT in message id %v", in.Message.ID)
				if storedMessage != nil {
					log.Printf("BOT: Have duplicate message. Will be quiet ignoring it...")
					return
				}
				if in.Message.Error != nil {
					if val, ok := ERRORS_MAP[in.Message.Error.Condition]; ok {
						_, err_type := val()
						log.Printf("BOT: Here is error %v %v. For message id %v", in.Message.Error.Code, err_type, in.Message.ID)
						db.Messages.UpdateMessageStatus(in.Message.ID, err_type, in.Message.Error.Condition)
					} else {
						log.Printf("BOT: Error not in my errors map %v", in.Message.Error.Condition)
						db.Messages.UpdateMessageStatus(in.Message.ID, in.Message.Error.Type, in.Message.Error.Condition)
					}
					return
				}
				if in.Message.Commands == nil {
					if non_commands_processor, ok := context.MessageProcessors[""]; ok {
						out, isDeferred, message_error = process_message(non_commands_processor, out, in)
					} else {
						log.Printf("warn will sended message without commands: %+v\n from %v (userdata: %+v)", in.Message, in.From, in.UserData)
					}
				} else {
					out, isDeferred, message_error = process_message_pkg(out, in, context)
				}
			}
			if in.Message == nil && in.Request == nil {
				global_error = errors.New("Ничего не понятно!")
			}
		}

		if DEBUG {
			log.Printf("package processed!\nrequest:%+v\nmessage:%+v\nmessage\nrequest_error: %+v, message error: %+v", out.Request, out.Message, request_error, message_error)
		}

		if message_error != nil {
			out = &s.OutPkg{}
			out.Message = &s.OutMessage{Type: "error", Thread: "0", ID: u.GenId(), Body: fmt.Sprintf("%+v", message_error)}
			isError = true
		} else if global_error != nil {
			out = &s.OutPkg{}
			out.Message = &s.OutMessage{Type: "error", Thread: "0", ID: u.GenId(), Body: fmt.Sprintf("%+v", global_error)}
			isError = true
		} else if request_error != nil {
			out = &s.OutPkg{}
			out.Request = &s.OutRequest{Type: "error", ID: u.GenId()}
			out.Request.Query.Text = fmt.Sprintf("%+v", request_error)
			isError = true
		}

		PutOutPackage(w, out, isError, isDeferred)
	}

}

