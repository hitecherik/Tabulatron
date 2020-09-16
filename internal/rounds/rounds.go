package rounds

import (
	"fmt"
	"strconv"
	"strings"
)

type Rounds []uint64

func (rs *Rounds) String() string {
	ids := make([]string, 0, len(*rs))

	for _, r := range *rs {
		ids = append(ids, fmt.Sprintf("%v", r))
	}

	return strings.Join(ids, ",")
}

func (rs *Rounds) Set(s string) error {
	round, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return err
	}

	*rs = append(*rs, round)
	return nil
}
