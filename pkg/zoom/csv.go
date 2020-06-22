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

	numAllocations := min(maxAllocations, len(allocations))

	for _, allocation := range allocations[:numAllocations] {
		if err := w.Write(allocation); err != nil {
			return nil, err
		}
	}

	w.Flush()

	return allocations[numAllocations:], nil
}

func min(x int, y int) int {
	if x > y {
		return y
	}

	return x
}
