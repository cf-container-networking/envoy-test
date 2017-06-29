package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type Host struct {
	IPAddress string `json:"ip_address"`
	Port      int    `json:"port"`
}

func main() {
	http.HandleFunc("/v1/registration/app2", func(responseWriter http.ResponseWriter, request *http.Request) {
		var response struct {
			Hosts []Host `json:"hosts"`
		}
		response.Hosts = []Host{
			{
				IPAddress: "172.17.0.2",
				Port:      10001,
			},
		}

		err := json.NewEncoder(responseWriter).Encode(response)
		if err != nil {
			log.Printf("writing response: %s\n", err)
		}
	})

	log.Fatal(http.ListenAndServe(":4913", nil))
}
