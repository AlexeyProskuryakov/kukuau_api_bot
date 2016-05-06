Ext.define("Console.model.BotConfig",{
	extend:"Ext.data.Model",
	idProperty:'id',
	fields:[
	{name:'information',type:'string'},
	{name:'id', type:'string'},
	],
	associations: [{
		type: 'hasMany',
		model: 'Console.model.TimedAnswer',
		name: 'notifications'
	}, {
		type:'hasMany',
		model:'Console.model.TimedAnswer',
		name:'answers'
	}
	],
});