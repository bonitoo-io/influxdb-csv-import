package cmd

import (
	"encoding/csv"
	"fmt"
	"io"
)

type lineReader struct {
	// csv reading
	csv        *csv.Reader
	table      CsvTable
	lineNumber int

	// reader results
	buffer   []byte
	index    int
	finished error
}

func (state *lineReader) Read(p []byte) (n int, err error) {
	if state.finished != nil {
		return 0, state.finished
	}
	if len(state.buffer) > state.index {
		// we have remaining bytes to copy
		if len(state.buffer)-state.index > len(p) {
			// copy a part of the buffer
			copy(p, state.buffer[state.index:state.index+len(p)])
			state.index += len(p)
			return len(p), nil
		}
		// copy the entire buffer
		n = len(state.buffer) - state.index
		copy(p[:n], state.buffer[state.index:])
		state.buffer = state.buffer[:0]
		state.index = 0
		return n, nil
	}
	for {
		// Read each record from csv
		state.lineNumber++
		row, err := state.csv.Read()
		if err != nil {
			state.finished = err
			return state.Read(p)
		}
		state.csv.FieldsPerRecord = 0 // because every row can have different count of columns
		if state.table.AddRow(row) {
			buffer, err := state.table.AppendLine(state.buffer, row)
			if err != nil {
				state.finished = fmt.Errorf("Line #%d: %v", state.lineNumber, err)
				return state.Read(p)
			}
			state.buffer = append(buffer, '\n')
			break
		}
	}
	return state.Read(p)
}

// CsvToProtocolLines transforms csv data into line protocol data
func CsvToProtocolLines(reader io.Reader) io.Reader {
	return &lineReader{
		csv: csv.NewReader(reader),
	}
}
