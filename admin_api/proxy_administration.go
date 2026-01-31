package adminapi

import (
	"fmt"
	"net/http"
)

func handle(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "hello")
}

func ProxyAdmin(port string) {
	fmt.Println("HHHHHHH")
	mux := http.NewServeMux()

	mux.HandleFunc("/", handle)

	admin_server := &http.Server{
		Addr: port,
		Handler: mux,
	}

	admin_server.ListenAndServe()
}