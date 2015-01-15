package goblet

import (
	"log"
	"os"
)

var LogFile *os.File

func (s *Server) initLog() {
	var err error

	if *s.logFile != "" {
		if LogFile, err = os.OpenFile(*s.logFile, os.O_APPEND|os.O_RDWR, 0666); err != nil {
			if os.IsNotExist(err) {
				LogFile, err = os.Create(*s.logFile)
				if err != nil {
					panic(err)
				}
			}
		}
		log.Println("Change ontput to ", *s.logFile)
		log.SetOutput(LogFile)
	} else {
		LogFile = os.Stderr
	}
}
