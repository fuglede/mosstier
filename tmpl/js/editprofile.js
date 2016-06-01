$( document ).ready(function() {
    textAreaCounter();
    changeEmailwr();
});

function changeCountry(flag, elementID){
    document.getElementById(elementID).src =
        "/img/flags/"+flag+".png";
}

function changeEmailwr() {
    var x = $("#challengewrCheckbox")[0];
    var y = $("#newwrCheckbox")[0];
    x.disabled = !y.checked;
    x.checked = y.checked;
}

function textAreaCounter() {
    var currentString = $("textarea").val()
    $("#counter").text(currentString.length);
    if (currentString.length <= 500)  {
        $("#counter").css("color", "#333");
    } else {
        $("#counter").css("color", "red");
    }
}