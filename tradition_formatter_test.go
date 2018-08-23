package logrus

import (
	"github.com/Qianlu5110/logrus"
	"os"
	"testing"
)

func InitLog(level logrus.Level) {
	traditionFormatter := new(logrus.TraditionFormatter)
	traditionFormatter.ForceColors = true
	traditionFormatter.FullTimestamp = true
	traditionFormatter.DisableLevelTruncation = false
	traditionFormatter.TimestampFormat = "2006-01-02 15:04:05.000000000"
	logrus.AddHook(logrus.FileContextHook())
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(level)
	logrus.SetFormatter(traditionFormatter)
}

func TestTraditionFormatting(t *testing.T) {
	InitLog(logrus.DebugLevel)

	logrus.Infof("hello : %d", 1)
	logrus.Debugf("hello : %d", 2)
	logrus.Warnf("hello : %s", "s")
	logrus.Errorf("hello : %s", "s")
}
