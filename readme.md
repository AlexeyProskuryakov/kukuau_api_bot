use demo_bot.sh for starting and stoping demo bot.
config at congig.json

## After first start run this command:


```
{
    "taxis":[ //here is configuration for taxis extensions
            {
                "name": <name of taxi extension>, //will used when form urls and streets urls
                ...
                "api"{
                    "name": "fake"|"infinity" //will use for logic 
                    "data": { ... }
                }
            }
    ],
    "shops":[ //here is configuration for shops extensions
        {
            "name": <name of shop extension> //will used when form urls
            ...
        }
    ]
    ...    
}
```
# Конфигурация для такси.

Конфигурирование такси происходит в объекте "taxis" (да, во множественном числе). И состоит из описания объекта
в котором сожержатся следующие поля:

  * available_commands - доступные комманды. Внутри поля характеризующие при каком состоянии какие команды доступны. 
  Т.е. каждому полю соответствует массив состояний. ~~Перевыеб конечно, но я же фильстеперстный.~~ 
  Вот таков пример:
  ```
  "available_commands": {
        "created_order": [ //при созданном заказе, доступны вот такие команды. 
          "where_it",  // где машина
          "callback_request", // запрос обратного звонка
          "car_position" // запрос позиции автомобиля (ответ в виде lat & lon)
        ]
      },
  ```
  
  * api - объект конфигурирующий подключение к АПИ такси. О нем чуть позже. Но вкратце, он содержит:
    * name - имя АПИ к которому подключаться можно. Один из [infinity, sedi, master]
    * data - собственно необходимые значения для подключения (хосты, порты, логины и прочее)
  
  * dict_url - урл по которому нужно отправлять запросы на автокомплит. Формируется так: http://<хост>:<п>     
  * geo_orbit - точка и радиус в котором работает это такси. Если не указать, и если заказ будет вне этого радиуса то
    заказ не произойдет и бот ответит что такси с этим адресом не работает. Объект включающий в себя поля lat, lon, radius (в метрах)
  * fake - конфигурация для фейкового такси.
  
  * not_send_price - boolean - отправлять или не отправлять от АПИ такси информацию о цене пользователю.
  * markups - список скидочек (массив номеров скидочек которые следует узнать у определенного такси). Будут включаться в заказ.




## Urls for taxis and shops

When you add new taxi see the *name* parameter because url of any taxi will next:

```
http://<fucking host>:<port from config>/taxi/<name>
```

For shop will similar url:

```
http://<fucking host>:<port from config>/shop/<name>
```


## Dict Url

Lesha! Dict url now must be formed like this:

```
http://<some host>/taxi/<taxi name>/streets
```

It is very important because my server will wait requests for streets in addresses as stated above.

# Ru Post


Add to config

```
"ru_post": {
    "external_url": "http://rupost.info/json", <---this is url for post api
    "work_url":"/ru_post/tracking" <---this is path which will work with handler
  }
```

And url <your_host>/ru_post/tracking will process your requests.
When send request with commands i will ask you only one state response - "tracking" with this command form:

```
    Title: "Форма запроса информации о посылке",
	Type:  "form",
	Name:  "tracking_form",
	Text:  "Номер отправления: ?(code)",
	Fields: {
			Name: "code",
			Type: "number",
			Attributes: {
				Label:    "<последние 14 цифр>",
				Required: true,
				EmptyText: "обязательно",
			},
	},
```

When send message with command "tracking"
I will wait response at result code to tracking at:
message/commands[0]/[action='tracking']/form[name='tracking_form']/field[name='code']/data/value

..That's action must be 'tracking'
..form must be have name = 'tracking_form'
..with field which have name = 'code'

I will return:

All about post message
All about operations

