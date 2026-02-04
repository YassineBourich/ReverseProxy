async function checkAccess() {
    const token = sessionStorage.getItem("token");

    // Immediate local check
    if (!token) {
        window.location.replace("/administration-login");
        return; // Stop execution
    }

    try {
        // Server-side validation
        const response = await fetch("/validate-token", {
            method: "GET",
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        // Handle invalid/expired token
        if (!response.ok) {
            logout();
            return;
        }

        document.querySelector("#admin_main").classList.replace("hided", "showed");
    } catch (error) {
        console.error("Network error during validation:", error);
        logout();
    }
}

checkAccess();

const add_backend_btn = document.getElementById("add_backend_btn"),
add_url_btn = document.getElementById("add_url_btn"),
url_field = document.getElementById("url_field"),
url_form = document.querySelector(".url_form"),
close_url_form_btn = document.getElementById("close_url_form"),
backends_div = document.querySelector(".backends"),
total_backends = document.querySelector(".total_backends"),
active_backends = document.querySelector(".active_backends");
var backends_status, token;

async function fetch_backends_status() {
    token = sessionStorage.getItem("token");
    try {
        const response = await fetch("/status", {
            method: "GET",
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        if (!response.ok) {
            if (response.status == 401) {
                logout();
            } else {
                console.error("Internal Server error");
            }
            return;
        }

        backends_status = await response.json();
        total_backends.innerHTML = `Total backends: ${backends_status.total_backends | 0}`;
        active_backends.innerHTML = `Active backends: ${backends_status.active_backends | 0}`;
        backends_div.innerHTML = "";
        let key = 0;
        for (let backend of backends_status.backends) {
            backends_div.insertAdjacentHTML('beforeend', get_backend_row(backend, key));
            document.querySelector(`button[key="${key}"]`).onclick = async () => {
                await remove_backend(backend.url);
            }
            key++;
        }
    } catch (error) {
        console.error("Network error during fetching:", error);
    }
}

function get_backend_row(backend, key) {
    return `
        <div>
            <span>${backend.url}</span>
            ${backend.alive ? "<span style='font-weight: bold; color: green'>Alive</span>" : "<span style='font-weight: bold; color: red'>Down</span>"}
            <span style='color: ${(backend.last_response_time / (1000 * 1000) > 999) ? "red" : "black"}'>${format_response_time(backend.last_response_time)}</span>
            <span>${backend.current_connections}</span>
            <button key="${key}"><i class="fa fa-trash"></i></button>
        </div>
    `;
}

function format_response_time(response_time) {
    if (0 < response_time && response_time < 1000) {
        return String(round_2(response_time)) + "ns";
    } else if (1000 <= response_time && response_time < 1000 * 1000) {
        return String(round_2(response_time / 1000)) + "µs";
    } else if (1000 * 1000 <= response_time && response_time < 1000 * 1000 * 1000) {
        return String(round_2(response_time / (1000 * 1000))) + "ms";
    } else {
        return String(round_2(response_time / (1000 * 1000 * 1000))) + "s";
    }
}

function round_2(num) {
    return Math.round(num * 100) / 100;
}

fetch_backends_status();
setInterval(fetch_backends_status, 1000);

function open_url_field() {
    if (url_form.classList.contains("hided")) {
        url_form.classList.replace("hided", "showed");
    }
}

function close_url_field() {
    if (url_form.classList.contains("showed")) {
        url_form.classList.replace("showed", "hided");
    }
}

add_backend_btn.onclick = () => {
    close_error();
    url_field.value = "";
    open_url_field();
}
close_url_form_btn.onclick = close_url_field;

async function send_add_backend_request(url) {
    close_error();
    token = sessionStorage.getItem("token");
    const backend = {
        url: url,
    }
    try {
        const response = await fetch("/backends", {
            method: "POST",
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(backend),
        });

        if (!response.ok) {
            if (response.status == 401) {
                logout();
            } else {
                set_error(["Internal server error"]);
            }
            return;
        }
    } catch (error) {
        console.error("Network error during adding:", error);
    }
}

add_url_btn.onclick = add_backends;

url_field.onkeydown = (e) => {
    if (e.key == 'Enter') {
        e.preventDefault();
        add_backends();
    }
}

async function add_backends() {
    urls = url_field.value.split(/[ |]+/);

    for (let i = 0; i < urls.length; i++) {
        urls[i] = urls[i].trim();
    }

    urls = urls.filter(url => url !== "");

    if (urls.length == 0) {
        set_error(["Empty URL"]);
        return;
    }

    err = [];
    for (let i = 0; i < urls.length; i++) {
        if (!URL.canParse(urls[i])) {
            err.push(urls[i] + "is not a valid URL");
        }
    }
    if (err.length > 0) {
        set_error(err);
        return;
    }

    for (let i = 0; i < urls.length; i++) {
        await send_add_backend_request(urls[i]);
    }
    close_url_field();
};

async function remove_backend(url) {
    token = sessionStorage.getItem("token");
    const backend = {
        url: url,
    }
    try {
        const response = await fetch("/backends", {
            method: "DELETE",
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(backend),
        });

        if (!response.ok) {
            if (response.status == 401) {
                logout();
            } else {
                console.error("Backend Not Found");
            }
            return;
        }
    } catch (error) {
        console.error("Network error during adding:", error);
    }
}

// Function to logout from administration
function logout() {
    sessionStorage.removeItem("token"); // Clean up the expired token
    window.location.replace("/administration-login");
}

const logout_btn = document.getElementById("logout_btn");
logout_btn.onclick = logout;