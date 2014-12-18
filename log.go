package goblet

import (
	"log"
	"os"
)

func (s *Server) initLog() {
	if *s.logFile != "" {
		if file, err := os.OpenFile(*s.logFile, os.O_APPEND|os.O_RDWR, 0666); err != nil {
			if os.IsNotExist(err) {
				file, err = os.Create(*s.logFile)
			}
			if err != nil {
				panic(err)
			}
			log.SetOutput(file)
		}
	}
}
