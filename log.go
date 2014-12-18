package goblet

import (
	"log"
)

func (s *Server) initLog() {
	if *s.logFile != "" {
		if file, err := os.OpenFile(config.Log.File, os.O_APPEND|os.O_RDWR, 0666); err != nil {
			if os.IsNotExist(err) {
				file, err = os.Create(config.Log.File)
			}
			if err != nil {
				panic(err)
			}
			log.SetOutput(file)
		}
	}
}
