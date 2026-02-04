const error_div = document.querySelector(".error_div");

function set_error(err) {
    if (error_div.classList.contains("hided")) {
        error_div.classList.replace("hided", "showed");
    }

    error_div.innerHTML = "<ul>";
    for (let e of err) {
        error_div.innerHTML += "<li>" + e + "</li>"
    }
    error_div.innerHTML += "</ul>"
}

function close_error() {
    if (error_div.classList.contains("showed")) {
        error_div.classList.replace("showed", "hided");
    }
}