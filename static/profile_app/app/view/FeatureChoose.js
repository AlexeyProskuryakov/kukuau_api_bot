Ext.define('Console.view.FeatureChoose', {
	extend: 'Ext.window.Window',
	alias: 'widget.featureWindow',
	title: 'Фичи которые будут у профайла',
	layout: 'fit',
	autoShow: false,
	width:600,
	height:250,
	config:{
		parent:undefined
	},
	initComponent: function() {
		console.log("init feature window");
		this.items= [{
			xtype:"form",
			items:[
			{
				xtype: "grid",
				markDirty:false,
				title: "Выберите фичу",
				alias: "widget.featureGlobalGrid",
				itemId: "choose_feature_grid",
				store: 'FeaturesGlobalStore',
				columns: [{
					header: "Имя фичи",
					dataIndex: 'name',
					flex: 1
				}, {
					header: "Непонятное var",
					dataIndex: 'var',
					flex: 1
				}, { 
					xtype : 'checkcolumn', 
					text : 'Выбрать',
					dataIndex:'_active'
				}],
			},
			]
		}];
		this.buttons = [{
			text: 'OK',
			scope: this,
			action: 'add_feature_end'
		}
		];

		this.callParent(arguments);
	}
	

});
