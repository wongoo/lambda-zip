package main

import (
	"archive/zip"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/urfave/cli.v1"
	"io"
	"strings"
)

func main() {
	app := cli.NewApp()
	app.Name = "gozip"
	app.Usage = "Put an executable and others into a zip file"

	app.Action = func(c *cli.Context) error {
		if !c.Args().Present() {
			return errors.New("No input provided")
		}

		outputZip := c.Args().First()
		if !strings.HasSuffix(outputZip, ".zip") {
			return errors.New("output file is not zip format")
		}

		files := c.Args().Tail()
		if len(files) < 1 {
			return errors.New("No input files")
		}

		if err := compressFiles(outputZip, files); err != nil {
			return fmt.Errorf("Failed to compress file: %v", err)
		}
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func compressFiles(outZipPath string, files []string) error {
	zipFile, err := os.Create(outZipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	return compressZipFiles(zipFile, files)
}

func compressZipFiles(zipFile *os.File, files []string) (err error) {
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	exeFile := files[0]
	err = compressFile(zipWriter, exeFile, true)
	if err != nil {
		return
	}

	if len(files) > 1 {
		for _, f := range files[1:] {
			err = compressFile(zipWriter, f, false)
			if err != nil {
				return err
			}
		}
	}
	return
}

func compressFile(zipWriter *zip.Writer, file string, exe bool) (err error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	pathInZip := filepath.Base(file)

	var fw io.Writer

	if exe {
		fw, err = zipWriter.CreateHeader(&zip.FileHeader{
			CreatorVersion: 3 << 8,     // indicates Unix
			ExternalAttrs:  0777 << 16, // -rwxrwxrwx file permissions
			Name:           pathInZip,
			Method:         zip.Deflate,
		})
	} else {
		fw, err = zipWriter.CreateHeader(&zip.FileHeader{
			Name:   pathInZip,
			Method: zip.Deflate,
		})
	}

	if err != nil {
		return err
	}

	_, err = fw.Write(data)
	return err
}
