package msngr

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"errors"
	s "msngr/structs"
	u "msngr/utils"
)

func _check(e error) {
	if e != nil {
		panic(e)
	}
}

func getInPackage(r *http.Request) (*s.InPkg, error) {
	var in s.InPkg
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("error at reading: %q \n", err)
	}
	log.Printf("<<<:%s", string(body))
	err = json.Unmarshal(body, &in)
	if err != nil {
		log.Printf("error at unmarshal: %q \n", err)
	}
	return &in, err
}

func setOutPackage(w http.ResponseWriter, out *s.OutPkg, isError bool, isDeferred bool) {
	jsoned_out, err := json.Marshal(out)
	if err != nil {
		log.Println("set out package: ", jsoned_out, err)
	}
	w.Header().Set("Content-type", "application/json")

	log.Printf(">>> %s\n", string(jsoned_out))

	if isError {
		w.WriteHeader(http.StatusBadRequest)
	} else if isDeferred {
		w.WriteHeader(http.StatusNoContent)
		return
	}else {
		w.WriteHeader(http.StatusOK)
	}

	fmt.Fprintf(w, "%s", string(jsoned_out))
}

type controllerHandler func(w http.ResponseWriter, r *http.Request)

func FormBotController(context *s.BotContext) controllerHandler {

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
				out.Message = &s.OutMessage{Type: "error", Thread: "0", ID: u.GenId(), Body: fmt.Sprintln(detail)}
				setOutPackage(w, out, true, false)
				return
			}
		}

		in, global_error = getInPackage(r)

		out.To = in.From

		if in.Request != nil {
			action := in.Request.Query.Action
			out.Request = &s.OutRequest{ID: u.GenId(), Type: "result"}
			out.Request.Query.Action = action
			if commandProcessor, ok := context.Request_commands[action]; ok {
				requestResult := commandProcessor.ProcessRequest(in)
				if requestResult.Error != nil {
					request_error = requestResult.Error
				}else {
					//normal our request forming
					out.Request.Query.Result = *requestResult.Commands
				}
			} else {
				request_error = errors.New("Команда не поддерживается.")
			}

		} else if in.Message != nil {
			out.Message = &s.OutMessage{Type: in.Message.Type, Thread: in.Message.Thread, ID: u.GenId()}

			in_commands := in.Message.Commands

			if in_commands == nil {
				message_error = errors.New("Команд не найдено.")
			} else {
				for _, command := range *in_commands {
					action := command.Action
					if commandProcessor, ok := context.Message_commands[action]; ok {
						messageResult := commandProcessor.ProcessMessage(in)
						if messageResult.Error != nil {
							message_error = messageResult.Error
						}else {
							//normal out message forming
							if messageResult.Type != "" {
								out.Message.Type = "chat"
							}else {
								out.Message.Type = messageResult.Type
							}
							out.Message.Body = messageResult.Body
							out.Message.Commands = messageResult.Commands
							isDeferred = messageResult.IsDeferred
						}
					} else {
						message_error = errors.New("Команда не поддерживается.")
					}
				}
			}
		} else {
			global_error = errors.New("Ничего не понятно!")
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

		setOutPackage(w, out, isError, isDeferred)
	}

}
