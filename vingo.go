package vingo

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// AutoEscape: template motoru değişken çıktılarında varsayılan olarak HTML kaçışını (escaping) aktifleştirir.
// Bunu false yaparak global olarak otomatik kaçışı kapatabilirsiniz (önerilmez).
var AutoEscape = true

type Template struct {
	Filepath string
	Nodes    []Node
	ModTime  time.Time
}

var (
	// cache: filepath -> compiled template
	tplCache   = map[string]*Template{}
	cacheMutex sync.RWMutex
)

// Render: template dosyasını oku, compile et (gerekirse cache'den), ve işle
func Render(file string, data map[string]interface{}) (string, error) {
	abs, err := filepath.Abs(file)
	if err != nil {
		abs = file
	}

	tpl, err := getOrCompile(abs)
	if err != nil {
		return "", err
	}

	// Evaluate
	out := &strings.Builder{}
	for _, n := range tpl.Nodes {
		out.WriteString(n.Eval(data))
	}
	return out.String(), nil
}

// getOrCompile: cache kontrolü + compile
func getOrCompile(path string) (*Template, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	mod := stat.ModTime()

	cacheMutex.RLock()
	tpl, exists := tplCache[path]
	cacheMutex.RUnlock()

	if exists && tpl.ModTime.Equal(mod) {
		return tpl, nil
	}

	// compile
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	content := string(b)

	tokens := tokenize(content)
	nodes, err := compileTokens(tokens)
	if err != nil {
		return nil, err
	}

	newTpl := &Template{
		Filepath: path,
		Nodes:    nodes,
		ModTime:  mod,
	}

	cacheMutex.Lock()
	tplCache[path] = newTpl
	cacheMutex.Unlock()

	return newTpl, nil
}

func evalNodes(nodes []Node, data map[string]interface{}) string {
	out := &strings.Builder{}
	for _, n := range nodes {
		out.WriteString(n.Eval(data))
	}
	return out.String()
}

func shallowCopyMap(m map[string]interface{}) map[string]interface{} {
	n := make(map[string]interface{}, len(m)+4)
	for k, v := range m {
		n[k] = v
	}
	return n
}
