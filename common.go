package main

import (
	"io/ioutil"
	"path/filepath"
)

func messagesInMaildir(path string) int {

	dirs := []string{"cur", "new"}

	sum := 0

	for _, ent := range dirs {
		known, _ := ioutil.ReadDir(filepath.Join(path, ent))
		sum += len(known)
	}

	return sum
}

func unreadMessagesInMaildir(path string) int {

	path = filepath.Join(path, "new")

	files, _ := ioutil.ReadDir(path)
	return (len(files))
}
