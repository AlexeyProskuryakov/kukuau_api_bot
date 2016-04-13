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

function send_messages_to_winners(){
    var winners = [],
        winners_chbx = $(".winner:checked").each(function(x, obj){
            winners.push(obj.id);
        }),
        text = $("#to-winner").val();

    $.ajax({
        type:           "POST",
        url:            "/send_messages_at_quest_end",
        data:           JSON.stringify({teams:winners, text:text, exclude:false}),
        dataType:       'json',
        success:        function(x){
                    if (x.ok == true) {
                        console.log(x);
                        text = "<div><p class='bg-success'>Сообщения для выбранных комманд поставлены в очередь на отправление.</p></div>"
                        el = Mustache.render(text, x);
                        $("#send-message-result").prepend(el);
                    }
        }
    });
}

function send_messages_to_losers(){
    var winners = [],
        winners_chbx = $(".winner:checked").each(function(x, obj){
            winners.push(obj.id);
        }),
        text = $("#to-not-winner").val();

    $.ajax({
        type:           "POST",
        url:            "/send_messages_at_quest_end",
        data:           JSON.stringify({teams:winners, text:text, exclude:true}),
        dataType:       'json',
        success:        function(x){
                    if (x.ok == true) {
                        console.log(x);
                        text = "<div><p class='bg-success'>Сообщения для комманд не входящих в выбранные поставлены в очередь на отправление.</p></div>"
                        el = Mustache.render(text, x);
                        $("#send-message-result").prepend(el);
                    }
        }
    });
}

