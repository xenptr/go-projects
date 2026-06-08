package main

import (
	"encoding/json"
	"errors"
	"os"
)

func LoadTasks() error {
	file, err := os.ReadFile("tasks.json")
	if errors.Is(err, os.ErrNotExist) {
		tasks = []Task{}
		return nil
	}
	if err != nil {
		return err
	}

	if len(file) == 0 {
		tasks = []Task{}
		return nil
	}

	return json.Unmarshal(file, &tasks)
}

func SaveTasks() error {
	file, err := os.OpenFile("tasks.json", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := json.MarshalIndent(tasks, "", " ")
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	return err
}
