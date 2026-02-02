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
    validate_fields();

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
            
        }

        const token = await res.text();
        sessionStorage.setItem("token", token);
        window.location.href = "/administration";
    } catch (error) {

    }
}

eye_btn.onclick = toggle_password;

login_btn.onclick = submit_credentials;

function validate_fields() {

}