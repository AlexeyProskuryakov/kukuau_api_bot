Ext.define('Console.view.GroupChoose', {
	extend: 'Ext.window.Window',
	alias: 'widget.groupWindow',
	title: 'Группы в которые будет входить профайл',
	layout: 'fit',
	autoShow: false,
	width:600,
	height:250,
	config:{
		parent:undefined
	},
	initComponent: function() {
		console.log("init group window");
		this.items= [{
			xtype:"form",
			items:[
			{
				xtype: "grid",
				title: "Выберите группу",
				alias: "widget.groupGlobalGrid",
				itemId: "choose_group_grid",
				store: 'ProfileAllowPhoneStore',
				columns: [{
					header: "Имя группы",
					dataIndex: 'name',
					flex: 1
				}, {
					header: "Описание",
					dataIndex: 'description',
					flex: 1
				}, {
					xtype: 'actioncolumn',
					header: 'Выбрать',
					action: "choose_group",
					editor: {
						xtype: 'checkbox',
						cls: 'x-grid-checkheader-editor'
					}
				}],
				buttons: [{
					text: "Добавить новую группу",
					action: "add_global_group",
					scope: this,
				}]
			},
			]
		}];
		this.buttons = [{
			text: 'OK',
			scope: this,
			action: 'choose_groups'
		}];

		this.callParent(arguments);
	}
	

});
