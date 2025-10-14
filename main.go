package main

import (
    "bufio"
    "errors"
    "flag"
    "fmt"
    "math/rand"
    "os"
    "path/filepath"
    "strings"
    "time"
    "runtime/pprof"
)

func main() {
    rand.Seed(time.Now().UnixNano())

    // Flags for generation
    n := flag.Int("n", 10, "生成题目数量（默认10）")
    r := flag.Int("r", -1, "数值范围（必填，生成模式）。数值均小于该范围")

    // Flags for grading
    efile := flag.String("e", "", "题目文件路径（判分模式）")
    afile := flag.String("a", "", "答案文件路径（判分模式）")

    // Performance flags
    cpuprofile := flag.String("cpuprofile", "", "CPU性能分析输出文件（例如 cpu.prof）")
    memprofile := flag.String("memprofile", "", "内存性能分析输出文件（例如 mem.prof）")

    flag.Usage = func() {
        exe := filepath.Base(os.Args[0])
        fmt.Fprintf(os.Stderr, "用法:\n")
        fmt.Fprintf(os.Stderr, "  生成题目: %s -r <范围> [-n <数量>]\n", exe)
        fmt.Fprintf(os.Stderr, "  判定对错: %s -e <exercises>.txt -a <answers>.txt\n", exe)
        fmt.Fprintf(os.Stderr, "说明:\n")
        fmt.Fprintf(os.Stderr, "  -r 必须在生成模式下提供，表示自然数、真分数及分母的取值均在 [0, r) / [1, r) 内。\n")
        fmt.Fprintf(os.Stderr, "  生成的表达式满足：不产生负数；除法子表达式结果为真分数；运算符个数≤3；去重考虑 + 与 × 的交换等价。\n")
    }

    flag.Parse()

    // Start CPU profiling if requested
    var stopCPU func()
    if *cpuprofile != "" {
        f, err := os.Create(*cpuprofile)
        if err != nil {
            fmt.Fprintln(os.Stderr, "无法创建CPU分析文件：", err)
        } else {
            if err := pprof.StartCPUProfile(f); err != nil {
                fmt.Fprintln(os.Stderr, "启动CPU分析失败：", err)
                f.Close()
            } else {
                stopCPU = func() {
                    pprof.StopCPUProfile()
                    f.Close()
                }
            }
        }
    }
    // Write heap profile on exit if requested
    if *memprofile != "" {
        defer func() {
            f, err := os.Create(*memprofile)
            if err != nil {
                fmt.Fprintln(os.Stderr, "无法创建内存分析文件：", err)
                return
            }
            pprof.WriteHeapProfile(f)
            f.Close()
        }()
    }

    // Decide mode
    if *efile != "" || *afile != "" {
        // Grading mode
        if *efile == "" || *afile == "" {
            fmt.Fprintln(os.Stderr, "错误：判分模式需同时提供 -e 与 -a。")
            flag.Usage()
            os.Exit(1)
        }
        if err := grade(*efile, *afile); err != nil {
            fmt.Fprintln(os.Stderr, "判分失败：", err)
            os.Exit(1)
        }
        if stopCPU != nil { stopCPU() }
        return
    }

    // Generation mode requires -r
    if *r <= 0 {
        fmt.Fprintln(os.Stderr, "错误：生成模式必须提供 -r 参数且为正整数。")
        flag.Usage()
        os.Exit(1)
    }
    if *n <= 0 {
        fmt.Fprintln(os.Stderr, "错误：生成数量 -n 必须为正整数。")
        os.Exit(1)
    }

    exercises, answers, err := GenerateProblems(*n, *r)
    if err != nil {
        fmt.Fprintln(os.Stderr, "生成题目失败：", err)
        os.Exit(1)
    }

    // Write files in current directory
    if err := writeLines("Exercises.txt", exercises); err != nil {
        fmt.Fprintln(os.Stderr, "写Exercises.txt失败：", err)
        os.Exit(1)
    }
    if err := writeLines("Answers.txt", answers); err != nil {
        fmt.Fprintln(os.Stderr, "写Answers.txt失败：", err)
        os.Exit(1)
    }

    fmt.Printf("已生成 %d 道题目到 Exercises.txt，并写入答案到 Answers.txt\n", len(exercises))
    if stopCPU != nil { stopCPU() }
}

func writeLines(path string, lines []string) error {
    f, err := os.Create(path)
    if err != nil {
        return err
    }
    defer f.Close()
    w := bufio.NewWriter(f)
    for _, line := range lines {
        if _, err := w.WriteString(line + "\n"); err != nil {
            return err
        }
    }
    return w.Flush()
}

// grade reads exercises and answers, outputs Grade.txt
func grade(exPath, ansPath string) error {
    exercises, err := readLines(exPath)
    if err != nil {
        return err
    }
    answers, err := readLines(ansPath)
    if err != nil {
        return err
    }
    if len(exercises) != len(answers) {
        return errors.New("题目与答案行数不一致")
    }

    var correctIdx []int
    var wrongIdx []int

    for i := 0; i < len(exercises); i++ {
        ex := strings.TrimSpace(exercises[i])
        // Expect format: e = [maybe empty or provided]; we only parse left of '='
        parts := strings.Split(ex, "=")
        exprStr := strings.TrimSpace(parts[0])

        rpn, err := ParseExpression(exprStr)
        if err != nil {
            // If parsing fails, consider wrong
            wrongIdx = append(wrongIdx, i+1)
            continue
        }
        got := rpn.Eval()

        ansStr := strings.TrimSpace(answers[i])
        ansVal, err := ParseNumber(ansStr)
        if err != nil {
            wrongIdx = append(wrongIdx, i+1)
            continue
        }

        if got.Equals(ansVal) {
            correctIdx = append(correctIdx, i+1)
        } else {
            wrongIdx = append(wrongIdx, i+1)
        }
    }

    // Build Grade.txt content
    correct := fmt.Sprintf("Correct: %d (%s)", len(correctIdx), joinInts(correctIdx))
    wrong := fmt.Sprintf("Wrong: %d (%s)", len(wrongIdx), joinInts(wrongIdx))
    return writeLines("Grade.txt", []string{correct, wrong})
}

func readLines(path string) ([]string, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer f.Close()
    var lines []string
    s := bufio.NewScanner(f)
    for s.Scan() {
        lines = append(lines, s.Text())
    }
    if err := s.Err(); err != nil {
        return nil, err
    }
    return lines, nil
}

func joinInts(ints []int) string {
    if len(ints) == 0 {
        return ""
    }
    var b strings.Builder
    for i, v := range ints {
        if i > 0 {
            b.WriteString(", ")
        }
        b.WriteString(fmt.Sprintf("%d", v))
    }
    return b.String()
}