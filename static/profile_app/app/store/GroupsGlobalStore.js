Ext.define('Console.store.GroupsGlobalStore', {
    extend: 'Ext.data.Store',
    model: 'Console.model.Group',
    storeId: 'GroupsGlobalStore',
    proxy: {
        type: 'ajax',
        url: '/profile/all_groups',
        reader: {
            type: 'json',
            root: 'groups',
            successProperty: 'success'
        }
    },
    setActives:function(profile_store){
        var profile_groups = [];
        Ext.each(profile_store.data.items, function(item){
            console.log("profile group", item);
            profile_groups.push(item.getData()['name']);
        });
        Ext.each(this.data.items, function(item){
            console.log("global group", item);
            var itemName = item.getData().name;
            if (profile_groups.indexOf(itemName) >= 0){
                item.getData['_active'] = true;
            }
        });
    }
});

