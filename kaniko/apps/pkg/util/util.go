package util

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

func SplitBeforeAfter(s string, splitChar string) (string, string) {
	var splitBefore, splitAfter string
	if i := strings.Index(s, splitChar); i >= 0 {
		splitBefore, splitAfter = s[:i], s[i+1:]
	} else {
		log.Fatal("Unhandled condition in string split")
	}
	return splitBefore, splitAfter
}

func CreateArchive(files []string, buf io.Writer, header string) error {
	// Create new Writers for gzip and tar
	// These writers are chained. Writing to the tar writer will
	// write to the gzip writer which in turn will write to
	// the "buf" writer
	gw := gzip.NewWriter(buf)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Iterate over files and add them to the tar archive
	for _, file := range files {
		err := addToArchive(tw, file, header)
		if err != nil {
			return err
		}
	}

	return nil
}

func addToArchive(tw *tar.Writer, filename, headerName string) error {
	// Open the file which will be written into the archive
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Get FileInfo about our file providing file size, mode, etc.
	info, err := file.Stat()
	if err != nil {
		return err
	}

	// Create a tar Header from the FileInfo data
	header, err := tar.FileInfoHeader(info, info.Name())
	if err != nil {
		return err
	}

	// Use full path as name (FileInfoHeader only takes the basename)
	// If we don't do this the directory structure would
	// not be preserved
	// https://golang.org/src/archive/tar/common.go?#L626
	tarName := strings.Split(filename, "/")
	finalTarName := tarName[len(tarName)-1]
	header.Name = headerName + "/" + finalTarName

	// Write file header to the tar archive
	err = tw.WriteHeader(header)
	if err != nil {
		return err
	}

	// Copy file content to tar archive
	_, err = io.Copy(tw, file)
	if err != nil {
		return err
	}

	return nil
}

func ExtractArchive(fileName string) {

	file, err := os.Open(fileName)

	if err != nil {
		log.Info(err)
		os.Exit(1)
	}
	defer file.Close()

	var fileReader io.ReadCloser = file
	tarBallReader := tar.NewReader(fileReader)

	// Extracting tarred files

	for {
		header, err := tarBallReader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Info(err)
			os.Exit(1)
		}

		// get the individual filename and extract to the current directory
		filename := header.Name

		switch header.Typeflag {
		case tar.TypeDir:
			// handle directory
			log.Info("Creating directory :", filename)
			err = os.MkdirAll(filename, os.FileMode(header.Mode)) // or use 0755 if you prefer

			if err != nil {
				log.Info(err)
				os.Exit(1)
			}

		case tar.TypeReg:
			// handle normal file
			log.Info("Untarring :", filename)
			writer, err := os.Create(filename)

			if err != nil {
				log.Info(err)
				os.Exit(1)
			}

			io.Copy(writer, tarBallReader)

			err = os.Chmod(filename, os.FileMode(header.Mode))

			if err != nil {
				log.Info(err)
				os.Exit(1)
			}

			writer.Close()
		default:
			fmt.Printf("Unable to untar type : %c in file %s", header.Typeflag, filename)
		}
	}
}

func iterateFilesInDir(path string) {
	items, _ := ioutil.ReadDir(path)
	for _, item := range items {
		if item.IsDir() {
			subitems, _ := ioutil.ReadDir(item.Name())
			for _, subitem := range subitems {
				if !subitem.IsDir() {
					// handle file there
					log.Info(item.Name() + "/" + subitem.Name())
				}
			}
		} else {
			log.Info(item.Name())
		}
	}
}

// mkdir, like os.Mkdir, creates a new directory
// with the specified name and permission bits.
// If the directory exists, mkdir returns nil.
func Mkdir(name string, perm os.FileMode) error {
	err := os.MkdirAll(name, perm)
	if err != nil {
		if !os.IsExist(err) {
			return err
		}
	}
	return nil
}

func ReadFileLineByLine(fileName string) []string {
	var list []string
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("failed to open")
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var text []string
	for scanner.Scan() {
		text = append(text, scanner.Text())
	}
	file.Close()
	for _, line := range text {
		list = append(list, line)
	}
	return list
}

func WriteToFile(inputData [] string, destinationFile string) {
	file, err := os.OpenFile(destinationFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}
	datawriter := bufio.NewWriter(file)
	for _, data := range inputData {
		_, _ = datawriter.WriteString(data + "\n")
	}
	datawriter.Flush()
	file.Close()
}
