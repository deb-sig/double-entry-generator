//go:build cgo

// Command libdeg builds the importer as a C library for apps that embed
// DEG (for example via Rust FFI):
//
//	go build -buildmode=c-archive -o libdeg.a ./cmd/libdeg   # static, plus libdeg.h
//	go build -buildmode=c-shared  -o libdeg.so ./cmd/libdeg  # shared
//
// The API is three functions. Every call is independent and safe from any
// thread; nothing is written to disk or the network.
//
//	char *deg_import(const char *template_yaml, const char *rules_yaml,
//	                 const char *bill_name, const uint8_t *bill, size_t bill_len);
//	void  deg_free(char *p);
//	int   deg_contract_version(void);
//
// deg_import returns a NUL-terminated UTF-8 JSON document (see
// pkg/embed.Result) that the caller releases with deg_free. Errors are
// reported inside the JSON ("ok": false, "error": ...), never as NULL.
package main

/*
#include <stdint.h>
#include <stdlib.h>
*/
import "C"

import (
	"io"
	"log"
	"unsafe"

	"github.com/deb-sig/double-entry-generator/v2/pkg/embed"
)

func init() {
	// The compiler logs progress lines; a host app has its own logging.
	log.SetOutput(io.Discard)
}

//export deg_import
func deg_import(templateYAML, rulesYAML, billName *C.char, bill *C.uint8_t, billLen C.size_t) *C.char {
	var data []byte
	if bill != nil && billLen > 0 {
		data = C.GoBytes(unsafe.Pointer(bill), C.int(billLen))
	}
	out := embed.JSON(goString(templateYAML), goString(rulesYAML), goString(billName), data)
	return C.CString(string(out))
}

//export deg_free
func deg_free(p *C.char) {
	C.free(unsafe.Pointer(p))
}

//export deg_contract_version
func deg_contract_version() C.int {
	return C.int(embed.ContractVersion)
}

func goString(p *C.char) string {
	if p == nil {
		return ""
	}
	return C.GoString(p)
}

func main() {}
