Ext.define('Console.view.AutoAnswer', {
	extend: 'Ext.window.Window',
	alias: 'widget.answerWindow',
	title: 'Отправка автоматического сообщения пользователю',
	layout: 'fit',
	autoShow: false,
	width:600,
	height:250,
	config:{
		parent:undefined,
		profileId:undefined
	},
	initComponent: function() {
		console.log("init answer window");

		this.items = [{
			xtype:"form",
			items:[
			{
				xtype: 'numberfield',
				name : 'after_min',
				fieldLabel: 'Через (количество минут)',
				width: 250,
				padding:10,
				itemId:'after_min',
				minValue: 0,
				allowBlank:false,
			},
			{
				xtype: 'textfield',
				name : 'text',
				fieldLabel: 'Сообщение',
				width: 550,
				padding:10,
				allowBlank:false,
				itemId:'text',
			}
			]	
		}];
		this.buttons = [{
			text: 'Сохранить',
			scope: this,
			action: 'add_answer_end'	
		}];
		this.callParent(arguments);
	}
});