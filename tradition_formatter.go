package logrus

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	KEY_SOURCE = "source"
)

func init() {
	baseTimestamp = time.Now()
}

// TraditionFormatter formats logs into text
type TraditionFormatter struct {
	// Set to true to bypass checking for a TTY before outputting colors.
	ForceColors bool

	// Force disabling colors.
	DisableColors bool

	// Override coloring based on CLICOLOR and CLICOLOR_FORCE. - https://bixense.com/clicolors/
	EnvironmentOverrideColors bool

	// Disable timestamp logging. useful when output is redirected to logging
	// system that already adds timestamps.
	DisableTimestamp bool

	// Enable logging the full timestamp when a TTY is attached instead of just
	// the time passed since beginning of execution.
	FullTimestamp bool

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

	// FieldMap allows users to customize the names of keys for default fields.
	// As an example:
	// formatter := &TextFormatter{
	//     FieldMap: FieldMap{
	//         FieldKeyTime:  "@timestamp",
	//         FieldKeyLevel: "@level",
	//         FieldKeyMsg:   "@message"}}
	FieldMap FieldMap

	// for show goroutine id
	ShowGoroutineId bool

	sync.Once
}

func (f *TraditionFormatter) init(entry *Entry) {
	if entry.Logger != nil {
		f.isTerminal = checkIfTerminal(entry.Logger.Out)
	}
}

func (f *TraditionFormatter) isColored() bool {
	isColored := f.ForceColors || f.isTerminal

	if f.EnvironmentOverrideColors {
		if force, ok := os.LookupEnv("CLICOLOR_FORCE"); ok && force != "0" {
			isColored = true
		} else if ok && force == "0" {
			isColored = false
		} else if os.Getenv("CLICOLOR") == "0" {
			isColored = false
		}
	}

	return isColored && !f.DisableColors
}

// Format renders a single log entry
func (f *TraditionFormatter) Format(entry *Entry) ([]byte, error) {

	keys := make([]string, 0, len(entry.Data))
	for k := range entry.Data {
		keys = append(keys, k)
	}

	if !f.DisableSorting {
		sort.Strings(keys)
	}

	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	f.Do(func() { f.init(entry) })

	timestampFormat := f.TimestampFormat
	if timestampFormat == "" {
		timestampFormat = defaultTimestampFormat
	}
	if f.isColored() {
		f.printColored(b, entry, keys, timestampFormat)
	} else {
		if !f.DisableTimestamp {
			f.appendKeyValue(b, f.FieldMap.resolve(FieldKeyTime), entry.Time.Format(timestampFormat))
		}
		f.appendKeyValue(b, f.FieldMap.resolve(FieldKeyLevel), entry.Level.String())

		for _, key := range keys {
			f.appendKeyValue(b, key, entry.Data[key])
		}

		if entry.Message != "" {
			f.appendKeyValue(b, f.FieldMap.resolve(FieldKeyMsg), entry.Message)
		}
	}

	b.WriteByte('\n')
	return b.Bytes(), nil
}

func (f *TraditionFormatter) printColored(b *bytes.Buffer, entry *Entry, keys []string, timestampFormat string) {
	var levelColor int
	switch entry.Level {
	case DebugLevel:
		levelColor = gray
	case WarnLevel:
		levelColor = yellow
	case ErrorLevel, FatalLevel, PanicLevel:
		levelColor = red
	default:
		levelColor = blue
	}

	levelText := strings.ToUpper(entry.Level.String())
	if !f.DisableLevelTruncation {
		levelText = levelText[0:4]
	}

	var source string
	if entry.Data[KEY_SOURCE] != nil {
		source = entry.Data[KEY_SOURCE].(string)
	}

	if f.ShowGoroutineId {
		if f.DisableTimestamp {
			fmt.Fprintf(b, "\x1b[%dm[%s] \x1b[0m [GoID:%d] %s %-44s ", levelColor, levelText, GoId(), source, entry.Message)
		} else if !f.FullTimestamp {
			fmt.Fprintf(b, "\x1b[%dm[%s] \x1b[0m[%04d] [GoID:%d] %s %-44s ", levelColor, levelText, int(entry.Time.Sub(baseTimestamp)/time.Second), GoId(), source, entry.Message)
		} else {
			fmt.Fprintf(b, "\x1b[%dm[%s] \x1b[0m[%s] [GoID:%d] %s %-44s ", levelColor, levelText, entry.Time.Format(timestampFormat), GoId(), source, entry.Message)
		}
	} else {
		if f.DisableTimestamp {
			fmt.Fprintf(b, "\x1b[%dm[%s] \x1b[0m %s %-44s ", levelColor, levelText, source, entry.Message)
		} else if !f.FullTimestamp {
			fmt.Fprintf(b, "\x1b[%dm[%s] \x1b[0m[%04d] %s %-44s ", levelColor, levelText, int(entry.Time.Sub(baseTimestamp)/time.Second), source, entry.Message)
		} else {
			fmt.Fprintf(b, "\x1b[%dm[%s] \x1b[0m[%s] %s %-44s ", levelColor, levelText, entry.Time.Format(timestampFormat), source, entry.Message)
		}
	}

	for _, k := range keys {
		if k == KEY_SOURCE {
			continue
		}
		v := entry.Data[k]
		fmt.Fprintf(b, " \x1b[%dm%s\x1b[0m=", levelColor, k)
		f.appendValue(b, v)
	}
}

func (f *TraditionFormatter) needsQuoting(text string) bool {
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

func (f *TraditionFormatter) appendKeyValue(b *bytes.Buffer, key string, value interface{}) {
	if b.Len() > 0 {
		b.WriteByte(' ')
	}
	b.WriteString(key)
	b.WriteByte('=')
	f.appendValue(b, value)
}

func (f *TraditionFormatter) appendValue(b *bytes.Buffer, value interface{}) {
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

//获取当前协程ID
func GoId() int {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(fmt.Sprintf("panic recover:panic info:%v", err))
		}
	}()

	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
}
