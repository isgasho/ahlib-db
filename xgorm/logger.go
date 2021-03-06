package xgorm

import (
	"database/sql/driver"
	"fmt"
	"github.com/sirupsen/logrus"
	"log"
	"reflect"
	"regexp"
	"strings"
	"time"
)

// SilenceLogger can hide SQL logs and INFO logs.
// Tips: Default `SetLogMode(false)` will hide only SQL log.
type SilenceLogger struct{}

// NewSilenceLogger creates a new SilenceLogger.
func NewSilenceLogger() *SilenceLogger {
	return &SilenceLogger{}
}

func (g *SilenceLogger) Print(...interface{}) {}

// LogrusLogger logs SQL and INFO message using logrus.Logger.
type LogrusLogger struct {
	logger *logrus.Logger
}

// NewLogrusLogger creates a new LogrusLogger with logrus.Logger.
func NewLogrusLogger(logger *logrus.Logger) *LogrusLogger {
	return &LogrusLogger{logger: logger}
}

// See gorm.LogFormatter for details.
func (g *LogrusLogger) Print(v ...interface{}) {
	if len(v) <= 1 {
		return
	}

	// info
	if len(v) == 2 {
		g.logger.WithFields(logrus.Fields{
			"module": "gorm",
			"type":   v[0],
			"info":   v[1],
		}).Infof(fmt.Sprintf("[Gorm] %v", v[1]))
		return
	}

	// sql
	if v[0] == "sql" {
		source := v[1]
		duration := v[2]
		sql := render(v[3].(string), v[4])
		rows := v[5]
		g.logger.WithFields(logrus.Fields{
			"module":   "gorm",
			"type":     "sql",
			"source":   source,
			"duration": duration,
			"sql":      sql,
			"rows":     rows,
		}).Info(fmt.Sprintf("[Gorm] #: %4d | %12s | %s | %s", rows, duration, sql, source))
		return
	}

	// other
	msg := fmt.Sprint(v[2:]...)
	g.logger.WithFields(logrus.Fields{
		"module":  "gorm",
		"type":    v[0],
		"message": msg,
	}).Info(fmt.Sprintf("[Gorm] [%s] %s", v[0], msg))
}

// StdLogLogger logs SQL and INFO message using log.Logger.
type StdLogLogger struct {
	logger *log.Logger
}

// NewStdLogLogger creates a new StdLogLogger with log.Logger.
func NewStdLogLogger(logger *log.Logger) *StdLogLogger {
	return &StdLogLogger{logger: logger}
}

func (g *StdLogLogger) Print(v ...interface{}) {
	if len(v) <= 1 {
		return
	}

	if len(v) == 2 {
		g.logger.Printf("[Gorm] %v", v[1])
		return
	}

	if v[0] == "sql" {
		source := v[1]
		duration := v[2]
		sql := render(v[3].(string), v[4])
		rows := v[5]
		g.logger.Printf("[Gorm] #: %4d | %12s | %s | %s", rows, duration, sql, source)
		return
	}

	g.logger.Printf("[Gorm] %s", fmt.Sprint(v[2:]...))
}

var sqlRegexp = regexp.MustCompile(`(\$\d+)|\?`)

func render(sql string, param interface{}) string {
	values := make([]interface{}, 0)
	for _, value := range param.([]interface{}) {
		indirectValue := reflect.Indirect(reflect.ValueOf(value))
		if indirectValue.IsValid() { // valid
			value = indirectValue.Interface()
			if t, ok := value.(time.Time); ok { // time
				values = append(values, fmt.Sprintf("'%v'", t.Format(time.RFC3339)))
			} else if b, ok := value.([]byte); ok { // bytes
				values = append(values, fmt.Sprintf("'%v'", string(b)))
			} else if r, ok := value.(driver.Valuer); ok { // driver
				if value, err := r.Value(); err == nil && value != nil {
					values = append(values, fmt.Sprintf("'%v'", value))
				} else {
					values = append(values, "NULL")
				}
			} else { // other value
				values = append(values, fmt.Sprintf("'%v'", value))
			}
		} else { // invalid
			values = append(values, fmt.Sprintf("'%v'", value))
		}
	}

	result := fmt.Sprintf(sqlRegexp.ReplaceAllString(sql, "%v"), values...)
	return strings.TrimSpace(result)
}
