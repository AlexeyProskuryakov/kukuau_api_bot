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
	log.Println("getting in! will retrieve body from request")
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
	log.Println("forming out! will marshaing out response")

	jsoned_out, err := json.Marshal(&out)
	if err != nil {
		log.Println(jsoned_out, err)
	}
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)

	fmt.Fprintf(w, "%s", string(jsoned_out))
}

type controllerHandler func(w http.ResponseWriter, r *http.Request)

func FormBotControllerHandler(request_cmds map[string]RequestCommandProcessor, message_cmds map[string]MessageCommandProcessor) controllerHandler {

	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("processing taxi request...")
		if r.Method != "POST" {
			http.Error(w, "I can not work with non POST methods", 405)
			return
		}

		in, err := getInPackage(r)
		out := new(OutPkg)

		log.Println("forming response...")

		out.To = in.From

		if in.Request != nil {
			log.Println("processing request")
			action := in.Request.Query.Action
			out.Request = &OutRequest{ID: genId(), Type: "result"}
			out.Request.Query.Action = action
			if commandProcessor, ok := request_cmds[action]; ok {
				out.Request.Query.Result, err = commandProcessor.ProcessRequest(in)
			} else {
				out.Request.Query.Text = "Команда не поддерживается."
			}

		} else if in.Message != nil {
			log.Println("processing message")
			out.Message = &OutMessage{Type: in.Message.Type, Thread: in.Message.Thread, ID: genId()}
			action := in.Message.Command.Action
			if commandProcessor, ok := message_cmds[action]; ok {
				out.Message.Body, err = commandProcessor.ProcessMessage(in)
			} else {
				out.Message.Body = "Команда не поддерживается."
			}

		}
		log.Printf("%+v\n", out)

		if err != nil {
			out.Message = &OutMessage{Type: "error", Thread: "0", ID: genId(), Body: fmt.Sprintf("%+v", err)}
		}

		setOutPackage(w, *out)
	}

}
