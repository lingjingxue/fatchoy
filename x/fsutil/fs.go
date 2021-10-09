// Copyright 2013 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package fsutil

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// 把文件内容按一行一行读取
func ReadFileToLines(filename string) ([]string, error) {
	fd, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	return ReadToLines(fd)
}

func ReadToLines(rd io.Reader) ([]string, error) {
	var lines []string
	var scanner = bufio.NewScanner(rd)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

// EnsureBaseDir ensures that path is always prefixed by baseDir,
// allowing for the fact that path might have a Window drive letter in
// it.
func EnsureBaseDir(baseDir, path string) string {
	if baseDir == "" {
		return path
	}
	volume := filepath.VolumeName(path)
	return filepath.Join(baseDir, path[len(volume):])
}

// UniqueDirectory returns "path/name" if that directory doesn't exist.  If it
// does, the method starts appending .1, .2, etc until a unique name is found.
func UniqueDirectory(path, name string) (string, error) {
	dir := filepath.Join(path, name)
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return dir, nil
	}
	for i := 1; ; i++ {
		dir := filepath.Join(path, fmt.Sprintf("%s.%d", name, i))
		_, err := os.Stat(dir)
		if os.IsNotExist(err) {
			return dir, nil
		} else if err != nil {
			return "", err
		}
	}
}

// CopyFile writes the contents of the given source file to dest.
func CopyFile(dest, source string) error {
	df, err := os.Create(dest)
	if err != nil {
		return err
	}
	f, err := os.Open(source)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(df, f)
	return err
}

// IsFileExist test if file exist
func IsFileExist(filename string) bool {
	_, err := os.Lstat(filename)
	return !os.IsNotExist(err)
}

// 压缩文件 my.log --> my.log.tar.gz
func ArchiveGzipFile(srcFile string) error {
	var filename = fmt.Sprintf("%s.tar.gz", srcFile)
	outf, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer outf.Close()
	var gw = gzip.NewWriter(outf)
	defer gw.Close()
	var tw = tar.NewWriter(gw)
	defer tw.Close()

	f, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return err
	}
	header, err := tar.FileInfoHeader(fi, "")
	if err != nil {
		return err
	}
	header.Name = srcFile
	if err := tw.WriteHeader(header); err != nil {
		return err
	}
	if _, err = io.Copy(tw, f); err != nil {
		return err
	}
	return nil
}
