function update_key_states(team_name){
    if (team_name != undefined || team_name != ""){
        $.ajax({
                type:"POST",
                url:            "/founded_keys",
                contentType:    'application/json',
                data:           JSON.stringify({team:team_name}),
                dataType:       'json',
                success:        function(x){
                    x.keys.forEach(function(k){
                        key = $("[key-id="+k.SID+"]");
                        key.removeClass("key-not-found");
                        key.addClass("key-found");
                    });

                }
        });
    }

}


setInterval(function(){
    update_key_states($("#team-name").text());
    return true;
}, 5000);

function delete_all_keys(){
     $.ajax({
        type:"POST",
        url:            "/delete_all_keys",
        success:        function(x){
            if (x.ok == true) {
                console.log(x);
                text = "<div><p class='bg-success'>Удалено шагов: {{steps_removed}}</p><p class='bg-success'>Обновленно пользователей: {{peoples_updated}}</p><p class='bg-success'>Удалено комманд: {{teams_removed}}</p><p class='bg-success'>Удалено сообщений от комманд: {{messages_removed}}</p></div>";
                el = Mustache.render(text, x);
                $("#delete-result").prepend(el);
            }
        }
    });
}

function start_quest(){
    $.ajax({
        type:"POST",
        url:"/start_quest",
        success: function(x){
            if (x.ok == true){
                window.location.reload();
            }
        }
    });
}

function stop_quest(){
    $.ajax({
        type:"POST",
        url:"/stop_quest",
        success: function(x){
            if (x.ok == true){
                window.location.reload();
            }
        }
    });
}