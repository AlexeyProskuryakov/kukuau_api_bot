
function update_keys(){
    $.ajax({
        type:"POST",
        url:            "/info_page/update",
        contentType:    'application/json',
        success:        function(result){
            if (result.ok==true){

                result.foundKeys.forEach(function(obj,i){
                    el = $("#"+obj.id)
                    if (el.length != 0){
                        if (el.attr("class").trim() == "key-not-found"){
                            el.removeClass("key-not-found");
                            el.addClass("key-found");
                        }
                    }
                });

            }
        }
    });
};

setInterval(function(){
        update_keys();
        return true;
}, 5000);
