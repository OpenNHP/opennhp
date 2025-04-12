package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	LogQueueSize       = 1024
	ShowCallerFileLine = true
)

// Log levels for use with NewLogger.
const (
	LogLevelSilent = iota
	LogLevelError
	LogLevelInfo
	LogLevelAudit
	LogLevelDebug
	LogLevelTrace
)

type AsyncLogWriter struct {
	sync.Mutex
	wg sync.WaitGroup

	DirPath  string
	Name     string
	currDate string

	dateUpdatedCh chan string
	msg           chan []byte
}

func (lw *AsyncLogWriter) Start() {
	lw.currDate = time.Now().Format("2006-01-02")

	if lw.msg == nil {
		if len(lw.DirPath) > 0 {
			err := os.MkdirAll(lw.DirPath, os.ModePerm)
			if err != nil {
				fmt.Printf("Warning: AsyncLogWriter cannot create directory %s (%v). Using current working directory instead.\n", lw.DirPath, err)
				lw.DirPath = ""
			}
		}

		lw.msg = make(chan []byte, LogQueueSize)
		lw.wg.Add(1)
		go lw.writeRoutine()
	}
}

// async writer
// must be initiated first, implements atomic write
func (lw *AsyncLogWriter) Write(buf []byte) (n int, err error) {
	lw.Lock()
	defer lw.Unlock()

	len := len(buf)
	msg := make([]byte, len)
	copy(msg, buf)
	lw.msg <- msg

	return len, nil
}

func (lw *AsyncLogWriter) writeRoutine() {
	defer lw.wg.Done()

	useStdout := len(lw.DirPath) == 0 && len(lw.Name) == 0

	for {
		var quit bool
		var err error
		msgArr := make([][]byte, 0, LogQueueSize)

		select {
		// block to wait for incoming log messages
		case msg := <-lw.msg:
			if msg == nil {
				quit = true
			} else {
				msgArr = append(msgArr, msg)
			}
			// collect all remaining messages if there are any
		collectRest:
			for {
				select {
				case msg = <-lw.msg:
					if msg == nil {
						quit = true
					} else {
						msgArr = append(msgArr, msg)
					}
				default:
					break collectRest
				}
			}

		case <-time.After(100 * time.Millisecond):
		}

		// check date update
		date := time.Now().Format("2006-01-02")

		// check if date has been updated
		if lw.dateUpdatedCh != nil && date != lw.currDate {
			// non-blocking update with the old date
			if len(lw.dateUpdatedCh) > 0 {
				<-lw.dateUpdatedCh
			}
			lw.dateUpdatedCh <- lw.currDate
			lw.currDate = date
		}

		// write messages to file
		if len(msgArr) > 0 {
			var file *os.File
			if useStdout {
				file = os.Stdout
			} else {
				filename := fmt.Sprintf("%s-%s.log", lw.Name, date)
				if len(lw.DirPath) > 0 {
					filename = filepath.Join(lw.DirPath, filename)
					file, err = os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
					if err != nil {
						fmt.Printf("Error: AsyncLogWriter cannot open file %s (%v)\n", filename, err)
						continue
					}
				}
			}

			// O_CREATE: create file if it does not exist
			// O_APPEND: open at the end of file
			// O_SYNC: sync data right into disk at write. Don't use this flag to reduce file i/o
			//file, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE|os.O_SYNC, 0644)
			for _, m := range msgArr {
				_, err := file.Write(m)
				if err != nil {
					fmt.Printf("Error: AsyncLogWriter failed to write file %s (%v)\n", file.Name(), err)
					break
				}
			}

			// close after all writes
			if !useStdout {
				file.Sync()
				file.Close()
			}
		}

		if quit {
			return
		}
	}
}

func (lw *AsyncLogWriter) Close() {
	lw.Lock()
	defer lw.Unlock()

	if lw.msg != nil {
		lw.msg <- nil
		lw.wg.Wait()
		close(lw.msg)
		lw.msg = nil
	}
	if lw.dateUpdatedCh != nil {
		close(lw.dateUpdatedCh)
		lw.dateUpdatedCh = nil
	}

}

// A logger that implements async logging by default (logging without impact on real business logic).
// It also provides customizable logging functions.
// At default implementation, it is recommended to call Close() at program termination.
type Logger struct {
	sync.Mutex
	lw         *AsyncLogWriter
	lwEvaluate *AsyncLogWriter
	lwAudit    *AsyncLogWriter
	lgWrn      *log.Logger
	lgErr      *log.Logger
	lgCrt      *log.Logger
	lgEva      *log.Logger
	lgInf      *log.Logger
	lgSts      *log.Logger
	lgAdt      *log.Logger
	lgTrx      *log.Logger
	lgDbg      *log.Logger
	lgTrc      *log.Logger
	lgVbs      *log.Logger

	Warning     func(format string, args ...any)
	Error       func(format string, args ...any)
	Critical    func(format string, args ...any)
	Evaluate    func(format string, args ...any)
	Info        func(format string, args ...any)
	Stats       func(format string, args ...any)
	Audit       func(format string, args ...any)
	Transaction func(format string, args ...any)
	Debug       func(format string, args ...any)
	Trace       func(format string, args ...any)
	Verbose     func(format string, args ...any)

	logLevel    int
	callDepth   int // call depth to be adjusted by how it is called
	isSubLogger bool
	isRunning   bool

	subLoggers []*Logger
}

// Function for use in Logger for discarding logged lines.
func BlackholeLogf(format string, args ...any) {}

// NewLogger constructs a Logger that logs at the specified log l.logLevel and above.
// It decorates log lines with the log l.logLevel, date, time, and prepend.
func NewLogger(prepend string, level int, dir string, filename string) *Logger {
	l := &Logger{
		logLevel:  level,
		callDepth: 2,
	}

	// start generic log writer
	l.lw = &AsyncLogWriter{
		DirPath:       dir,
		Name:          filename,
		dateUpdatedCh: make(chan string, 1),
	}
	l.lw.Start()

	// start evaluate log writer
	l.lwEvaluate = &AsyncLogWriter{
		DirPath: dir,
		Name:    filename + "-evaluate",
	}
	l.lwEvaluate.Start()

	// start audit log writer
	l.lwAudit = &AsyncLogWriter{
		DirPath: dir,
		Name:    filename + "-audit",
	}
	l.lwAudit.Start()

	l.initActions(prepend)
	return l
}

func (l *Logger) initActions(prepend string) {
	flag := log.Ldate | log.Ltime | log.Lmsgprefix
	if ShowCallerFileLine {
		flag |= log.Lshortfile
	}

	l.lgWrn = log.New(l.lw, prepend+" [Warning] ", flag)
	l.Warning = func(format string, args ...any) {
		if l.logLevel >= LogLevelError {
			l.lgWrn.Output(l.callDepth, fmt.Sprintf(format, args...))
		}
	}

	l.lgErr = log.New(l.lw, prepend+" [Error] ", flag)
	l.Error = func(format string, args ...any) {
		if l.logLevel >= LogLevelError {
			l.lgErr.Output(l.callDepth, fmt.Sprintf(format, args...))
		}
	}

	l.lgCrt = log.New(l.lw, prepend+" [Critical] ", flag)
	l.Critical = func(format string, args ...any) {
		if l.logLevel >= LogLevelError {
			l.lgCrt.Output(l.callDepth, fmt.Sprintf(format, args...))
		}
	}

	l.lgEva = log.New(l.lwEvaluate, prepend+" [Evaluate] ", flag|log.Lmicroseconds)
	l.Evaluate = func(format string, args ...any) {
		if l.logLevel >= LogLevelError {
			l.lgEva.Output(l.callDepth, fmt.Sprintf(format, args...))
		}
	}

	l.lgInf = log.New(l.lw, prepend+" [Info] ", flag)
	l.Info = func(format string, args ...any) {
		if l.logLevel >= LogLevelInfo {
			l.lgInf.Output(l.callDepth, fmt.Sprintf(format, args...))
		}
	}

	l.lgSts = log.New(l.lw, prepend+" [Stats] ", flag)
	l.Stats = func(format string, args ...any) {
		if l.logLevel >= LogLevelInfo {
			l.lgSts.Output(l.callDepth, fmt.Sprintf(format, args...))
		}
	}

	// output to audit log writer
	l.lgAdt = log.New(l.lwAudit, prepend+" [Audit] ", flag)
	l.Audit = func(format string, args ...any) {
		if l.logLevel >= LogLevelAudit {
			l.lgAdt.Output(l.callDepth, fmt.Sprintf(format, args...))
		}
	}

	l.lgTrx = log.New(l.lwAudit, prepend+" [Transaction] ", flag)
	l.Transaction = func(format string, args ...any) {
		if l.logLevel >= LogLevelAudit {
			l.lgTrx.Output(l.callDepth, fmt.Sprintf(format, args...))
		}
	}

	l.lgDbg = log.New(l.lw, prepend+" [Debug] ", flag)
	l.Debug = func(format string, args ...any) {
		if l.logLevel >= LogLevelDebug {
			l.lgDbg.Output(l.callDepth, fmt.Sprintf(format, args...))
		}
	}

	l.lgVbs = log.New(l.lw, prepend+" [Verbose] ", flag)
	l.Verbose = func(format string, args ...any) {
		if l.logLevel >= LogLevelTrace {
			l.lgVbs.Output(l.callDepth, fmt.Sprintf(format, args...))
		}
	}

	l.lgTrc = log.New(l.lw, prepend+" [Trace] ", flag)
	l.Trace = func(format string, args ...any) {
		if l.logLevel >= LogLevelTrace {
			l.lgTrc.Output(l.callDepth, fmt.Sprintf(format, args...))
		}
	}
	l.isRunning = true
}

func (l *Logger) SetLogLevel(level int) {
	l.Lock()
	l.logLevel = level
	l.Unlock()

	if l.isSubLogger {
		return
	}

	for _, subl := range l.subLoggers {
		subl.SetLogLevel(level)
	}
}

func (l *Logger) Close() {
	if !l.isRunning {
		return
	}
	l.isRunning = false
	// stop pushing further messages by SetOutput() because it is thread-safe
	if l.lgWrn != nil {
		l.lgWrn.SetOutput(io.Discard)
	}
	if l.lgErr != nil {
		l.lgErr.SetOutput(io.Discard)
	}
	if l.lgCrt != nil {
		l.lgCrt.SetOutput(io.Discard)
	}
	if l.lgEva != nil {
		l.lgEva.SetOutput(io.Discard)
	}
	if l.lgInf != nil {
		l.lgInf.SetOutput(io.Discard)
	}
	if l.lgAdt != nil {
		l.lgAdt.SetOutput(io.Discard)
	}
	if l.lgSts != nil {
		l.lgSts.SetOutput(io.Discard)
	}
	if l.lgTrx != nil {
		l.lgTrx.SetOutput(io.Discard)
	}
	if l.lgDbg != nil {
		l.lgDbg.SetOutput(io.Discard)
	}
	if l.lgTrc != nil {
		l.lgTrc.SetOutput(io.Discard)
	}
	if l.lgVbs != nil {
		l.lgVbs.SetOutput(io.Discard)
	}
	if l.isSubLogger {
		// sublogger reuses writer so must not close the writer routine
		return
	}

	for _, subl := range l.subLoggers {
		subl.Close()
	}
	if l.lw != nil {
		l.lw.Close()
	}
	if l.lwEvaluate != nil {
		l.lwEvaluate.Close()
	}
	if l.lwAudit != nil {
		l.lwAudit.Close()
	}
}

func (l *Logger) Writer() io.Writer {
	return l.lw
}

func (l *Logger) NewSubLogger(prepend string, level int) *Logger {
	newl := &Logger{
		logLevel:    level,
		callDepth:   2,
		isSubLogger: true,
		lw:          l.lw,
		lwEvaluate:  l.lwEvaluate,
		lwAudit:     l.lwAudit,
	}

	// reuse parent's log writer
	newl.initActions(prepend)
	l.Lock()
	l.subLoggers = append(l.subLoggers, newl)
	l.Unlock()

	return newl
}

func (l *Logger) DateUpdateChan() chan string {
	return l.lw.dateUpdatedCh
}
