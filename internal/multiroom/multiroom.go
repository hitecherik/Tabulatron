package multiroom

import (
	"fmt"
	"log"
	"strings"

	"github.com/pelletier/go-toml"
)

type Category struct {
	Name   string
	Prefix string
	Suffix string
	Url    string
}

type Categories []Category

type rawCategories struct {
	Category Categories
}

func (cs *Categories) String() string {
	raw, err := toml.Marshal(rawCategories{*cs})
	if err != nil {
		log.Fatalf("could not marshal Categories: %v", err)
	}

	return string(raw)
}

func (cs *Categories) Set(path string) error {
	tree, err := toml.LoadFile(path)
	if err != nil {
		return err
	}

	var raw rawCategories
	if err := tree.Unmarshal(&raw); err != nil {
		return err
	}

	*cs = raw.Category
	return nil
}

func (cs *Categories) Lookup(name string) (Category, error) {
	if len(*cs) == 0 {
		return Category{}, nil
	}

	for _, category := range *cs {
		if strings.HasPrefix(name, category.Prefix) && strings.HasSuffix(name, category.Suffix) {
			return category, nil
		}
	}

	return Category{}, fmt.Errorf("no matching category found for room \"%v\"", name)
}
