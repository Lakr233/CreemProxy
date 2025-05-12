package src

import (
	"log"
	"net/http"
	"os"
)

var CREEM_API_KEY = ""
var CREEM_API_HOST = "https://api.creem.io"
var SERVER_LISTEN_ADDRESS = "0.0.0.0"
var SERVER_LISTEN_PORT = "8443"
var SERVER_DATA_DIR = "/app/data/"

var (
	SSLDir         = SERVER_DATA_DIR + "server_ssl/"
	CertPath       = SSLDir + "public.key"
	PrivateKeyPath = SSLDir + "private.key"

	SigningDir            = SERVER_DATA_DIR + "signing_keys/"
	SigningPublicKeyPath  = SigningDir + "signing_public.key"
	SigningPrivateKeyPath = SigningDir + "signing_private.key"
)

var ALLOWED_PATHS = []string{
	"/v1/licenses/activate",
	"/v1/licenses/validate",
	"/v1/licenses/deactivate",
}

func TestApiKeyWithListProducts() bool {
	url := CREEM_API_HOST + "/v1/products/search"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("[-] error creating request: %s", err)
		return false
	}
	req.Header.Set("x-api-key", CREEM_API_KEY)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[-] error sending request: %s", err)
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("[-] request failed with status code: %d", resp.StatusCode)
		return false
	}
	return true
}

func PopulateEnv() {
	if os.Getenv("CREEM_API_KEY") != "" {
		CREEM_API_KEY = os.Getenv("CREEM_API_KEY")
	}
	if os.Getenv("CREEM_API_HOST") != "" {
		CREEM_API_HOST = os.Getenv("CREEM_API_HOST")
	}
	if os.Getenv("SERVER_LISTEN_ADDRESS") != "" {
		SERVER_LISTEN_ADDRESS = os.Getenv("SERVER_LISTEN_ADDRESS")
	}
	if os.Getenv("SERVER_LISTEN_PORT") != "" {
		SERVER_LISTEN_PORT = os.Getenv("SERVER_LISTEN_PORT")
	}
	if os.Getenv("SERVER_DATA_DIR") != "" {
		SERVER_DATA_DIR = os.Getenv("SERVER_DATA_DIR")
	}

	if CREEM_API_KEY == "" {
		log.Fatal("[-] CREEM_API_KEY is not set")
	}

	SSLDir = SERVER_DATA_DIR + "server_ssl/"
	CertPath = SSLDir + "public.key"
	PrivateKeyPath = SSLDir + "private.key"

	SigningDir = SERVER_DATA_DIR + "signing_keys/"
	SigningPublicKeyPath = SigningDir + "signing_public.key"
	SigningPrivateKeyPath = SigningDir + "signing_private.key"

	if TestApiKeyWithListProducts() {
		log.Print("[+] list product indicates API key is valid")
	} else {
		log.Fatal("[-] unable to validate API key")
	}
}
