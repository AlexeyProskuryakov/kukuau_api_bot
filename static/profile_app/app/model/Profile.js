Ext.define('Console.model.Profile', {
	extend: 'Ext.data.Model',
	fields: [
		'id', 
		"image_url", 
		'name', 
		'short_description', 
		'text_description', 
		"address", 
		"place"
	],
	hasMany:{model:'Console.model.Contact', name:'contacts'}
});

Ext.define("Console.model.Contact",{
	extend:"Ext.data.Model",
	fields:[
		'id',
		'type',
		'value',
		'showed_text',
		'description'
	],
	belongsTo:"Console.model.Profile"

});