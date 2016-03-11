Ext.define('Console.model.Profile', {
	extend: 'Ext.data.Model',
	fields: [
	'id', 
	"image_url", 
	'name', 
	'short_description', 
	'text_description', 
	"address", 
	"enable",
	"public"
	],
	hasMany:{model:'Console.model.Contact', name:'contacts'},

});

