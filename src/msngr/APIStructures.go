package msngr

/*

from: “username”,
	request: {
		(данные запроса)
}
}

●	from(*) | to(*) - имя получателя или отправителя
●	request(*) - тело запроса
○	id(*) - уникальный идентификатор
○	type(*) - тип запроса, может содержать значения:
■	get - запрос на получение каких-либо данных
■	result - результат запроса
■	error - ошибка возникшая в результате запроса
○	query(*) - если поле “type” имеет значение “get” содержит строку с названием команды,
если имеет значение “result”, то должно содержать результат в виде списка элементов.
Если элементы отсутствуют, список должен быть пустым.


*/

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
	ID      string     `json:"id"`
	Type    string     `json:"type"`
	Thread  string     `json:"thread"`
	Body    *string    `json:"body"`
	Command *InCommand `json:"command"`
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

type InPkg struct {
	From string `json:"from"`

	Message *InMessage `json:"message"`
	Request *InRequest `json:"request"`
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

type Command struct {
	Title    string   `json:"title"`
	Action   string   `json:"action"`
	Position int      `json:"position"`
	Form     *OutForm `json:"form,omitempty"`
}

type OutMessage struct {
	ID     string `json:"id"`
	Type   string `json:"type,omitempty"`
	Thread string `json:"thread,omitempty"`
	Body   string `json:"body"`
}

type OutRequest struct {
	ID    string `json:"id,omitempty"`
	Type  string `json:"type,omitempty"`
	Query struct {
		Title  string    `json:"title,omitempty"`
		Action string    `json:"action"`
		Text   string    `json:"text,omitempty"`
		Form   *OutForm  `json:"form,omitempty"`
		Result []Command `json:"result,omitempty"`
	} `json:"query"`
}

type OutPkg struct {
	To      string      `json:"to"`
	Message *OutMessage `json:"message,omitempty"`
	Request *OutRequest `json:"request,omitempty"`
}
