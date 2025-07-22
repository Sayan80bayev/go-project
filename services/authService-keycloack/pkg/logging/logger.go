package logging

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
	"sync"
)

const (
	Reset   = "\033[0m"
	Green   = "\033[32m" // INFO & 2xx status
	Yellow  = "\033[33m" // WARN & 4xx status
	Red     = "\033[31m" // ERROR & 5xx status
	Blue    = "\033[34m" // 3xx status & POST method
	Magenta = "\033[35m" // PUT method
	Cyan    = "\033[36m" // GET method
	White   = "\033[37m" // Default
)

type CustomTextFormatter struct{}

func (f *CustomTextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var color string
	switch entry.Level {
	case logrus.InfoLevel:
		color = Green
	case logrus.WarnLevel:
		color = Yellow
	case logrus.ErrorLevel:
		color = Red
	default:
		color = Reset
	}

	logLine := fmt.Sprintf("%s%s%s %s %s",
		color, strings.ToUpper(entry.Level.String()), Reset,
		entry.Message,
		entry.Time.Format("2006-01-02 15:04:05"),
	)

	for key, value := range entry.Data {
		logLine += fmt.Sprintf(" %s=%v", key, value)
	}

	return []byte(logLine + "\n"), nil
}

var (
	logInstance *logrus.Logger
	once        sync.Once
)

func GetLogger() *logrus.Logger {
	once.Do(func() {
		logInstance = logrus.New()
		logInstance.SetOutput(os.Stdout)
		logInstance.SetFormatter(&CustomTextFormatter{})
	})
	return logInstance
}

func Middleware(c *gin.Context) {
	methodColor := getMethodColor(c.Request.Method)

	logInstance.WithFields(logrus.Fields{
		"method": fmt.Sprintf("%s%s%s", methodColor, c.Request.Method, Reset),
		"path":   c.Request.URL.Path,
	}).Info("Incoming request")

	c.Next()

	statusCode := c.Writer.Status()
	statusColor := getStatusColor(statusCode)

	logInstance.WithFields(logrus.Fields{
		"status": fmt.Sprintf("%s%d%s", statusColor, statusCode, Reset),
		"method": fmt.Sprintf("%s%s%s", methodColor, c.Request.Method, Reset),
		"path":   c.Request.URL.Path,
	}).Info("Request handled")
}

func getMethodColor(method string) string {
	switch method {
	case "GET":
		return Cyan
	case "POST":
		return Blue
	case "PUT":
		return Magenta
	case "DELETE":
		return Red
	default:
		return White
	}
}

func getStatusColor(status int) string {
	switch {
	case status >= 200 && status < 300:
		return Green
	case status >= 300 && status < 400:
		return Blue
	case status >= 400 && status < 500:
		return Yellow
	case status >= 500:
		return Red
	default:
		return White
	}
}
