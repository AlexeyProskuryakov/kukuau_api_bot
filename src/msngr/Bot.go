package msngr

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func genId() string {
	//не привязывайся ко времени, может бть в 1 микросекуну много сообщений и ид долэны ыть разными
	return fmt.Sprintf("%d", time.Now().UnixNano()/int64(time.Millisecond))
}

func getInPackage(r *http.Request) (InPkg, error) {

	var in InPkg

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("error at reading: %q \n", err)
	}

	err = json.Unmarshal(body, &in)
	if err != nil {
		log.Printf("error at unmarshal: %q \n", err)
	}

	log.Printf("request data is:\n%+v\n", in)
	return in, err
}

func setOutPackage(w http.ResponseWriter, out OutPkg) {

	jsoned_out, err := json.Marshal(&out)
	if err != nil {
		log.Println("set out package: ", jsoned_out, err)
	}
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)

	fmt.Fprintf(w, "%s", string(jsoned_out))
}

type controllerHandler func(w http.ResponseWriter, r *http.Request)

func FormBotController(request_cmds map[string]RequestCommandProcessor, message_cmds map[string]MessageCommandProcessor) controllerHandler {

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "I can not work with non POST methods", 405)
			return
		}

		in, err := getInPackage(r)
		out := new(OutPkg)

		log.Println("forming response...")

		out.To = in.From

		if in.Request != nil {
			log.Printf("processing request %+v", in)
			action := in.Request.Query.Action
			out.Request = &OutRequest{ID: genId(), Type: "result"}
			out.Request.Query.Action = action
			if commandProcessor, ok := request_cmds[action]; ok {
				out.Request.Query.Result, err = commandProcessor.ProcessRequest(in)
			} else {
				out.Request.Query.Text = "Команда не поддерживается."
			}

		} else if in.Message != nil {
			log.Printf("processing message %+v", in)
			out.Message = &OutMessage{Type: in.Message.Type, Thread: in.Message.Thread, ID: genId()}

			in_commands := in.Message.Commands

			if in_commands == nil {
				out.Message.Body = "Команд не найдено"
			} else {
				for _, command := range *in_commands {
					action := command.Action
					if commandProcessor, ok := message_cmds[action]; ok {
						out.Message.Body, out.Message.Commands, err = commandProcessor.ProcessMessage(in)
					} else {
						out.Message.Body = "Команда не поддерживается."
					}
				}
			}

		}
		log.Printf("out >>> %+v\n", out)

		if err != nil {
			out.Message = &OutMessage{Type: "error", Thread: "0", ID: genId(), Body: fmt.Sprintf("%+v", err)}
		}

		setOutPackage(w, *out)
	}

}
