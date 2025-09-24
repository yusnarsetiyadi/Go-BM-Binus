package log

import (
	"bm_binus/internal/config"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
)

func Init() {
	logrus.SetReportCaller(true)
	logrus.SetOutput(os.Stdout)
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors:   false,
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	var level logrus.Level = logrus.InfoLevel
	if logrusLevel, _ := strconv.Atoi(config.Get().Logging.LogrusLevel); logrusLevel != 0 {
		switch logrusLevel {
		case 1:
			level = 1
		case 2:
			level = 2
		case 3:
			level = 3
		case 4:
			level = 4
		case 5:
			level = 5
		case 6:
			level = 6
		}
	}
	logrus.SetLevel(level)
}
