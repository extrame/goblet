package goblet

import (
	"log"
	"os"
)

var LogFile *os.File

func (s *Server) initLog() {
	var err error

	if s.Config.Log.File != "" {
		if LogFile, err = os.OpenFile(s.Config.Log.File, os.O_APPEND|os.O_RDWR, 0666); err != nil {
			if os.IsNotExist(err) {
				LogFile, err = os.Create(s.Config.Log.File)
				if err != nil {
					panic(err)
				}
			}
		}
		log.Println("Change ontput to ", s.Config.Log.File)
		log.SetOutput(LogFile)
	} else {
		LogFile = os.Stderr
	}
}
