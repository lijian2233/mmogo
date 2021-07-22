package logger

import (
	"fmt"
	"mmogo/lib/net"
	"os"
	"path"
	"runtime"
	"strings"
	"time"
)

type Config struct {
	Dir        string `toml:"dir" json:"dir"`
	Prefix     string `toml:"prefix" json:"prefix"`
	Level      string `toml:"level" json:"level"`
	Color      bool   `toml:"color" json:"color"`
	Terminal   bool   `toml:"terminal" json:"terminal"`
	ShowIp     bool   `toml:"show_ip" json:"show_ip"`
	TimeFormat string `toml:"time_format" json:"time_format"`
}

func DefaultConfig() *Config {
	return &Config{
		Dir:        "./logs",
		Level:      "debug",
		Color:      true,
		Terminal:   true,
		ShowIp:     true,
		TimeFormat: "2006-01-02 15:04:05",
	}
}

type Logger struct {
	conf   *Config
	writer *asyncWriter
	level  Level
	ip     string
}

func NewLogger(conf *Config) (l *Logger, err error) {
	l = &Logger{conf: conf, level: GetLevel(conf.Level)}

	if l.ip, err = net.GetLocalIp(); err != nil {
		return
	}

	l.writer, err = newWriter(conf.Dir, l.getFile)
	return
}

func (l *Logger) getFile() string {
	return path.Join(l.conf.Dir, l.conf.Prefix+time.Now().Format("20060102.log"))
}

func (l *Logger) prefix(level Level, file string, line int) string {
	nowTime := time.Now().Format(l.conf.TimeFormat)
	levelText := GetLevelText(level, l.conf.Color)
	loc := fmt.Sprintf("<%s:%d>", file, line)

	if l.conf.Color {
		loc = Blue(loc)
	}

	if l.conf.ShowIp {
		return fmt.Sprintf("%s (%s) %s %s ", levelText, l.ip, nowTime, loc)
	} else {
		return fmt.Sprintf("%s %s %s ", levelText, nowTime, loc)
	}
}

func (l *Logger) getFileInfo() (file string, line int) {
	_, file, line, ok := runtime.Caller(3)

	if !ok {
		return "???", 1
	}

	if dirs := strings.Split(file, "/"); len(dirs) >= 2 {
		return dirs[len(dirs)-2] + "/" + dirs[len(dirs)-1], line
	}

	return
}

func (l *Logger) log(level Level, format string, args ...interface{}) {
	file, line := l.getFileInfo()
	prefix := l.prefix(level, file, line)
	msg := &message{prefix: prefix, format: format, args: args}

	if l.conf.Terminal {
		_, _ = fmt.Fprint(os.Stdout, string(l.writer.bytes(msg)))
		l.writer.write(msg)
	} else {
		l.writer.write(msg)
	}
}

func (l *Logger) Write(p []byte) (n int, err error) {
	msg := &message{args: []interface{}{string(p)}, ignoreLF: true}

	if l.conf.Terminal {
		n, err = fmt.Fprint(os.Stdout, string(l.writer.bytes(msg)))
		l.writer.write(msg)
	} else {
		l.writer.write(msg)
		n = len(p)
	}

	return
}

func (l *Logger) Debug(args ...interface{}) {
	l.log(DebugLevel, "", args...)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.log(DebugLevel, format, args...)
}

func (l *Logger) Info(args ...interface{}) {
	l.log(InfoLevel, "", args...)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.log(InfoLevel, format, args...)
}

func (l *Logger) Warning(args ...interface{}) {
	l.log(WarnLevel, "", args...)
}

func (l *Logger) Warningf(format string, args ...interface{}) {
	l.log(WarnLevel, format, args...)
}

func (l *Logger) Error(args ...interface{}) {
	l.log(ErrorLevel, "", args...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log(ErrorLevel, format, args...)
}

func (l *Logger) Fatal(args ...interface{}) {
	l.log(FatalLevel, "", args...)
}

func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.log(FatalLevel, format, args...)
}

func (l *Logger) Config() *Config {
	return l.conf
}

func (l *Logger) Close() {
	l.writer.close()
}

