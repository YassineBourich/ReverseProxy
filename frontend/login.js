const username = document.getElementById("username"),
password = document.getElementById("password"),
login_btn = document.getElementById("submit"),
eye_btn = document.getElementById("eye_icon");

var admin_endpoint = "/login";

function toggle_password() {
    if (password.type == "password") {
        password.type = "text"
        eye_btn.classList.replace("fa-eye", "fa-eye-slash");
    } else {
        password.type = "password"
        eye_btn.classList.replace("fa-eye-slash", "fa-eye");
    }
}

async function submit_credentials(e) {
    e.preventDefault();
    close_error();
    if (!validate_fields()) {
        return;
    }

    const credentials = {
        "username": username.value,
        "password": password.value,
    }

    try {
        const res = await fetch(admin_endpoint, {
            method: "POST",
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(credentials),
        });

        if (!res.ok) {
            if (res.status === 401) {
                set_error(["Wrong credentials"]);
            } else if (res.status === 500) {
                set_error(["Internal server error"]);
            } else {
                set_error(["Unknown error"]);
            }
            return;
        }

        const token = await res.text();
        sessionStorage.setItem("token", token);
        window.location.href = "/administration";
    } catch (error) {
        set_error(["Connection error: " + error]);
    }
}

eye_btn.onclick = toggle_password;

login_btn.onclick = submit_credentials;

function validate_fields() {
    err = [];
    if (username.value == "") {
        err.push("Empty username");
    }
    if (password.value == "") {
        err.push("Empty password");
    }
    
    if (err.length > 0) {
        set_error(err);
    }
    return err.length == 0;
}