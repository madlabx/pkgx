package log

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	nocolor = 0
	red     = 31
	green   = 32
	yellow  = 33
	blue    = 36
	gray    = 37
)

var (
	baseTimestamp time.Time
	emptyFieldMap logrus.FieldMap
)

// const defaultTimestampFormat = time.RFC3339
const defaultTimestampFormat = "2006/01/02 15:04:05.000"

func init() {
	baseTimestamp = time.Now()
}

// TextFormatter formats logs into text
type TextFormatter struct {
	// Set to true to bypass checking for a TTY before outputting colors.
	ForceColors bool

	// Force disabling colors.
	DisableColors bool

	// Disable timestamp logging. useful when output is redirected to logging
	// system that already adds timestamps.
	DisableTimestamp bool

	// TimestampFormat to use for display when a full timestamp is printed
	TimestampFormat string

	// The fields are sorted by default for a consistent output. For applications
	// that log extremely frequently and don't use the JSON formatter this may not
	// be desired.
	DisableSorting bool

	// Disables the truncation of the level text to 4 characters.
	DisableLevelTruncation bool

	// QuoteEmptyFields will wrap empty fields in quotes if true
	QuoteEmptyFields bool

	// Whether the logger's out is to a terminal
	isTerminal bool

	// Show Field Keys
	// true:  time="2020-02-22T21:33:31+08:00" level=info msg="File read done:conf/input!"
	// false: "2020-02-22T21:42:57+08:00" info "Start all services..."
	EnableFieldKey bool

	// show file:line
	DisableFileLine bool

	// need quote
	EnableQuoting bool

	sync.Once
}

func (f *TextFormatter) init(entry *logrus.Entry) {
	if entry.Logger != nil {
		f.isTerminal = checkIfTerminal(entry.Logger.Out)
	}

	if f.TimestampFormat == "" {
		f.TimestampFormat = defaultTimestampFormat
	}
}

func getRunTimeInfo(frame int) (file, fName string, line int, ok bool) {
	frame += 1 //skip this function call
	_, file, line, ok = runtime.Caller(frame)
	if ok {
		if idx := strings.LastIndex(file, "/"); idx >= 0 {
			file = file[idx+1:]
		}
		/*
			fName = runtime.FuncForPC(pc).Name()
			if idx := strings.LastIndex(fName, "/"); idx >= 0 {
				fName = fName[idx+1:]
			}
		*/

	}
	return
}

func getRunTimeInfoString(frame int) (string, bool) {
	frame += 1 //skip this function call
	if file, _, line, ok := getRunTimeInfo(frame); ok {
		//return fmt.Sprintf("%s:%d(%s)", file, line, fName), true
		return fmt.Sprintf("%s:%d", file, line), true
	}
	return "", false
}

func checkIfTerminal(w io.Writer) bool {
	switch v := w.(type) {
	case *os.File:
		return terminal.IsTerminal(int(v.Fd()))
	default:
		return false
	}
}

// Convert the Level to a string. E.g. PanicLevel becomes "panic".
func LevelToString(level logrus.Level) string {
	switch level {
	case logrus.DebugLevel:
		return "DEBU"
	case logrus.InfoLevel:
		return "INFO"
	case logrus.WarnLevel:
		return "WARN"
	case logrus.ErrorLevel:
		return "ERRO"
	case logrus.FatalLevel:
		return "FATA"
	case logrus.PanicLevel:
		return "PANI"
	}

	return "UNKN"
}

// Format renders a single log entry
func (f *TextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	keys := make([]string, 0, len(entry.Data))
	for k := range entry.Data {
		keys = append(keys, k)
	}

	if !f.DisableSorting {
		sort.Strings(keys)
	}
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	f.Do(func() { f.init(entry) })

	isColored := (f.ForceColors || f.isTerminal) && !f.DisableColors

	levelStr := LevelToString(entry.Level)
	if isColored {
		levelStr = f.withColored(levelStr, entry)
	}

	if !f.DisableTimestamp {
		f.appendMsg(b, "time", entry.Time.Format(f.TimestampFormat))
	}
	f.appendMsg(b, "level", levelStr)
	if !f.DisableFileLine {
		fl, _ := getRunTimeInfoString(9)
		f.appendMsg(b, "filen", fl)
	}
	if entry.Message != "" {
		f.appendMsg(b, "msg", entry.Message)
	}
	for _, key := range keys {
		f.appendKeyValueItf(b, key, entry.Data[key])
	}

	b.WriteByte('\n')
	return b.Bytes(), nil
}
func (f *TextFormatter) withColored(str string, entry *logrus.Entry) string {
	var levelColor int
	switch entry.Level {
	case logrus.DebugLevel:
		levelColor = gray
	case logrus.WarnLevel:
		levelColor = yellow
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		levelColor = red
	default:
		levelColor = blue
	}

	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", levelColor, str)
}

func (f *TextFormatter) needsQuoting(text string) bool {
	if !f.EnableQuoting {
		return false
	}
	if f.QuoteEmptyFields && len(text) == 0 {
		return true
	}
	for _, ch := range text {
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '-' || ch == '.' || ch == '_' || ch == '/' || ch == '@' || ch == '^' || ch == '+') {
			return true
		}
	}
	return false
}

func (f *TextFormatter) appendMsg(b *bytes.Buffer, key string, value string) {
	if f.EnableFieldKey {
		f.appendKeyValue(b, key, value)
	} else {
		if b.Len() > 0 {
			b.WriteByte(' ')
		}
		f.appendValue(b, value)
	}
}

func (f *TextFormatter) appendKeyValue(b *bytes.Buffer, key string, value string) {
	if b.Len() > 0 {
		b.WriteByte(' ')
	}
	b.WriteString(key)
	b.WriteByte('=')
	f.appendValue(b, value)
}

func (f *TextFormatter) appendValue(b *bytes.Buffer, value string) {
	if !f.needsQuoting(value) {
		b.WriteString(value)
	} else {
		b.WriteString(fmt.Sprintf("%q", value))
	}
}

func (f *TextFormatter) appendKeyValueItf(b *bytes.Buffer, key string, value interface{}) {
	if b.Len() > 0 {
		b.WriteByte(' ')
	}
	b.WriteString(key)
	b.WriteByte('=')
	f.appendValueItf(b, value)
}

func (f *TextFormatter) appendValueItf(b *bytes.Buffer, value interface{}) {
	stringVal, ok := value.(string)
	if !ok {
		stringVal = fmt.Sprint(value)
	}

	if !f.needsQuoting(stringVal) {
		b.WriteString(stringVal)
	} else {
		b.WriteString(fmt.Sprintf("%q", stringVal))
	}
}
