package main

import (
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type goTestEvent struct {
	TimeRaw string  `json:"Time"`
	Action  string  `json:"Action"`
	Package string  `json:"Package"`
	Test    string  `json:"Test"`
	Elapsed float64 `json:"Elapsed"`
	Output  string  `json:"Output"`
}

type testRecord struct {
	Name      string
	Package   string
	Status    string
	Output    []string
	StartTime time.Time
	StopTime  time.Time
}

type allureStatusDetails struct {
	Message string `json:"message"`
	Trace   string `json:"trace"`
}

type allureLabel struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type allureResult struct {
	UUID          string              `json:"uuid"`
	HistoryID     string              `json:"historyId"`
	Name          string              `json:"name"`
	FullName      string              `json:"fullName"`
	Status        string              `json:"status"`
	StatusDetails allureStatusDetails `json:"statusDetails"`
	Start         int64               `json:"start"`
	Stop          int64               `json:"stop"`
	Labels        []allureLabel       `json:"labels"`
}

func main() {
	suite := flag.String("suite", "integration", "Allure suite label for generated tests")
	output := flag.String("output", "allure-results", "Directory to store Allure result files")
	flag.Parse()

	if err := os.MkdirAll(*output, 0o755); err != nil {
		exitErr(fmt.Errorf("create allure output directory: %w", err))
	}

	records, err := parseEvents(os.Stdin)
	if err != nil {
		exitErr(err)
	}

	if err := writeResults(records, *suite, *output); err != nil {
		exitErr(err)
	}
}

func parseEvents(r io.Reader) (map[string]*testRecord, error) {
	scanner := bufio.NewScanner(r)
	records := make(map[string]*testRecord)

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var ev goTestEvent
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			// ignore parse errors – they are usually unrelated lines
			continue
		}

		if ev.Test == "" {
			continue
		}

		key := ev.Package + "::" + ev.Test
		rec, ok := records[key]
		if !ok {
			rec = &testRecord{
				Name:    ev.Test,
				Package: ev.Package,
			}
			records[key] = rec
		}

		eventTime := parseTime(ev.TimeRaw)

		switch ev.Action {
		case "run":
			rec.StartTime = eventTime
		case "output":
			rec.Output = append(rec.Output, ev.Output)
		case "pass", "fail", "skip":
			rec.Status = mapStatus(ev.Action)
			if rec.StartTime.IsZero() {
				rec.StartTime = eventTime
			}
			rec.StopTime = eventTime
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read go test output: %w", err)
	}

	return records, nil
}

func writeResults(records map[string]*testRecord, suite, outputDir string) error {
	for _, rec := range records {
		if rec.Status == "" {
			continue
		}

		result := allureResult{
			UUID:     uuid.NewString(),
			Name:     rec.Name,
			FullName: rec.Package + "." + rec.Name,
			Status:   rec.Status,
			Start:    rec.StartTime.UnixMilli(),
			Stop:     rec.StopTime.UnixMilli(),
			Labels: []allureLabel{
				{Name: "language", Value: "go"},
				{Name: "framework", Value: "go test"},
				{Name: "suite", Value: suite},
				{Name: "package", Value: rec.Package},
			},
		}

		if result.Start == 0 {
			result.Start = time.Now().UnixMilli()
		}
		if result.Stop == 0 {
			result.Stop = result.Start
		}

		result.HistoryID = hashString(result.FullName)
		if msg := strings.TrimSpace(strings.Join(rec.Output, "")); msg != "" {
			result.StatusDetails = allureStatusDetails{
				Message: msg,
				Trace:   msg,
			}
		}
		if rec.Status == "passed" {
			result.StatusDetails = allureStatusDetails{}
		}

		data, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("marshal allure result for %s: %w", rec.Name, err)
		}

		file := filepath.Join(outputDir, fmt.Sprintf("%s-result.json", result.UUID))
		if err := os.WriteFile(file, data, 0o644); err != nil {
			return fmt.Errorf("write allure result %s: %w", file, err)
		}
	}
	return nil
}

func mapStatus(action string) string {
	switch action {
	case "pass":
		return "passed"
	case "fail":
		return "failed"
	case "skip":
		return "skipped"
	default:
		return "broken"
	}
}

func parseTime(raw string) time.Time {
	if raw == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil {
		return time.Time{}
	}
	return t
}

func hashString(value string) string {
	h := sha1.Sum([]byte(value))
	return hex.EncodeToString(h[:])
}

func exitErr(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
