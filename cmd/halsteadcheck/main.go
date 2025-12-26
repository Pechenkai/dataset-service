package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"os"
	"path/filepath"
	"strings"
)

type metrics struct {
	n1 int
	n2 int
	N1 int
	N2 int
}

type counter struct {
	operators map[string]int
	operands  map[string]int
}

func newCounter() *counter {
	return &counter{
		operators: make(map[string]int),
		operands:  make(map[string]int),
	}
}

func (c *counter) addOperator(op string) {
	c.operators[op]++
}

func (c *counter) addOperand(op string) {
	c.operands[op]++
}

func (c *counter) metrics() metrics {
	m := metrics{}
	for _, v := range c.operators {
		m.N1 += v
	}
	for _, v := range c.operands {
		m.N2 += v
	}
	m.n1 = len(c.operators)
	m.n2 = len(c.operands)
	return m
}

func main() {
	maxVolume := flag.Float64("max-volume", 4000, "maximum allowed Halstead volume per function")
	flag.Parse()

	targets := flag.Args()
	if len(targets) == 0 {
		targets = []string{"./internal", "./cmd", "./main.go"}
	}

	files := collectGoFiles(targets)
	failures := analyzeFiles(files, *maxVolume)

	if len(failures) > 0 {
		fmt.Fprintf(os.Stderr, "Halstead volume exceeded %.2f in %d function(s):\n", *maxVolume, len(failures))
		for _, f := range failures {
			fmt.Fprintf(os.Stderr, "- %s: volume=%.2f length=%d vocabulary=%d\n", f.name, f.volume, f.length, f.vocabulary)
		}
		os.Exit(1)
	}
	fmt.Printf("Halstead check passed for %d file(s). Max volume %.2f\n", len(files), *maxVolume)
}

type failure struct {
	name       string
	volume     float64
	length     int
	vocabulary int
}

func collectGoFiles(paths []string) []string {
	var files []string
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			continue
		}
		if info.IsDir() {
			filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil
				}
				if info.IsDir() {
					if strings.Contains(path, string(filepath.Separator)+".git") || strings.Contains(path, string(filepath.Separator)+"vendor") {
						return filepath.SkipDir
					}
					return nil
				}
				if strings.HasSuffix(info.Name(), ".go") && !strings.HasSuffix(info.Name(), "_test.go") {
					files = append(files, path)
				}
				return nil
			})
		} else if strings.HasSuffix(p, ".go") && !strings.HasSuffix(p, "_test.go") {
			files = append(files, p)
		}
	}
	return files
}

func analyzeFiles(files []string, maxVolume float64) []failure {
	var failures []failure
	fset := token.NewFileSet()
	for _, file := range files {
		astFile, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to parse %s: %v\n", file, err)
			continue
		}
		for _, decl := range astFile.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			name := fmt.Sprintf("%s:%d", file, fset.Position(fn.Pos()).Line)
			m := computeMetrics(fn)
			volume, length, vocab := halsteadVolume(m)
			if volume > maxVolume {
				failures = append(failures, failure{
					name:       name,
					volume:     volume,
					length:     length,
					vocabulary: vocab,
				})
			}
		}
	}
	return failures
}

func computeMetrics(fn *ast.FuncDecl) metrics {
	cnt := newCounter()
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.BinaryExpr:
			cnt.addOperator(node.Op.String())
		case *ast.UnaryExpr:
			cnt.addOperator(node.Op.String())
		case *ast.AssignStmt:
			cnt.addOperator(node.Tok.String())
		case *ast.IncDecStmt:
			cnt.addOperator(node.Tok.String())
		case *ast.BranchStmt:
			cnt.addOperator(node.Tok.String())
		case *ast.IfStmt:
			cnt.addOperator("if")
		case *ast.ForStmt:
			cnt.addOperator("for")
		case *ast.RangeStmt:
			cnt.addOperator("range")
		case *ast.SwitchStmt:
			cnt.addOperator("switch")
		case *ast.TypeSwitchStmt:
			cnt.addOperator("type-switch")
		case *ast.SelectStmt:
			cnt.addOperator("select")
		case *ast.DeferStmt:
			cnt.addOperator("defer")
		case *ast.GoStmt:
			cnt.addOperator("go")
		case *ast.ReturnStmt:
			cnt.addOperator("return")
		case *ast.CallExpr:
			if ident, ok := node.Fun.(*ast.Ident); ok {
				cnt.addOperator("call:" + ident.Name)
			}
		case *ast.Ident:
			if node.Name != "_" {
				cnt.addOperand(node.Name)
			}
		case *ast.BasicLit:
			cnt.addOperand(node.Value)
		}
		return true
	})
	return cnt.metrics()
}

func halsteadVolume(m metrics) (float64, int, int) {
	vocabulary := m.n1 + m.n2
	length := m.N1 + m.N2
	if vocabulary == 0 || length == 0 {
		return 0, length, vocabulary
	}
	volume := float64(length) * math.Log2(float64(vocabulary))
	return volume, length, vocabulary
}
