package setupMockServer

import (
	"log"

	smtpmock "github.com/mocktools/go-smtp-mock"
)

func MockSMTPServer() (string, int, func()) {

	server := smtpmock.New(smtpmock.ConfigurationAttr{
		LogToStdout:       true,
		LogServerActivity: true,
	})

	if err := server.Start(); err != nil {
		log.Println(err)
	}
	hostAddress, portNumber := "127.0.0.1", server.PortNumber
	stopServer := func() {
		server.Stop()
	}
	return hostAddress, portNumber, stopServer
}
