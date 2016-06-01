$( document ).ready(function() {
    commentCounter();
    updateFormType(1);
    $("#error").css("display", "none");
    $("#working").css("display", "none");
});

function updateFormType(cat) {
    if (cat == 0) {
        $("#scorerun").css("display", "none");
        $("#speedrun").css("display", "none");
    } else if (cat == 1) {
        $("#scorerun").css("display", "");
        $("#speedrun").css("display", "none");
    } else {
        $("#scorerun").css("display", "none");
        $("#speedrun").css("display", "");
    }
    $("#inputLevel").val("4");
    if (cat == 11) {
        $("#inputWorld").val("3");
    } else if (cat == 2 || cat == 4 || cat == 5 || cat == 9 || cat == 12 || cat == 13 || cat == 15) {
        $("#inputWorld").val("4");
    } else {
        $("#inputWorld").val("5");
    }
}

function commentCounter() {
    var currentString = $("#inputComment").val()
    $("#counter").text(currentString.length);
    if (currentString.length <= 30 )  {
        $("#counter").css("color", "black");
    } else {
        $("#counter").css("color","red");
    }
}

function findResult(runType) {
    $("#working").show();
    $("#error").hide();

    $.ajax({
        url: "/steam-lookup",
        method: "POST",
        data: {runType: runType},
        dataType: 'json',
        success: function(data) {
            if (!("error" in data)) {
                console.log(data);
                var spelunker = data["SpelunkerID"];
                var world = Math.floor((data["Level"]-1)/4) + 1;
	            var floor = (data["Level"]-1)%4 + 1;
                if (runType == "score") {
                    $("#inputScore").val(data["Result"]);
                    $("#inputCategory").val("1");
                    $("#scorerun").show();
                    $("#speedrun").hide();
                } else {
                    time = data["Result"];
                    minutes = Math.floor(time/60000);
                    time = time - 60000*minutes;
                    seconds = Math.floor(time/1000);
                    millisecs = time - 1000*seconds;
                    $("#inputMinutes").val(minutes);
                    $("#inputSeconds").val(seconds);
                    $("#inputMilliseconds").val(millisecs);
                    $("#inputCategory").val("2");
                    $("#scorerun").hide();
                    $("#speedrun").show();
                }
                $("#inputLevel").val(floor);
                $("#inputWorld").val(world);
                $("#inputSpelunker").val(spelunker);
                changeSpelunker(spelunker, 'spelunker');
                $("#inputPlatform").val("1");
            } else {
                $("#error").show();
            }
            $("#working").hide();
        }
    })
}