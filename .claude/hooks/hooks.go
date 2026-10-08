// Command hooks holds forgego's Claude Code PostToolUse hooks. It runs as a single file,
// `go run .claude/hooks/hooks.go <hook>`, so the repository needs nothing beyond Go.
//
//	lint   (Edit|Write): golangci-lint fmt, then run --fix on the edited file's package,
//	       exiting 2 with what it can't fix, so the findings reach Claude.
//	reads  (Read|Grep|Glob): past a threshold, nudges Claude toward an Explore agent.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	warningThreshold = 10
	warningInterval  = 8
)

type input struct {
	SessionID string `json:"session_id"`
	ToolInput struct {
		FilePath string `json:"file_path"`
	} `json:"tool_input"`
}

func main() {
	var in input
	if err := json.NewDecoder(os.Stdin).Decode(&in); err != nil || len(os.Args) < 2 {
		return
	}
	switch os.Args[1] {
	case "lint":
		lint(in.ToolInput.FilePath)
	case "reads":
		reads(in.SessionID)
	}
}

func lint(file string) {
	if !strings.HasSuffix(file, ".go") {
		return
	}
	golangci := []string{"tool", "-modfile=.forgego/golangci-lint.mod", "golangci-lint"}
	_ = exec.Command("go", append(golangci, "fmt", file)...).Run()
	pkg := "./" + filepath.ToSlash(filepath.Dir(relative(file)))
	out, err := exec.Command("go", append(golangci, "run", "--fix", pkg)...).CombinedOutput()
	if err != nil {
		os.Stderr.Write(out)
		os.Exit(2)
	}
}

func relative(file string) string {
	root := os.Getenv("CLAUDE_PROJECT_DIR")
	if rel, err := filepath.Rel(root, file); err == nil && root != "" {
		return rel
	}
	return file
}

func reads(session string) {
	if session == "" {
		session = "default"
	}
	path := filepath.Join(os.TempDir(), "claude-forgego-readcount-"+session)
	count := 0
	if b, err := os.ReadFile(path); err == nil {
		count, _ = strconv.Atoi(strings.TrimSpace(string(b)))
	}
	count++
	_ = os.WriteFile(path, []byte(strconv.Itoa(count)), 0o600)

	if count != warningThreshold && (count < warningThreshold || (count-warningThreshold)%warningInterval != 0) {
		return
	}
	message := fmt.Sprintf("Read/Grep/Glob call #%d this session. If you are still searching rather than "+
		"working on files you will edit, delegate the rest of the search to an Explore agent.", count)
	_ = json.NewEncoder(os.Stdout).Encode(map[string]any{
		"hookSpecificOutput": map[string]string{"hookEventName": "PostToolUse", "additionalContext": message},
	})
}
