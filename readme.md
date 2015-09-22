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

