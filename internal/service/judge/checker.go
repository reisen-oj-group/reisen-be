package judge

import (
    "bytes"
    "fmt"
    "os/exec"
	"reisen-be/internal/model"
    "strings"
)

type Checker interface {
    Check(output, expectedOutput string) (bool, string)
}

type StrictChecker struct{}

func (c *StrictChecker) Check(output, answer string) (bool, string) {
    row := 1
    col := 1
    for i := range(len(answer) + 1) {
        outEof := i >= len(output)
        ansEof := i >= len(answer)
        if outEof != ansEof {
            if outEof {
                return false, "Unexpected end of file."
            } else 
            if ansEof {
                return false, fmt.Sprintf("Expect end of file, read '%c'.", output[i])
            }
        } else 
        if ansEof {
            break
        } else
        if output[i] != answer[i] {
            return false, fmt.Sprintf("On line %d column %d, read '%c', expected '%c'.", row, col, output[i], answer[i])
        } else {
            if answer[i] == '\n' {
                row = row + 1
                col = 1
            } else {
                col = col + 1
            }
        }
    }
    return true, fmt.Sprintf("OK, %d character(s)", len(answer))
}

type LooseChecker struct{}

func (c *LooseChecker) Check(output, answer string) (bool, string) {
    cleanLines := func(s string) []string {
        lines := strings.Split(s, "\n")
        for i := range lines {
            lines[i] = strings.TrimRight(strings.ReplaceAll(lines[i], "\r", ""), " \t")
        }
        // 移除末尾的空行
        for len(lines) > 0 && lines[len(lines)-1] == "" {
            lines = lines[:len(lines)-1]
        }
        return lines
    }

    outLines := cleanLines(output)
    ansLines := cleanLines(answer)

    minLen := len(outLines)
    if len(ansLines) < minLen {
        minLen = len(ansLines)
    }
    for i := 0; i < minLen; i++ {
        if outLines[i] != ansLines[i] {
            return false, fmt.Sprintf("Line %d differs.\nOutput:  \"%s\"\nAnswer:  \"%s\"", i+1, outLines[i], ansLines[i])
        }
    }
    if len(outLines) != len(ansLines) {
        return false, fmt.Sprintf("Line count mismatch. Output has %d line(s), answer has %d line(s).", len(outLines), len(ansLines))
    }
    charCount := 0
    for _, line := range ansLines {
        charCount += len(line)
    }
    return true, fmt.Sprintf("OK, %d character(s)", charCount)
}

type CustomChecker struct {
    CheckerPath string
}

func (c *CustomChecker) Check(output, expectedOutput string) (bool, string) {
    cmd := exec.Command(c.CheckerPath, expectedOutput)
    cmd.Stdin = strings.NewReader(output)
    
    var out bytes.Buffer
    cmd.Stdout = &out
    cmd.Stderr = &out
    
    if err := cmd.Run(); err != nil {
        return false, out.String()
    }
    
    return true, out.String()
}

func NewChecker(config model.JudgeConfig) (Checker, error) {
    switch config.CheckerType {
    case "strict":
        return &StrictChecker{}, nil
    case "loose":
        return &LooseChecker{}, nil
    default:
        return nil, fmt.Errorf("unknown checker type: %s", config.CheckerType)
    }
}