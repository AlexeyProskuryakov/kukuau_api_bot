Ext.Loader.setConfig({
    enabled: true,
    paths: { 'Classes' : 'ext_classes' }
});

Ext.require('Classes.Profile');

Ext.define('Person.Panel', {
        alias: 'widget.profile_panel',
        extend: 'Ext.panel.Panel',
        title: 'Персональная панель',
        html : '<a href="/">Домой</a>'
});

Ext.onReady(function(){
//    Ext.create('Ext.container.Viewport', {
//        layout: 'fit',
//        items: [{
//            xtype:"profile_panel"
//        }]
//    });
    Ext.get("profile-photo-img").on({
        click:function(e,t,o){
            Ext.select("div .contact").setStyle("color","red");
            Ext.select("div .contact-link").setStyle("color","blue");

            var form = Ext.get("profile-form");

            Ext.DomHelper.append(form, {tag:"h2", html:"Пыщь пыщь"});
            console.log(form.child("div"));
            Ext.DomHelper.insertBefore(form.child("div"), {tag:"h2", html:"Пыщь пыщь до"});
            Ext.DomHelper.insertAfter(form.child("div"), {tag:"h2", html:"Пыщь пыщь после"});
        },
        contextmenu:function(e,t,o){
            var form = Ext.get("profile-form");

            Ext.DomHelper.append(form, {tag:"h2", html:"Пыщь пыщь"});
            console.log(form.child("div"));
            Ext.DomHelper.insertBefore(form.child("div"), {tag:"h2", html:"Пыщь пыщь до"});
            Ext.DomHelper.insertAfter(form.child("div"), {tag:"h2", html:"Пыщь пыщь после"});
        }
    });

    Ext.create("Ext.Button", {
        margin:"10 0 30",
        text:"Click here!!!",
        renderTo:Ext.getBody(),
        listeners:{
            click:function(e,t,o){
                console.log(e);
                console.log(t);
                console.log(o);
            },
            scope:this
        }
    });

    var element = Ext.get('content');
    element.on('click', function(e, target, options){
        Ext.DomHelper.append(element, {tag:"h3", html:"Ебта хуёбта"});
    }, this, {
        delegate: '.text'
    });


    Ext.onReady(function(){
    var element = Ext.get('menu');
    element.on('click', function(e, target, options){
            if(e.getTarget('li .buy')) {
                console.log('Покупка');
            }
            else if(e.getTarget('li .sell')) {
                console.log('Продажа');
            }
            else if(e.getTarget('li .exit')) {
                console.log('выход');
            }
        }, this, {
            delegate: 'li'
        });
    });

    Ext.create('Ext.panel.Panel', {
        renderTo: Ext.getBody(),
        width: 300,
        height: 230,
        padding:10,
        title: 'Основной контейнер',
        layout: 'fit',
        items: {
                title: 'Внутренняя панель',
                html: 'Внутренняя панель при Fit Layout',
                padding: 20,
                border: true
            }
    });

   Ext.create('Ext.panel.Panel', {
        renderTo: Ext.getBody(),
        width: 300,
        height: 330,
        padding:10,
        title: 'Приложение Ext JS 4',
        layout: {
                type: 'vbox',
                align: 'stretch'
            },
            items: [{
                    xtype: 'panel',
                    title: 'Первая панель',
                    padding:10,
                    height:100
                },{
                    xtype: 'panel',
                    title: 'Вторая панель',
                    height:80
                },{
                    xtype: 'panel',
                    title: 'Третья панель',
                    height:100
                }]
    });

   Ext.create('Ext.panel.Panel', {
        renderTo: Ext.getBody(),
        width: 400,
        height: 200,
        padding:10,
        title: 'hbox',
        layout: {
                type: 'hbox',
                align: 'stretch',
                pack:'end'
            },
        items: [{
                xtype: 'panel',
                title: 'Первая панель',
                width:120
            },{
                xtype: 'panel',
                title: 'Вторая панель',
                width:140
            },{
                xtype: 'panel',
                title: 'Третья панель',
                width:120
            }]
    });
    Ext.create('Ext.Panel', {
            title: 'Таблица хуеблица',
            width: 500,
            height: 100,
            padding: 10,
            layout:'column',
            items: [
                {
                    xtype: 'panel',
                    title: 'Первый столбец',
                    html: 'Поле 1',
                    width: 100
                },
                {
                    xtype: 'panel',
                    title: 'Второй столбец',
                    html: 'Поле 2',
                    columnWidth:.4
                },
                {
                    xtype: 'panel',
                    title: 'Третий столбец',
                    html: 'Поле 3',
                    columnWidth:.6
                }],
            renderTo: Ext.getBody()
        });

});

