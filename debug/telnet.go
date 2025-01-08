package debug

import (
	"log"

	telnet "github.com/reiver/go-telnet"
)

func InitTelnetServer(listen string) error {
	log.Println("Telnet listen at:", listen)
	return telnet.ListenAndServe(listen, shellHandler)
}
