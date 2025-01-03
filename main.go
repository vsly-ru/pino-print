package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	targetPrintTime = 333 * time.Millisecond // target print time for typewriter mode
	minPrintTime    = 1 * time.Millisecond   // minimum to maintain some animation
)

var (
	flagTypewriter = flag.Bool("tw", false, "Enable typewriter animation mode")
	printQueue     = make(chan string, 1024)
	printWg        sync.WaitGroup
)

func init() {
	// Start the print worker
	go printWorker()
}

func printWorker() {
	for text := range printQueue {
		// Calculate target print time with exponential reduction
		queueLen := len(printQueue)
		printTime := targetPrintTime
		if queueLen > 0 {
			// Reduce print time as queue grows
			reduction := time.Duration(queueLen) * time.Millisecond
			printTime = max(targetPrintTime-reduction, minPrintTime)
		}

		// Calculate delay between characters
		chars := len([]rune(text))
		if chars == 0 {
			chars = 1
		}
		charDelay := printTime / time.Duration(chars)

		// Print the text
		for _, char := range text {
			fmt.Print(string(char))
			time.Sleep(charDelay)
		}
		fmt.Println()

		printWg.Done()
	}
}

func print(text string) {
	if *flagTypewriter {
		printWg.Add(1) // Add a print job to wait for
		printQueue <- text
	} else {
		fmt.Println(text)
	}
}

// PinoLog represents the structure of a Pino log entry
type PinoLog struct {
	Level    int            `json:"level"`
	Time     int64          `json:"time"`
	Pid      int            `json:"pid"`
	Hostname string         `json:"hostname"`
	Msg      string         `json:"msg"`
	Module   string         `json:"module"`  // Module name
	Service  string         `json:"service"` // Service name
	Data     map[string]any `json:"-"`       // Will catch any additional fields
}

// ANSI color codes
const (
	reset   = "\033[0m"
	bold    = "\033[1m"
	red     = "\033[31m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	blue    = "\033[34m"
	magenta = "\033[95m"
	purple  = "\033[35m"
	grey    = "\033[90m"
)

func getLevelColor(level int) string {
	switch level {
	case 60: // fatal
		return red
	case 50: // error
		return red
	case 40: // warn
		return yellow
	case 30: // info
		return green
	case 20: // debug
		return blue
	case 10: // trace
		return grey
	default:
		return reset
	}
}

func getLevelName(level int) string {
	switch level {
	case 60:
		return "FATAL"
	case 50:
		return "ERROR"
	case 40:
		return "WARN"
	case 30:
		return "INF"
	case 20:
		return "DBG"
	case 10:
		return "TRC"
	default:
		return "UNKNOWN"
	}
}

// Add this new helper function
func formatDataValue(value any, indent string) string {
	switch v := value.(type) {
	case map[string]any:
		if len(v) == 0 {
			return "{}"
		}
		var fields []string
		for k, val := range v {
			fields = append(fields, fmt.Sprintf("%s%s%s%s: %s",
				indent+" • ",
				blue,
				k,
				reset,
				formatDataValue(val, indent+" • "),
			))
		}
		return "\n" + strings.Join(fields, "\n")
	case []any:
		if len(v) == 0 {
			return "[]"
		}
		var items []string
		for _, item := range v {
			items = append(items, fmt.Sprintf("%s- %s",
				indent+"    ",
				formatDataValue(item, indent+"    "),
			))
		}
		return "\n" + strings.Join(items, "\n")
	case float64:
		if v == float64(int64(v)) {
			return fmt.Sprintf("%.0f", v)
		}
		return fmt.Sprintf("%.6f", v)
	case float32:
		if float64(v) == float64(int32(v)) {
			return fmt.Sprintf("%.0f", v)
		}
		return fmt.Sprintf("%.6f", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "-v" || os.Args[1] == "--version") {
		fmt.Printf("%spino-print%s %sv0.1.5%s\n", green, reset, blue, reset)
		os.Exit(0)
	}

	if len(os.Args) > 1 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		fmt.Printf("A pretty printer for %sPino%s logs.\n", blue, reset)
		fmt.Printf("  Updates, issues: %shttps://github.com/vsly-ru/pino-print%s\n", blue, reset)
		fmt.Printf("\n%sUsage:%s\n", blue, reset)
		fmt.Printf("  Pipe your %sPino%s JSON logs to %spino-print%s\n", blue, reset, green, reset)
		fmt.Printf("\n%sExample:%s\n", blue, reset)
		fmt.Printf("  %snode app.js%s | %spino-print%s\n", grey, reset, green, reset)
		fmt.Printf("\n%sOptions:%s\n", blue, reset)
		fmt.Printf("  %s-h, --help%s     Show this help message\n", yellow, reset)
		fmt.Printf("  %s-v, --version%s  Show version information\n", yellow, reset)
		fmt.Printf("  %s-tw, --tw%s      Enable typewriter animation mode\n", yellow, reset)
		os.Exit(0)
	}

	// Move ourselves out of the foreground process group
	// This allows signals to be propagated to all processes in the pipeline
	// Fixes the issue where ^C would only kill the main process and not the pipeline
	if err := syscall.Setpgid(0, 0); err != nil {
		fmt.Fprintf(os.Stderr, "pino-print: Warning: couldn't set process group: %v\n", err)
	}

	flag.Parse()

	scanner := bufio.NewScanner(os.Stdin)
	// buffer 400 MB
	buf := make([]byte, 400*1024*1024)
	scanner.Buffer(buf, len(buf))

	for scanner.Scan() {
		line := scanner.Text()

		// Parse the entire JSON into a map first
		var rawLog map[string]any
		if err := json.Unmarshal([]byte(line), &rawLog); err != nil {
			print(line)
			continue
		}

		// Parse the known fields into our struct
		var log PinoLog
		if err := json.Unmarshal([]byte(line), &log); err != nil {
			print(line)
			continue
		}

		// Skip if no timestamp (probably not a Pino log)
		if log.Time == 0 {
			print(line)
			continue
		}

		// Get extra data fields (excluding the known ones)
		dataFields := []string{}
		for k, v := range rawLog {
			switch k {
			case "level", "time", "pid", "hostname", "msg", "module", "service":
				continue
			default:
				dataFields = append(dataFields, fmt.Sprintf("%s%s%s: %s",
					blue, k, reset,
					formatDataValue(v, ""),
				))
			}
		}

		// Format timestamp
		timestamp := time.UnixMilli(log.Time).Format("2006-01-02 15:04:05.000")

		levelColor := getLevelColor(log.Level)
		levelName := getLevelName(log.Level)

		module := log.Module
		if module == "" {
			module = log.Service
		}
		if module != "" {
			module = "|" + module
		}

		// Format
		logLine := fmt.Sprintf("%s %s[%s%s%s%s%s]%s %s",
			timestamp,
			levelColor,
			levelName,
			bold+purple,
			module,
			reset,
			levelColor,
			reset,
			log.Msg,
		)

		// Add extra data fields if any
		if len(dataFields) > 0 {
			logLine += fmt.Sprintf("\n» %s", strings.Join(dataFields, "\n» "))
		}

		print(logLine)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	if *flagTypewriter {
		close(printQueue)
		printWg.Wait()
	}
}
