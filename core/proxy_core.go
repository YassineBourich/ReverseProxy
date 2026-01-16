package core

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func ForwardRequest(w http.ResponseWriter, r *http.Request, server_url url.URL) error {
	var u = "http:" + server_url.String() + r.URL.RequestURI()
	fmt.Println(u)
	req, err := http.NewRequest(r.Method, u, r.Body)

	if err != nil {
		return err
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	

	fmt.Println(res.Header.Get("HH"))
	io.Copy(w, res.Body)
	return nil
}