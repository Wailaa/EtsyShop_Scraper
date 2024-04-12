package setupMockServer

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
)

var MockServer *httptest.Server

func GlobalTestSetupMockServer(file string) {
	fileContent := []byte{}

	if file != "" {

		fileContent, _ = os.ReadFile(file)
		// if err != nil {
		// 	log.Fatalf("Failed to read test data file: %v", err)
		// }
	}

	MockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		log.Println("mocked server running now........")
		_, err := w.Write(fileContent)
		if err != nil {
			log.Fatalf("Failed to write response body: %v", err)
		}
	}))

}
