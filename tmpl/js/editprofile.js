$( document ).ready(function() {
    var currentString = $("textarea").val()
    $("#counter").text(currentString.length);
});

function change_country(flag, elementID){
    document.getElementById(elementID).src =
        "/img/flags/"+flag+".png";
}

function change_emailwr() {
    var x = $("#emailChallenge")[0];
    var y = $("#emailwr")[0];
    x.disabled = !y.checked;
    x.checked = y.checked;
}

function text_area_counter() {
    var currentString = $("textarea").val()
    $("#counter").text(currentString.length);
    if (currentString.length <= 500)  {
        $("#counter")[0].setAttribute('style','color:#333');
    } else {
        $("#counter")[0].setAttribute('style','color:red');
    }
}