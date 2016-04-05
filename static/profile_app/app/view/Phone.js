var phoneNumberVType = {
        phoneNumber: function(val, field){
            var phoneNumberRegex = /^7\d{10}$/;
            return phoneNumberRegex.test(val);
        },
        phoneNumberText: 'Телефон должен начинаться с 7 и иметь 10 цифр после. К примеру: 79138973664',
        phoneNumberMask: /[\d]/
    };
Ext.apply(Ext.form.field.VTypes, phoneNumberVType);

Ext.define('Console.view.Phone', {
	extend: 'Ext.window.Window',
	alias: 'widget.phoneWindow',
	title: 'Номер телефона у которого есть доступ до сего профайла',
	layout: 'fit',
	autoShow: false,
	width:300,
	height:150,
	config:{
		parent:undefined
	},
	initComponent: function() {
		console.log("init phone window");
		this.items= [{
			xtype:"form",
			items:[
			 {
				xtype: 'textfield',
				name : 'value',
				fieldLabel: 'Номер телефона',
				itemId: 'phone_value',
				width: 250,
				padding:10,
				allowBlank:false,
				vtype:'phoneNumber'
			}
			]
		}];
		this.buttons = [{
			text: 'OK',
			scope: this,
			action: 'add_phone_end'
		}];

		this.callParent(arguments);
	}
	

});