Ext.define('Console.view.ContactLink', {
	extend: 'Ext.window.Window',
	alias: 'widget.contactLinkWindow',
	title: 'Способ связи',
	layout: 'fit',
	autoShow: false,
	width:600,
	height:250,
	config:{
		parent:undefined
	},
	initComponent: function() {
		console.log("init contact link window");
		var store = Ext.create('Ext.data.Store', {
			
			fields: ['name', 'show'],
			data:[
			{name:"phone", show:"Телефон"},
			{name:"email",show:"Электронная почта"},
			{name:"www",show:"Сайт"}
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
				padding:10,
				allowBlank:false,
				itemId:'cl_type'
			}, 		{
				xtype: 'textfield',
				name : 'value',
				fieldLabel: 'Значение',
				width: 550,
				padding:10,
				allowBlank:false,
				itemId:'cl_value'
			},{
				xtype: 'textfield',
				name : 'description',
				fieldLabel: 'Описание',
				width: 550,
				padding:10
			}, {
				xtype: 'textfield',
				name : 'order_number',
				fieldLabel: 'Порядковый номер',
				width: 550,
				padding:10
			}

			]
		}];
		this.buttons = [{
			text: 'Сохранить',
			scope: this,
			action: 'save_contact_link'
		}];

		this.callParent(arguments);
	}
	

});