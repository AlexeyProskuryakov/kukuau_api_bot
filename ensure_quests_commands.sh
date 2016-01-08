curl --user alesha:sederfes100500 -i -L -X POST \
   -H "Content-Type:application/json" \
   -d \
'{
  "provider":"quests",
  "name":"unsubscribed",
  "command":{
    "title":"Учавствовать",
    "action":"subscribe",
    "position":0,
    "repeated":false
  }
}' \
 'http://localhost:9595/configuration'

 curl --user alesha:sederfes100500 -i -L -X POST \
   -H "Content-Type:application/json" \
   -d \
'{
  "provider":"quests",
  "name":"subscribed",
  "command":{
    "title":"Ввод найденного кода",
    "action":"key_input",
    "position":0,
    "repeated":false,
    "form":{
      "title":"Форма ввода ключа для следующего задания",
      "type":"form",
      "name":"key_form",
      "text":"Код: ?(code)",
      "fields":[
        {"name":"code",
         "type":"text",
         "attrs":{
           "label":"Ваш найденный код",
           "required":true
         	}
        }
      ]
    }
  }
}' \
 'http://localhost:9595/configuration'

 curl --user alesha:sederfes100500 -i -L -X POST \
   -H "Content-Type:application/json" \
   -d \
'{
  "provider":"quests",
  "name":"subscribed",
  "command":{
    "title":"Перестать учавствовать",
    "action":"unsubscribe",
    "position":1,
    "repeated":false
  }
}' \
 'http://localhost:9595/configuration'