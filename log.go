package goblet

import (
	"log/slog"
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
		slog.Info("Change log output to file", "file", s.Config.Log.File)
		// 配置slog输出到文件
		handler := slog.NewJSONHandler(LogFile, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
		slog.SetDefault(slog.New(handler))
	} else {
		LogFile = os.Stderr
	}
}
