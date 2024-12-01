package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// PinoLog represents the structure of a Pino log entry
type PinoLog struct {
	Level    int            `json:"level"`
	Time     int64          `json:"time"`
	Pid      int            `json:"pid"`
	Hostname string         `json:"hostname"`
	Msg      string         `json:"msg"`
	Data     map[string]any `json:"-"` // Will catch any additional fields
}

// ANSI color codes
const (
	reset   = "\033[0m"
	red     = "\033[31m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	blue    = "\033[34m"
	magenta = "\033[35m"
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

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "-v" || os.Args[1] == "--version") {
		fmt.Printf("%spino-print%s %sv0.1.0%s\n", green, reset, blue, reset)
		os.Exit(0)
	}

	if len(os.Args) > 1 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		fmt.Printf("A pretty printer for %sPino%s logs.\n", blue, reset)
		fmt.Printf("  Updates, issues: %shttps://github.com/vsly-ru/pino-print%s\n", blue, reset)
		fmt.Printf("\n%sUsage:%s\n", blue, reset)
		fmt.Printf("  pipe your %sPino%s JSON logs to %spino-print%s\n", blue, reset, green, reset)
		fmt.Printf("\n%sExample:%s\n", blue, reset)
		fmt.Printf("  %snode app.js%s | %spino-print%s\n", grey, reset, green, reset)
		fmt.Printf("\n%sOptions:%s\n", blue, reset)
		fmt.Printf("  %s-h, --help%s     Show this help message\n", yellow, reset)
		fmt.Printf("  %s-v, --version%s  Show version information\n", yellow, reset)
		os.Exit(0)
	}

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		line := scanner.Text()

		// Parse the entire JSON into a map first
		var rawLog map[string]any
		if err := json.Unmarshal([]byte(line), &rawLog); err != nil {
			fmt.Fprintf(os.Stdout, "%s\n", line)
			continue
		}

		// Parse the known fields into our struct
		var log PinoLog
		if err := json.Unmarshal([]byte(line), &log); err != nil {
			fmt.Fprintf(os.Stdout, "%s\n", line)
			continue
		}

		// Get extra data fields (excluding the known ones)
		dataFields := []string{}
		for k, v := range rawLog {
			switch k {
			case "level", "time", "pid", "hostname", "msg":
				continue
			default:
				dataFields = append(dataFields, fmt.Sprintf("%s%s%s:%v", yellow, k, reset, v))
			}
		}

		// Format timestamp
		timestamp := time.UnixMilli(log.Time).Format("2006-01-02 15:04:05.000")

		// Get level color and name
		levelColor := getLevelColor(log.Level)
		levelName := getLevelName(log.Level)

		// Print formatted log
		fmt.Printf("%s %s[%s]%s %s",
			timestamp,
			levelColor,
			levelName,
			reset,
			log.Msg,
		)

		// Print extra data fields if any
		if len(dataFields) > 0 {
			fmt.Printf("\n • %s", strings.Join(dataFields, "\n • "))
		}
		fmt.Println()
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}
}
