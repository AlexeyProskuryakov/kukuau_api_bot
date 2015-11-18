use demo_bot.sh for starting and stoping demo bot.
config at congig.json

## After first start run this command:


go get gopkg.in/mgo.v2

See update of config! 


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



May be also add difference at auth key?


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

