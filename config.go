package main

import (
	"encoding/json"
	"fmt"
	"io"
)

type Updates map[string][]string

func GetUpdates(r io.Reader) (Updates, error) {
	updates := new(Updates)
	err := readFile(r, updates)
	if err != nil {
		return Updates{}, fmt.Errorf("can't decode Targets file: %w", err)
	}
	return *updates, nil
}

func readFile(r io.Reader, object any) error {
	d := json.NewDecoder(r)

	err := d.Decode(object)
	if err != nil {
		return fmt.Errorf("can't decode object: %w", err)
	}
	return nil
}
