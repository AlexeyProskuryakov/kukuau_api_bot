Ext.define('Console.view.Contact', {
	extend: 'Ext.window.Window',
	alias: 'widget.contactwindow',

	title: 'Контакт',
	layout: 'fit',
	autoShow: true,
	width:400,
	height:400,
	config:{
		parent:undefined
	},
	initComponent: function() {
		console.log("init contact window");

		this.items= [{
			xtype:"form",
			items:[
			{
				xtype: 'textfield',
				name : 'type',
				fieldLabel: 'Тип',
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
			},{
				xtype: 'textfield',
				name : 'show_description',
				fieldLabel: 'Отображаемое название',
				width: 350,
				padding:10
			}

			]
		}];
		this.buttons = [{
			text: 'Добавить',
			scope: this,
			action: 'add_contact_end'
		}];

		this.callParent(arguments);
	}
	

});