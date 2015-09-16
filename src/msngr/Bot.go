package msngr

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
	"errors"
)

func _check(e error) {
	if e != nil {
		panic(e)
	}
}

func get_time_after(d time.Duration, format string) string {
	result := time.Now().Add(d)
	return result.Format(format)
}

func genId() string {
	//не привязывайся ко времени, может бть в 1 микросекуну много сообщений и ид должны ыть разными
	return fmt.Sprintf("%d", time.Now().UnixNano() / int64(time.Millisecond))
}

func getInPackage(r *http.Request) (InPkg, error) {
	var in InPkg
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("error at reading: %q \n", err)
	}
	log.Printf("<<<:%s", string(body))
	err = json.Unmarshal(body, &in)
	if err != nil {
		log.Printf("error at unmarshal: %q \n", err)
	}
	return in, err
}

func setOutPackage(w http.ResponseWriter, out OutPkg, isError bool, isDeferred bool) {

	jsoned_out, err := json.Marshal(&out)
	if err != nil {
		log.Println("set out package: ", jsoned_out, err)
	}
	w.Header().Set("Content-type", "application/json")

	log.Printf(">>> %s\n", string(jsoned_out))

	if isError{
		w.WriteHeader(http.StatusBadRequest)
	} else if isDeferred{
		w.WriteHeader(http.StatusNoContent)
		return
	}else{
		w.WriteHeader(http.StatusOK)
	}

	fmt.Fprintf(w, "%s", string(jsoned_out))
}

type checkFunc func() (string, bool)

type BotContext struct {
	Check checkFunc
	Request_commands map[string]RequestCommandProcessor
	Message_commands map[string]MessageCommandProcessor
}


type controllerHandler func(w http.ResponseWriter, r *http.Request)

func FormBotController(context *BotContext) controllerHandler {

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "I can not work with non POST methods", 405)
			return
		}

		out := &OutPkg{}
		var in InPkg
		var isError, isDeferred bool
		var global_error, request_error, message_error error

		if detail, ok := context.Check(); !ok{
			out.Message = &OutMessage{Type: "error", Thread: "0", ID: genId(), Body: fmt.Sprintln(detail)}
			setOutPackage(w, *out, true, isDeferred)
			return
		}

		in, global_error = getInPackage(r)

		out.To = in.From

		if in.Request != nil {
			log.Printf("processing request %+v", in)
			action := in.Request.Query.Action
			out.Request = &OutRequest{ID: genId(), Type: "result"}
			out.Request.Query.Action = action
			if commandProcessor, ok := context.Request_commands[action]; ok {
				requestResult := commandProcessor.ProcessRequest(in)
				if requestResult.Error != nil {
					request_error = requestResult.Error
				}else{
					out.Request.Query.Result = *requestResult.Commands
				}
			} else {
				request_error = errors.New("Команда не поддерживается.")
			}

		} else if in.Message != nil {
			log.Printf("processing message %+v", in)
			out.Message = &OutMessage{Type: in.Message.Type, Thread: in.Message.Thread, ID: genId()}

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
						}else{
							out.Message.Body = messageResult.Body
							out.Message.Commands = messageResult.Commands

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
			out = &OutPkg{}
			out.Message = &OutMessage{Type: "error", Thread: "0", ID: genId(), Body: fmt.Sprintf("%+v", message_error)}
			isError = true
		} else if global_error != nil {
			out = &OutPkg{}
			out.Message = &OutMessage{Type: "error", Thread: "0", ID: genId(), Body: fmt.Sprintf("%+v", global_error)}
			isError = true
		} else if request_error != nil {
			out = &OutPkg{}
			out.Request = &OutRequest{Type: "error", ID: genId()}
			out.Request.Query.Text = fmt.Sprintf("%+v", request_error)
			isError = true
		}

		setOutPackage(w, *out, isError, isDeferred)
	}

}
