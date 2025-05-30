package main

import (
	"io"
	"os"
	"path/filepath"
)

func fullPath(n string) string {
	return filepath.Join(".", "images") + "/" + n
}

func SaveImage(src io.Reader, name string) error {
	fd, err := os.OpenFile(fullPath(name), os.O_WRONLY|os.O_CREATE, os.ModePerm)

	if err != nil {
		return err
	}

	defer fd.Close()

	_, err = io.Copy(fd, src)

	if err != nil {
		return err
	}

	return nil
}

func DeleteImage(name string) error {
	err := os.Remove(fullPath(name))

	if err != nil {
		return err
	}

	return nil
}
