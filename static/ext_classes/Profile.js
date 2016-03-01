Ext.define('Classes.Profile', {
    config: {
            name: 'some name',
            short_info : 'short  ddd',
            text_info: 'text dddd',
    },
    constructor: function(name, short_info, text_info){
            this.initConfig();
            if(name){
                this.name = name;
            }
            if(short_info){
                this.short_info = short_info;
            }
            if(text_info){
                this.text_info = text_info;
            }
        },

    applyName: function(name){
            name = name.replace(/\s/g, '');
            if(name.length===0){
                alert('Имя не может быть равно нулю');
            }
            else{
                return name;
            }
        },

    getinfo: function() {
            alert("Профайл : " + this.name + " " + this.short_info+" "+this.text_info);
        }
    }
);