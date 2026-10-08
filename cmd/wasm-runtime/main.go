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
	"strings"
	"syscall/js"

	"github.com/deb-sig/double-entry-generator/v2/pkg/analyser/api"
	"github.com/deb-sig/double-entry-generator/v2/pkg/compiler/beancount"
	"github.com/deb-sig/double-entry-generator/v2/pkg/config"
	"github.com/deb-sig/double-entry-generator/v2/pkg/consts"
	"github.com/deb-sig/double-entry-generator/v2/pkg/importer"
	"github.com/deb-sig/double-entry-generator/v2/pkg/io/writer"
	"github.com/deb-sig/double-entry-generator/v2/pkg/ir"
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
//	-> { ok, beancount, error, transactions, fixme, warnings[], report }
func runtimeImport(_ js.Value, args []js.Value) any {
	result := map[string]any{"ok": false}
	if len(args) < 4 {
		result["error"] = "degRuntimeImport(templateYAML, rulesYAML, billName, billBytes)"
		return js.ValueOf(result)
	}
	templateYAML := args[0].String()
	rulesYAML := args[1].String()
	billName := args[2].String()
	bill := make([]byte, args[3].Get("length").Int())
	js.CopyBytesToGo(bill, args[3])

	text, stats, err := importInMemory(templateYAML, rulesYAML, billName, bill)
	if err != nil {
		result["error"] = err.Error()
		return js.ValueOf(result)
	}
	result["ok"] = true
	result["beancount"] = text
	for k, v := range stats {
		result[k] = v
	}
	return js.ValueOf(result)
}

func importInMemory(templateYAML, rulesYAML, billName string, bill []byte) (string, map[string]any, error) {
	profile, err := importer.LoadProfileBytes([]byte(templateYAML), "template")
	if err != nil {
		return "", nil, fmt.Errorf("template: %w", err)
	}
	warnings := []any{}
	if strings.TrimSpace(rulesYAML) != "" {
		rf, err := importer.ParseRulesFile([]byte(rulesYAML))
		if err != nil {
			return "", nil, err
		}
		profile.ApplyRulesFile(rf)
		for _, w := range importer.PersonalRuleWarnings(profile, rf.AllPersonalRules(), "", "") {
			warnings = append(warnings, w)
		}
	}
	out, report, err := importer.ImportBytes(profile, billName, bill)
	if err != nil {
		return "", nil, err
	}
	cfg := &config.Config{
		Title:               firstNonEmpty(profile.Name, profile.ID, "DEG Import"),
		DefaultMinusAccount: profile.Template.DefaultMinus,
		DefaultPlusAccount:  profile.Template.DefaultPlus,
		DefaultCurrency:     firstNonEmpty(profile.Template.DefaultCurrency, "CNY"),
	}
	c, err := beancount.New(importer.DefaultProviderName, consts.CompilerBeanCount, "memory:runtime", false, cfg, out, api.Runtime{})
	if err != nil {
		return "", nil, err
	}
	if err := c.Compile(); err != nil {
		return "", nil, err
	}
	mem := writer.GetLastMemoryWriter()
	if mem == nil {
		return "", nil, fmt.Errorf("no in-memory output")
	}
	text := mem.String()
	return text, map[string]any{
		"transactions": len(out.Orders),
		"fixme":        countFIXME(out),
		"warnings":     warnings,
		"report":       report.String(),
	}, nil
}

func countFIXME(out *ir.IR) int {
	n := 0
	for _, o := range out.Orders {
		for _, src := range o.Sources {
			if src.Origin == ir.FieldOriginEngine {
				n++
				break
			}
		}
	}
	return n
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
