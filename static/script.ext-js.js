Ext.onReady(function(){
        Ext.create('Ext.container.Viewport', {
            layout: 'fit',
            items: [
                {
                    title: 'Добавление профиля',
                    html : 'Вот так'
                }
            ]
        });
    });

Ext.define('Classes.Profile', {
            name: 'name',
            short_info : 'short',
            text_info: 'text',

            config: {
                name: 'some name',
                short_info : 'short  ddd',
                text_info: 'text dddd',
            },
            applyName: function(name){
                // удаляем из имени все пробелы
                name = name.replace(/\s/g, '');
                // если после этого в имени ничего не осталось
                // выбрасываем ошибку
                if(name.length===0){
                    alert('Имя не может быть равно нулю');
                }
                else{
                    return name;
                }
            },
            constructor: function(config) {
                this.initConfig(config);
            },
            getinfo: function() {
                alert("Профайл : " + this.name + " " + this.short_info+" "+this.text_info);
            }
});

var profile1 = Ext.create('Classes.Profile', "Foo", "Bar");
profile1.getinfo();



var pre = Ext.Class.getDefaultPreprocessors(); // получаем предобработчики
var post = Ext.ClassManager.defaultPostprocessors; // постобработчики
console.log(pre);
console.log(post);