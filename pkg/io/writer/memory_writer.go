package writer

import (
	"bytes"
	"strings"
	"sync"
)

// MemoryPrefix marks an output name that writes to memory instead of a
// file or page element. Each name gets its own buffer, so concurrent
// imports (an app calling the engine from several threads) don't share one.
const MemoryPrefix = "memory:"

// MemoryWriter collects compiler output in memory.
type MemoryWriter struct {
	buf bytes.Buffer
}

func (m *MemoryWriter) Write(p []byte) (n int, err error) {
	return m.buf.Write(p)
}

func (m *MemoryWriter) Close() error {
	return nil
}

func (m *MemoryWriter) String() string {
	return m.buf.String()
}

var (
	memoryMu      sync.Mutex
	memoryWriters = map[string]*MemoryWriter{}
	lastMemory    *MemoryWriter
)

func memoryWriter(name string) (*MemoryWriter, bool) {
	if !strings.HasPrefix(name, MemoryPrefix) || len(name) == len(MemoryPrefix) {
		return nil, false
	}
	memoryMu.Lock()
	defer memoryMu.Unlock()
	w := &MemoryWriter{}
	memoryWriters[name] = w
	lastMemory = w
	return w, true
}

// TakeMemory returns what was written to the named memory output and
// forgets it.
func TakeMemory(name string) (string, bool) {
	memoryMu.Lock()
	defer memoryMu.Unlock()
	w, ok := memoryWriters[name]
	delete(memoryWriters, name)
	if !ok {
		return "", false
	}
	return w.String(), true
}

// GetLastMemoryWriter returns the most recently created memory writer.
// Kept for the single-threaded browser build; prefer TakeMemory.
func GetLastMemoryWriter() *MemoryWriter {
	memoryMu.Lock()
	defer memoryMu.Unlock()
	return lastMemory
}
