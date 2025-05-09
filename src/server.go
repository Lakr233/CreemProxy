package src

import (
	"fmt"
	"log"
	"net/http"
)

func Serve() {
	http.HandleFunc("/", ProxyHandler)
	log.Printf("[*] starting server on %s:%s", SERVER_LISTEN_ADDRESS, SERVER_LISTEN_PORT)
	listenAddress := fmt.Sprintf("%s:%s", SERVER_LISTEN_ADDRESS, SERVER_LISTEN_PORT)
	if err := http.ListenAndServeTLS(listenAddress, CertPath, PrivateKeyPath, nil); err != nil {
		log.Fatalf("[-] failed to start server: %v", err)
	}
}
