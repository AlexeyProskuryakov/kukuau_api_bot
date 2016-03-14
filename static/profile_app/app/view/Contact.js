Ext.define('Console.view.Contact', {
	extend: 'Ext.window.Window',
	alias: 'widget.contactWindow',

	title: 'Контакт',
	layout: 'fit',
	autoShow: false,
	width:400,
	height:400,
	config:{
		parent:undefined
	},
	initComponent: function() {
		console.log("init contact window");
		this.items= [{
			xtype:"grid",
			title:"Способы связи",
			itemId:"profile_contact_links",
			store: 'ContactLinksStore',
			columns:[
				{header: 'Тип',  dataIndex: 'type'},
				{header: 'Значение', dataIndex: 'value', flex:1},
				{header: 'Описание', dataIndex: 'description', flex:1},
				{
					xtype : 'actioncolumn',
					header : 'Delete',
					width : 100,
					align : 'center',
					action:"delete_contact_link",
					items : [
					{
						icon:'img/delete-icon.png',
						tooltip : 'Delete',
						scope : me
					}]
				}
			]
		}];
		this.buttons = [{
			text: 'Сохранить',
			scope: this,
			action: 'add_contact_end'
		}];

		this.callParent(arguments);
	}
});