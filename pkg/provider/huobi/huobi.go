package huobi

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/gaocegege/double-entry-generator/pkg/ir"
)

type Huobi struct {
	Statistics Statistics `json:"statistics,omitempty"`
	LineNum    int        `json:"line_num,omitempty`
	Orders     []Order    `json:"orders,omitempty`
}

func New() *Huobi {
	return &Huobi{
		Statistics: Statistics{},
		LineNum:    0,
		Orders:     make([]Order, 0),
	}
}

func (h *Huobi) Translate(filename string) (*ir.IR, error) {
	log.SetPrefix("[Provider-Huobi] ")

	csvFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	reader := csv.NewReader(bufio.NewReader(csvFile))
	reader.LazyQuotes = true
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

		h.LineNum++
		if h.LineNum <= 1 {
			// The first line is csv file header.
			continue
		}

		err = h.translateToOrders(line)
		if err != nil {
			return nil, fmt.Errorf("Failed to translate bill: line %d: %v", h.LineNum, err)
		}
	}
	log.Printf("Finished to parse the file %s", filename)
	return h.convertToIR(), nil
}
