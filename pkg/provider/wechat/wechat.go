package wechat

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/gaocegege/double-entry-generator/pkg/ir"
)

// Wechat is the provider for Wechat.
type Wechat struct {
	Statistics Statistics `json:"statistics,omitempty"`
	LineNum    int        `json:"line_num,omitempty"`
	Orders     []Order    `json:"orders,omitempty"`
}

// New creates a new Alipay provider.
func New() *Wechat {
	return &Wechat{
		Statistics: Statistics{},
		LineNum:    0,
		Orders:     make([]Order, 0),
	}
}

// Translate translates the alipay bill records to IR.
func (w *Wechat) Translate(filename string) (*ir.IR, error) {
	log.SetPrefix("[Provider-Wechat] ")

	csvFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	reader := csv.NewReader(bufio.NewReader(csvFile))
	// If FieldsPerRecord is negative, no check is made and records
	// may have a variable number of fields.
	reader.FieldsPerRecord = -1

	for {
		line, err := reader.Read()

		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		w.LineNum++
		if w.LineNum <= 17 {
			// The first 17 lines are useless for us.
			continue
		}

		err = w.translateToOrders(line)
		if err != nil {
			return nil, fmt.Errorf("Failed to translate bill: line %d: %v",
				w.LineNum, err)
		}
	}
	log.Printf("Finished to parse the file %s", filename)
	return w.convertToIR(), nil
}
