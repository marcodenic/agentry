package debug

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	// DebugEnabled controls whether debug output is enabled
	DebugEnabled bool
	// DebugLogger is the logger used for debug output
	DebugLogger *log.Logger
	// debugLevel stores the current debug level
	debugLevel string
	// fileLogger handles rolling log files
	fileLogger *RollingLogger
	// logMutex protects concurrent access to file logger
	logMutex sync.Mutex
)

// RollingLogger implements a rotating log file system
type RollingLogger struct {
	dir       string
	basename  string
	maxSize   int64
	file      *os.File
	logger    *log.Logger
	size      int64
	mutex     sync.Mutex
}

// NewRollingLogger creates a new rolling logger with specified max size
func NewRollingLogger(dir, basename string, maxSize int64) (*RollingLogger, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	
	rl := &RollingLogger{
		dir:      dir,
		basename: basename,
		maxSize:  maxSize,
	}
	
	if err := rl.openLogFile(); err != nil {
		return nil, err
	}
	
	return rl, nil
}

func (rl *RollingLogger) openLogFile() error {
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("%s-%s.log", rl.basename, timestamp)
	path := filepath.Join(rl.dir, filename)
	
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	
	if rl.file != nil {
		rl.file.Close()
	}
	
	rl.file = file
	rl.logger = log.New(file, "", log.LstdFlags|log.Lmicroseconds)
	rl.size = 0
	
	if info, err := file.Stat(); err == nil {
		rl.size = info.Size()
	}
	
	return nil
}

func (rl *RollingLogger) Write(data []byte) (n int, err error) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	
	if rl.size+int64(len(data)) > rl.maxSize {
		if err := rl.openLogFile(); err != nil {
			return 0, err
		}
	}
	
	n, err = rl.file.Write(data)
	if err == nil {
		rl.size += int64(n)
	}
	return n, err
}

func (rl *RollingLogger) Printf(format string, v ...interface{}) {
	if rl.logger != nil {
		rl.logger.Printf(format, v...)
	}
}

func (rl *RollingLogger) Close() error {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	if rl.file != nil {
		return rl.file.Close()
	}
	return nil
}

func init() {
	// Check for consolidated debug level first
	debugLevel = os.Getenv("AGENTRY_DEBUG_LEVEL")

	// Backward compatibility with old flags
	if debugLevel == "" {
		if os.Getenv("AGENTRY_DEBUG") == "1" || os.Getenv("AGENTRY_DEBUG") == "true" {
			debugLevel = "debug"
		} else if os.Getenv("AGENTRY_COMM_LOG") == "1" {
			debugLevel = "trace" // communication logging is trace level
		} else if os.Getenv("AGENTRY_DEBUG_CONTEXT") == "1" {
			debugLevel = "debug"
		}
	}

	// Set debug enabled based on level
	DebugEnabled = debugLevel == "debug" || debugLevel == "trace"

	// Initialize rolling file logger (always, regardless of debug level)
	initFileLogger()

	if DebugEnabled {
		// In debug mode, output to stderr AND file
		DebugLogger = log.New(os.Stderr, "[DEBUG] ", log.LstdFlags)
	} else {
		// In non-debug mode, discard stderr output but still log to file
		DebugLogger = log.New(io.Discard, "", 0)
	}
}

func initFileLogger() {
	logMutex.Lock()
	defer logMutex.Unlock()
	
	// Always initialize file logger for debug capture
	var err error
	fileLogger, err = NewRollingLogger("debug", "agentry-debug", 1024*1024) // 1MB per file
	if err != nil {
		fmt.Fprintf(os.Stderr, "[DEBUG] Failed to initialize file logger: %v\n", err)
		return
	}
	
	// Log startup
	fileLogger.Printf("=== AGENTRY DEBUG SESSION STARTED ===")
	fileLogger.Printf("Debug Level: %s", debugLevel)
	fileLogger.Printf("Process PID: %d", os.Getpid())
	if wd, err := os.Getwd(); err == nil {
		fileLogger.Printf("Working Directory: %s", wd)
	}
	fileLogger.Printf("=====================================")
}

// Printf writes debug output if debug mode is enabled
func Printf(format string, v ...interface{}) {
	// Always log to file
	LogToFile("DEBUG", format, v...)
	
	// Also log to stderr if debug is enabled
	if DebugEnabled {
		DebugLogger.Printf(format, v...)
	}
}

// LogToFile always writes to the rolling log file, regardless of debug level
func LogToFile(level, format string, v ...interface{}) {
	logMutex.Lock()
	defer logMutex.Unlock()
	
	if fileLogger != nil {
		message := fmt.Sprintf(format, v...)
		fileLogger.Printf("[%s] %s", level, message)
	}
}

// LogEvent logs structured events to the debug file
func LogEvent(category, event string, data map[string]interface{}) {
	logMutex.Lock()
	defer logMutex.Unlock()
	
	if fileLogger != nil {
		fileLogger.Printf("[EVENT] %s.%s: %+v", category, event, data)
	}
}

// LogToolCall logs tool execution details
func LogToolCall(toolName string, args map[string]interface{}, result string, err error) {
	data := map[string]interface{}{
		"tool":   toolName,
		"args":   args,
		"result": truncateString(result, 500),
	}
	if err != nil {
		data["error"] = err.Error()
	}
	LogEvent("TOOL", "call", data)
}

// LogAgentAction logs agent-specific actions
func LogAgentAction(agentID, action string, details map[string]interface{}) {
	data := map[string]interface{}{
		"agent_id": agentID,
		"action":   action,
	}
	for k, v := range details {
		data[k] = v
	}
	LogEvent("AGENT", action, data)
}

// LogModelInteraction logs model API calls and responses
func LogModelInteraction(provider, model string, msgs int, tokens map[string]int, duration time.Duration) {
	data := map[string]interface{}{
		"provider": provider,
		"model":    model,
		"messages": msgs,
		"tokens":   tokens,
		"duration": duration.String(),
	}
	LogEvent("MODEL", "interaction", data)
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// IsTraceEnabled returns true if trace-level debugging is enabled
func IsTraceEnabled() bool {
	return debugLevel == "trace"
}

// IsContextDebugEnabled returns true if context debugging is enabled
func IsContextDebugEnabled() bool {
	return debugLevel == "debug" || debugLevel == "trace"
}

// IsCommLogEnabled returns true if communication logging is enabled
func IsCommLogEnabled() bool {
	return debugLevel == "trace" || os.Getenv("AGENTRY_COMM_LOG") == "1"
}

// EnableDebug dynamically enables debug output at runtime
func EnableDebug() {
	DebugEnabled = true
	debugLevel = "debug"
	if DebugLogger == nil || DebugLogger.Writer() == io.Discard {
		DebugLogger = log.New(os.Stderr, "[DEBUG] ", log.LstdFlags)
	}
	LogToFile("INFO", "Debug mode enabled at runtime")
}

// SetTUIMode adjusts logging behavior for TUI mode
func SetTUIMode(enabled bool) {
	if enabled && DebugEnabled {
		// In TUI mode, keep file logging but disable stderr to avoid interference
		DebugLogger = log.New(io.Discard, "", 0)
		LogToFile("INFO", "TUI mode enabled - stderr output disabled")
	} else if DebugEnabled {
		// Restore normal debug output to stderr
		DebugLogger = log.New(os.Stderr, "[DEBUG] ", log.LstdFlags)
		LogToFile("INFO", "TUI mode disabled - stderr output restored")
	}
}

// GetDebugLogPath returns the current debug log directory
func GetDebugLogPath() string {
	if fileLogger != nil {
		return fileLogger.dir
	}
	return "debug"
}

// CloseDebugLogger cleanly closes the debug log file
func CloseDebugLogger() {
	logMutex.Lock()
	defer logMutex.Unlock()
	
	if fileLogger != nil {
		fileLogger.Printf("=== AGENTRY DEBUG SESSION ENDED ===")
		fileLogger.Close()
	}
}
