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
            sessionStorage.removeItem("token"); // Clean up the bad token
            window.location.replace("/administration-login");
            return;
        }
    } catch (error) {
        console.error("Network error during validation:", error);
    }
}

checkAccess();

const add_backend_btn = document.getElementById("add_backend_btn"),
add_url_btn = document.getElementById("add_url_btn"),
url_field = document.getElementById("url_field"),
url_form = document.getElementById("url_form"),
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
                sessionStorage.removeItem("token"); // Clean up the bad token
                window.location.replace("/administration-login");
                return;
            } else {

            }
        }

        backends_status = await response.json();
        total_backends.innerHTML = backends_status.total_backends | 0;
        active_backends.innerHTML = backends_status.active_backends | 0;
        backends_div.innerHTML = "";
        for (let backend of backends_status.backends) {
            backends_div.innerHTML += "<div>" + backend.url + " | " + backend.alive + " | " + backend.current_connections + " | " + backend.last_response_time + "<button id='" + backend.url + "'>remove</button></div>";
            document.getElementById(backend.url).onclick = async () => {
                console.log("Must be removed" + backend.url);
                await remove_backend(backend.url);
            }
        }
    } catch (error) {
        console.error("Network error during fetching:", error);
    }
}

setInterval(fetch_backends_status, 1000);

async function add_backend(url) {
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
                sessionStorage.removeItem("token"); // Clean up the bad token
                window.location.replace("/administration-login");
                return;
            } else {

            }
        }

        if (response.status == 201) {
            console.log("YAY!");
        }
    } catch (error) {
        console.error("Network error during adding:", error);
    }
}

add_url_btn.onclick = async () => {
    await add_backend(url_field.value);
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
                sessionStorage.removeItem("token"); // Clean up the bad token
                window.location.replace("/administration-login");
                return;
            } else {

            }
        }

        if (response.status == 204) {
            console.log("YAY!");
        }
    } catch (error) {
        console.error("Network error during adding:", error);
    }
}