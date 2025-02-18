package logger

import (
	"github.com/sirupsen/logrus"
	"os"
)

var Log = logrus.New()

func Init() {
	// 使用文本格式化器
	Log.SetFormatter(&logrus.TextFormatter{
		// 时间格式
		TimestampFormat: "2006-01-02 15:04:05",
		// 完整时间戳
		FullTimestamp: true,
		// 禁用颜色
		DisableColors: false,
		// 强制显示完整时间戳
		ForceColors: true,
	})

	// 禁用调用者信息
	Log.ReportCaller = false

	// 设置输出到标准输出
	Log.SetOutput(os.Stdout)

	// 设置日志级别
	Log.SetLevel(logrus.InfoLevel)
}

// WithFields 添加自定义字段的辅助函数
func WithFields(fields map[string]interface{}) *logrus.Entry {
	return Log.WithFields(logrus.Fields(fields))
}
