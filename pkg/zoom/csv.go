package zoom

import (
	"encoding/csv"
	"io"
)

const maxAllocations = 200

var header = []string{"Pre-assign Room Name", "Email Address"}

func WriteCsv(out io.Writer, allocations [][]string) ([][]string, error) {
	w := csv.NewWriter(out)

	if err := w.Write(header); err != nil {
		return nil, err
	}

	for _, allocation := range allocations[:maxAllocations] {
		if err := w.Write(allocation); err != nil {
			return nil, err
		}
	}

	w.Flush()

	return allocations[maxAllocations:], nil
}
