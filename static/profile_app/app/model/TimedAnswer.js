Ext.define("Console.model.TimedAnswer",{
	extend:"Ext.data.Model",
	idProperty:'_id',
	fields:[
	"_id",
	{name:'after_min',type:'int'},
	{name:'text', type:'string'}
	]
});