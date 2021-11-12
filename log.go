package goblet

import (
	"log"
	"os"
)

var LogFile *os.File

func (s *Server) initLog() {
	var err error

	if s.Log.File != "" {
		if LogFile, err = os.OpenFile(s.Log.File, os.O_APPEND|os.O_RDWR, 0666); err != nil {
			if os.IsNotExist(err) {
				LogFile, err = os.Create(s.Log.File)
				if err != nil {
					panic(err)
				}
			}
		}
		log.Println("Change ontput to ", s.Log.File)
		log.SetOutput(LogFile)
	} else {
		LogFile = os.Stderr
	}
}
