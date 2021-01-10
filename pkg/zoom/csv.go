package zoom

import (
	"encoding/csv"
	"io"
)

var header = []string{"Pre-assign Room Name", "Email Address"}

func WriteCsv(out io.Writer, allocations [][]string) error {
	w := csv.NewWriter(out)

	if err := w.Write(header); err != nil {
		return err
	}

	return w.WriteAll(allocations)
}
