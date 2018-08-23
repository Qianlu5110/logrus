# logrus-plus使用说明

### 源
fork from https://github.com/sirupsen/logrus

### 效果图
![image](https://apis.loveke.xin/uploads/images/7731ad61-faab-4641-973f-f84872f73884.png)


### 引入
```
import "git.icsoc.net/paas/logrus-plus"
```

### 初始化配置
```
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
```

### 使用
```
package main

import (
	"log-test/log"
	"git.icsoc.net/paas/logrus-plus"
)

func main() {
	log.InitLog(logrus.DebugLevel)

	logrus.Infof("hello : %d", 1)
	logrus.Debugf("hello : %d", 2)
	logrus.Warnf("hello : %s", "s")
	logrus.Errorf("hello : %s", "s")
}
```

### DEMO
https://github.com/Qianlu5110/logrus/blob/master/tradition_formatter_test.go
