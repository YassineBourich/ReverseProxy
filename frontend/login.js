const username = document.getElementById("username"),
password = document.getElementById("password"),
login_btn = document.getElementById("submit"),
eye_btn = document.getElementById("eye_icon");

var login_endpoint = "/login";

// Function to show or hide the password in the password field
function toggle_password() {
    if (password.type == "password") {
        password.type = "text"
        eye_btn.classList.replace("fa-eye", "fa-eye-slash");
    } else {
        password.type = "password"
        eye_btn.classList.replace("fa-eye-slash", "fa-eye");
    }
}

// Function to submit the login credentials to login_endpoint
async function submit_credentials(e) {
    e.preventDefault();
    // Validating the input
    close_error();
    if (!validate_fields()) {
        return;
    }

    // building a json Object
    const credentials = {
        "username": username.value,
        "password": password.value,
    }

    try {
        // Sending a POST request with credentials
        const res = await fetch(login_endpoint, {
            method: "POST",
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(credentials),
        });

        // Resolving errors if any
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

        // If no error, store the token in the browser and redirect to administration page
        const token = await res.text();
        sessionStorage.setItem("token", token);
        window.location.href = "/administration";
    } catch (error) {
        set_error(["Connection error: " + error]);
    }
}

// click events definition
eye_btn.onclick = toggle_password;
login_btn.onclick = submit_credentials;

// Fields validation checks if fields are empty
function validate_fields() {
    err = [];
    if (username.value == "") {
        err.push("Empty username");
    }
    if (password.value == "") {
        err.push("Empty password");
    }
    
    // If a field is empty show the error
    if (err.length > 0) {
        set_error(err);
    }
    // return true if no error
    return err.length == 0;
}