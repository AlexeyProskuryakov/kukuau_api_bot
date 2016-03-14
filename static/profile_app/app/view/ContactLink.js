Ext.define('Console.view.Contact', {
	extend: 'Ext.window.Window',
	alias: 'widget.ContactLinkWindow',

	title: 'Способ связи',
	layout: 'fit',
	autoShow: false,
	width:400,
	height:400,
	config:{
		parent:undefined
	},
	initComponent: function() {
		console.log("init contact window");
		var store = Ext.create('Ext.data.Store', {
			fields: ['name', 'show'],
			data:[
				{name:"phone", show:"Телефон"},
				{name:"email",show:"Электронная почта"},
				{name:"adress",show:"Адресс"},
				{name:"WWW",show:"Сайт"},
				{name:"vk",show:"Вконтачъ"},
				{name:"twitter",show:"Твиттеръ"},
				{name:"facebook",show:"Фейсбукъ"},
			]
		});
		this.items= [{
			xtype:"form",
			items:[
			{
				xtype: 'combo',
				name : 'type',
				store:store,
				queryMode: 'local',
				displayField: 'show',
				valueField: 'name',
				fieldLabel: 'Тип',
				typeAhead: true,
		        typeAheadDelay: 100,
		        hideTrigger: true,
				width: 350,
				padding:10
			}, {
				xtype: 'textfield',
				name : 'value',
				fieldLabel: 'Значение',
				width: 350,
				padding:10
			},{
				xtype: 'textfield',
				name : 'description',
				fieldLabel: 'Описание',
				width: 350,
				padding:10
			}

			]
		}];
		this.buttons = [{
			text: 'Сохранить',
			scope: this,
			action: 'add_contact_link_end'
		}];

		this.callParent(arguments);
	}
	

});