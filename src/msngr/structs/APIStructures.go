package structs

type FieldAttribute struct {
	Label    string  `json:"label"`
	Required bool    `json:"required"`
	Regex    *string `json:"regex,omitempty"`
	URL      *string `json:"url,omitempty"`
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
	Data struct {
		Value string `json:"value"`
		Text  string `json:"text"`
	} `json:"data,omitempty"`
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
	Name string `json:"name"`
	Type string `json:"type,omitempty"`
	Data *struct {
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


type checkFunc func() (string, bool)

type BotContext struct {
	Check checkFunc
	Request_commands map[string]RequestCommandProcessor
	Message_commands map[string]MessageCommandProcessor
}

type MessageResult struct {
	Commands *[]OutCommand
	Body string
	Error error
	IsDeferred bool
}

type RequestResult struct {
	Commands *[]OutCommand
	Error error
}

type RequestCommandProcessor interface {
	ProcessRequest(in InPkg) RequestResult
}

type MessageCommandProcessor interface {
	ProcessMessage(in InPkg) MessageResult
}
