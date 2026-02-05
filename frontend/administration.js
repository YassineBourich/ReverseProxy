// Function to check if the user have the right to access administration page
async function checkAccess() {
    const token = sessionStorage.getItem("token");

    // Immediate local check
    if (!token) {
        window.location.replace("/administration-login");
        return;
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

        // If the token exist and it is valid, show the main content
        document.querySelector("#admin_main").classList.replace("hided", "showed");
    } catch (error) {
        // If a network error happens, then logout
        console.error("Network error during validation:", error);
        logout();
    }
}

checkAccess();

// Getting DOM elements
const add_backend_btn = document.getElementById("add_backend_btn"),
add_url_btn = document.getElementById("add_url_btn"),
url_field = document.getElementById("url_field"),
url_form = document.querySelector(".url_form"),
close_url_form_btn = document.getElementById("close_url_form"),
backends_div = document.querySelector(".backends"),
total_backends = document.querySelector(".total_backends"),
active_backends = document.querySelector(".active_backends");
var backends_status, token;

//__________________________________Using /status endpoint__________________________________
async function fetch_backends_status() {
    // getting the token
    token = sessionStorage.getItem("token");
    try {
        // Send a GET request to the endpoint /status with the authorization token
        const response = await fetch("/status", {
            method: "GET",
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        if (!response.ok) {
            // If status code is 401, then the token is not valid or expired
            if (response.status == 401) {
                logout();
            } else {
                console.error("Internal Server error");
            }
            return;
        }

        // If the response is ok, populate the fields with corresponding values
        backends_status = await response.json();
        total_backends.innerHTML = `Total backends: ${backends_status.total_backends | 0}`;
        active_backends.innerHTML = `Active backends: ${backends_status.active_backends | 0}`;
        backends_div.innerHTML = "";
        let key = 0;
        for (let backend of backends_status.backends) {
            backends_div.insertAdjacentHTML('beforeend', get_backend_row(backend, key));
            // Each row have its own delete button identified by a unique key, and calls remove_backend function
            document.querySelector(`button[key="${key}"]`).onclick = async () => {
                await remove_backend(backend.url);
            }
            key++;
        }
    } catch (error) {
        console.error("Network error during fetching:", error);
    }
}

// Function to dynamically render a backend status row in HTML
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

// Function to get the response time with appropriate unit of time (ns, µs, ms, s)
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
// Rounding a decimal to digits after the decimal point
function round_2(num) {
    return Math.round(num * 100) / 100;
}

// Fetching the status every second
fetch_backends_status();
setInterval(fetch_backends_status, 1000);

// Open and close URL pop-up
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


//__________________________________Using /backends endpoint with method POST__________________________________
async function send_add_backend_request(url) {
    close_error();
    // Getting the token
    token = sessionStorage.getItem("token");
    // Defining the backend to add
    const backend = {
        url: url,
    }
    try {
        // Sending a POST request to /backends with authorization token
        const response = await fetch("/backends", {
            method: "POST",
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(backend),
        });

        if (!response.ok) {
            // If status code is 401, then the token is not valid or expired
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

// Add one or many backends separated by space or |
async function add_backends() {
    // Splitting the input
    urls = url_field.value.split(/[ |]+/);

    // trim to remove spaces
    for (let i = 0; i < urls.length; i++) {
        urls[i] = urls[i].trim();
    }

    // Filter empty urls
    urls = urls.filter(url => url !== "");

    // Check if the input is empty
    if (urls.length == 0) {
        set_error(["Empty input"]);
        return;
    }

    err = [];
    // Check if an input is not a valid URL
    for (let i = 0; i < urls.length; i++) {
        if (!URL.canParse(urls[i])) {
            err.push(urls[i] + "is not a valid URL");
        }
    }
    // Show errors if there exist
    if (err.length > 0) {
        set_error(err);
        return;
    }

    // Send the 'add backend request' to the endpoint
    for (let i = 0; i < urls.length; i++) {
        await send_add_backend_request(urls[i]);
    }
    close_url_field();
}

// Call add_backends if click on add_url_btn or press enter on url field
add_url_btn.onclick = add_backends;

url_field.onkeydown = (e) => {
    if (e.key == 'Enter') {
        e.preventDefault();
        add_backends();
    }
}


//__________________________________Using /backends endpoint with method DELETE__________________________________
async function remove_backend(url) {
    // Getting the token
    token = sessionStorage.getItem("token");
    // Defining the backend to remove
    const backend = {
        url: url,
    }
    try {
        // Sending a DELETE request to /backends with authorization token
        const response = await fetch("/backends", {
            method: "DELETE",
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(backend),
        });

        if (!response.ok) {
            // If status code is 401, then the token is not valid or expired
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

// Define the logout event when clicking the logout_btn
const logout_btn = document.getElementById("logout_btn");
logout_btn.onclick = logout;