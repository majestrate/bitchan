package util

import (
	"io"
	"os"
)

// Move a file from src to dst
func Move(src, dst string) error {

	s_file, err := os.Open(src)
	if err != nil {
		return err
	}
	d_file, err := os.Create(dst)
	if err != nil {
		s_file.Close()
		return err
	}
	defer func() {
		d_file.Sync()
		d_file.Close()
		s_file.Close()
		os.Remove(src)
	}()
	_, err = io.Copy(d_file, s_file)
	return err
}
