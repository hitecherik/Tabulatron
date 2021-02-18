package util

import (
	"strconv"

	"github.com/andersfylling/disgord"
)

func StringToSnowflake(str string) (disgord.Snowflake, error) {
	snowflake, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return 0, err
	}

	return disgord.NewSnowflake(snowflake), nil
}

func StringsToSnowflakes(strs []string) ([]disgord.Snowflake, error) {
	snowflakes := make([]disgord.Snowflake, 0, len(strs))
	for _, discord := range strs {
		snowflake, err := StringToSnowflake(discord)
		if err != nil {
			return nil, err
		}

		snowflakes = append(snowflakes, snowflake)
	}

	return snowflakes, nil
}
