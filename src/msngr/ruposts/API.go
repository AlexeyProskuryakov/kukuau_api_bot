package ruposts
import (
	"encoding/json"
	u "msngr/utils"
	"log"
)
/*






id	Число от -10 до 99	Идентификатор ответа, отрицательное значение - ошибка, нулевое - отсутствие результата, положительное - количество операций
msg	Текст или объект	При положительных значениях идентификатора содержит два объекта info и operations. При нулевом или отрицательных значениях - текст сообщения.
info	Объект	Содержит поля:
code - номер идентификатора
name - наименование отправления
destination - адрес назначения: id, county, index, adress
weight - вес отправления в граммах
category - общие категории отправления: info, rank, mark, type
weight - финансовые категории отправления: payment, value, weight, insurance, air, rate
latest - информация о последней операции
operations	Объект	Содержит поля:
date - дата в формате "день месяц год, часы минуты"
dateiso - дата в формате ISO 8601
timestamp - дата в формате Unix Timestamp
adress - адрес регистрации: index, description
operation - наименование операции
attr - описание операции
 */

type PostInformation struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	Destination string `json:"destination"`
	Weight      int `json:"weight"`
	Category    string `json:"category"`
	Latest      string `json:"latest"`
}
type TrackingOperation struct {
	Date      string `json:"date"`
	DateIso   string `json:"dateiso"`
	Timestamp float32 `json:"timestamp"`
	Address   struct {
				  Index       int `json:"index"`
				  Description string `json:"description"`
			  } `json:"adress"`
	Operation string `json:"operation"`
	Attribute string `json:"attr"`
}
type PostTrackingWrapper struct {
	ResponseId int `json:"id"`
	Message string `json:"msg"`
	Info       PostInformation `json:"info"`
	Operations []TrackingOperation `json:"operations"`
}

func Load(code string, url string) (*PostTrackingWrapper, error) {
	data, err := u.GET(url, &map[string]string{"code":code})
	if err != nil {
		log.Printf("RUPOST LOAD err %v at getting request %v?code=%v", err, url, code)
		return nil, err
	}
	psw := PostTrackingWrapper{}
	err = json.Unmarshal(*data, &psw)
	if err != nil {
		log.Printf("RUPOST LOAD err %v\n at unmarshal %q ", err, string(*data))
		return nil, err
	}
	return &psw, nil
}