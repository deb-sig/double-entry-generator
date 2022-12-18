/*
Copyright © 2019 Ce Gao

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package alipay

import (
	"encoding/csv"
	"fmt"
	"github.com/deb-sig/double-entry-generator/pkg/io/reader"
	"github.com/deb-sig/double-entry-generator/pkg/ir"
	"io"
	"log"
)

// Alipay is the provider for alipay.
type Alipay struct {
	Statistics Statistics `json:"statistics,omitempty"`
	LineNum    int        `json:"line_num,omitempty"`
	Orders     []Order    `json:"orders,omitempty"`

	// TitleParsed is a workaround to ignore the title row.
	TitleParsed bool `json:"title_parsed,omitempty"`
}

// New creates a new Alipay provider.
func New() *Alipay {
	return &Alipay{
		Statistics:  Statistics{},
		LineNum:     0,
		Orders:      make([]Order, 0),
		TitleParsed: false,
	}
}

// Translate translates the alipay bill records to IR.
func (a *Alipay) Translate(filename string) (*ir.IR, error) {
	log.SetPrefix("[Provider-Alipay] ")

	billReader, err := reader.GetGBKReader(filename)
	if err != nil {
		return nil, fmt.Errorf("can't get bill reader, err: %v", err)
	}

	csvReader := csv.NewReader(billReader)
	csvReader.LazyQuotes = true
	// If FieldsPerRecord is negative, no check is made and records
	// may have a variable number of fields.
	csvReader.FieldsPerRecord = -1

	for {
		line, err := csvReader.Read()

		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		if len(line) != 12 {
			// TODO(gaocegege): Support statistics.
			a.LineNum++
			continue
		}

		err = a.translateToOrders(line)
		if err != nil {
			return nil, fmt.Errorf("Failed to translate bill: line %d: %v",
				a.LineNum, err)
		}
	}
	log.Printf("Finished to parse the file %s", filename)

	return a.convertToIR(), nil
}
