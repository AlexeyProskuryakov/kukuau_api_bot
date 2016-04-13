Ext.define('Console.model.Profile', {
	extend: 'Ext.data.Model',
	fields: [
	'id', 
	'image_url', 
	'name', 
	'short_description', 
	'text_description', 
	{ name: 'enable', type:"boolean"},
	{ name: 'public', type:"boolean"}
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

