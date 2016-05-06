Ext.define('Console.model.Profile', {
	extend: 'Ext.data.Model',
	fields: [
	'id', 
	'image_url', 
	'name', 
	'short_description', 
	'text_description', 
	{ name: 'enable', type:"boolean"},
	{ name: 'public', type:"boolean"},
	'botconfig'
	],
	associations: [{
		type: 'hasMany',
		model: 'Console.model.Contact',
		name: 'contacts'
	}, {
		type:'hasMany',
		model:'Console.model.Group',
		name:'groups'
	}, {
		type:'hasMany', 
		model:'Console.model.ProfileAllowPhone',
		name:'phones'
	},
	{
		type:'hasMany',
		model:'Console.model.Feature',
		name:'features'
	},
	{
		type:'hasMany',
		model:'Console.model.Employee',
		name:'employees'
	},
	{
		type:'hasOne',
		model:'Console.model.BotConfig',
		//associatedModel:'Console.model.BotConfig',
		name:'botconfig',
		getterName:'getBotConfig',
		setterName:'setBotConfig',
	}
	],
	proxy: {
		type: 'ajax',
		api: {
			read: '/profile/read',
			create: '/profile/create',
			update: '/profile/update',
			destroy: '/profile/delete'
		}
	}
});

