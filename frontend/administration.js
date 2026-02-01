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