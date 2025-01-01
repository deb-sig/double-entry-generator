/*
Copyright © 2025 None

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

package mt

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"

	"github.com/deb-sig/double-entry-generator/pkg/io/reader"
	"github.com/deb-sig/double-entry-generator/pkg/ir"
)

// MT is the provider for Meituan
type MT struct {
	Statistics Statistics `json:"statistics,omitempty"`
	LineNum    int        `json:"line_num,omitempty"`
	Orders     []Order    `json:"orders,omitempty"`

	// TitleParsed is a workaround to ignore the title row.
	TitleParsed bool `json:"title_parsed,omitempty"`
}

func New() *MT {
	return &MT{
		Statistics:  Statistics{},
		LineNum:     0,
		Orders:      make([]Order, 0),
		TitleParsed: false,
	}
}

func (mt *MT) Translate(filename string) (*ir.IR, error) {
	log.SetPrefix("[Provider-MT] ")
	billReader, err := reader.GetReader(filename)
	if err != nil {
		return nil, fmt.Errorf("can't get bill reader, err: %v", err)
	}

	csvReader := csv.NewReader(billReader)
	csvReader.LazyQuotes = true // 可以处理不规范的引号
	// If FieldsPerRecord is negative, no check is made and records
	// may have a variable number of fields.
	csvReader.FieldsPerRecord = -1

	for {
		line, err := csvReader.Read()
		if err == io.EOF { // 文件读取结束
			break
		} else if err != nil {
			return nil, err
		}
		mt.LineNum++
		if mt.LineNum <= 20 {
			// bypass the useless content
			continue
		}
		err = mt.translateToOrders(line)
		if err != nil {
			return nil, fmt.Errorf("Failed to translate bill: line %d: %v", mt.LineNum, err)
		}
	}
	log.Printf("Finished to parse the file %s", filename)

	ir := mt.convertToIR()
	// return mt.postProcess(ir), nil
	return ir, nil
}

// func (mt *MT) postProcess(ir_ *ir.IR) *ir.IR {

// }
