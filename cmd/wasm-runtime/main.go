//go:build js && wasm

// Command wasm-runtime is the browser build of the template importer. It
// exposes one function, degRuntimeImport, that takes a template, a rules
// file and a bill and returns Beancount text. Nothing leaves the page: the
// template hub uses it so people can try their own statement locally.
package main

import (
	"fmt"
	"io"
	"log"
	"syscall/js"
	_ "time/tzdata" // the browser has no zoneinfo directory

	"github.com/deb-sig/double-entry-generator/v2/pkg/embed"
)

func main() {
	// The compiler logs progress; a blocking write from inside a JS callback
	// would deadlock the scheduler, and the page has its own status line.
	log.SetOutput(io.Discard)
	js.Global().Set("degRuntimeImport", js.FuncOf(runtimeImport))
	js.Global().Set("degRuntimeVersion", js.ValueOf("v3"))
	select {}
}

// runtimeImport(templateYAML string, rulesYAML string, billName string, bill Uint8Array)
//
//	-> Promise<{ ok, beancount, error, transactions, fixme, warnings[], report }>
//
// The work runs on its own goroutine: anything that reaches syscall/js fs
// (time.LoadLocation probing for zoneinfo, for one) waits on a JS callback,
// which can never fire while this function still holds the event loop.
func runtimeImport(_ js.Value, args []js.Value) any {
	args = append([]js.Value(nil), args...)
	var executor js.Func
	executor = js.FuncOf(func(_ js.Value, p []js.Value) any {
		resolve := p[0]
		go func() {
			defer executor.Release()
			resolve.Invoke(runImport(args))
		}()
		return nil
	})
	return js.Global().Get("Promise").New(executor)
}

func runImport(args []js.Value) (out js.Value) {
	defer func() {
		if r := recover(); r != nil {
			out = js.ValueOf(map[string]any{"ok": false, "error": fmt.Sprint(r)})
		}
	}()
	if len(args) < 4 {
		return js.ValueOf(map[string]any{"ok": false, "error": "degRuntimeImport(templateYAML, rulesYAML, billName, billBytes)"})
	}
	bill := make([]byte, args[3].Get("length").Int())
	js.CopyBytesToGo(bill, args[3])
	res := embed.Run(args[0].String(), args[1].String(), args[2].String(), bill)
	if !res.OK {
		return js.ValueOf(map[string]any{"ok": false, "error": res.Error})
	}
	warnings := make([]any, len(res.Warnings))
	for i, w := range res.Warnings {
		warnings[i] = w
	}
	return js.ValueOf(map[string]any{
		"ok":           true,
		"beancount":    res.Beancount,
		"transactions": len(res.Transactions),
		"fixme":        res.NeedsAccount,
		"warnings":     warnings,
		"report":       res.Report,
	})
}
