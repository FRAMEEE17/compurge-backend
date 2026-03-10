package parser

import (
	"fmt"
	"io"

	"github.com/xuri/excelize/v2"
)

type XLSXParser struct{}

func (p XLSXParser) Open(reader io.Reader) ([]string, func() ([]string, error), error) {
	workbook, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, nil, fmt.Errorf("open xlsx reader: %w", err)
	}

	sheets := workbook.GetSheetList()
	if len(sheets) == 0 {
		return nil, nil, fmt.Errorf("xlsx workbook has no sheets")
	}

	rows, err := workbook.Rows(sheets[0])
	if err != nil {
		return nil, nil, fmt.Errorf("open rows for first sheet: %w", err)
	}

	if !rows.Next() {
		return nil, nil, fmt.Errorf("xlsx sheet has no header row")
	}

	header, err := rows.Columns()
	if err != nil {
		return nil, nil, fmt.Errorf("read xlsx header: %w", err)
	}

	next := func() ([]string, error) {
		if !rows.Next() {
			if err := rows.Error(); err != nil {
				return nil, fmt.Errorf("iterate xlsx rows: %w", err)
			}
			if err := rows.Close(); err != nil {
				return nil, fmt.Errorf("close xlsx rows: %w", err)
			}
			return nil, io.EOF
		}

		record, err := rows.Columns()
		if err != nil {
			return nil, fmt.Errorf("read xlsx row: %w", err)
		}
		return record, nil
	}

	return header, next, nil
}
