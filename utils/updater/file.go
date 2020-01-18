package updater

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func MakeDir(filename string) (dbdir string, err error) {
	dbdir = filepath.Dir(filename)
	_, err = os.Stat(dbdir)
	if err != nil {
		err = os.MkdirAll(dbdir, 0755)
		if err != nil {
			return "", err
		}
	}
	return dbdir, nil
}

func RenameFile(fromName string, toName string) error {
	if err := os.Rename(toName, toName+".bak"); err != nil {}
	if _, err := MakeDir(toName); err != nil {
		return err
	}
	return os.Rename(fromName, toName)
}

func (c *Config) ProcessFile() (error, string) {
	f, err := os.Open(c.Archive)
	if err != nil {
		return err, ""
	}
	defer f.Close()

	err, _ = c.ExtractTarGz(f)
	if err != nil {
		err, _ = c.ExtractZip()
	}
	return err, c.File
}

func (c *Config) ExtractZip() (error, string) {
	r, err := zip.OpenReader(c.Archive)
	if err != nil {
		return err, ""
	}
	defer func() {
		if err := r.Close(); err != nil {}
	}()

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				// panic(err)
			}
		}()

		if f.FileInfo().IsDir() == false {
			if strings.Contains(f.Name, "mmdb") || strings.Contains(f.Name, "BIN") {
				outFile, err := os.Create(c.File)
				if err != nil {
					return err
				}
				defer outFile.Close()
				if _, err := io.Copy(outFile, rc); err != nil {
					return err
				}else{
					return nil
				}
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			break
		}
	}

	return nil, c.File
}

func (c *Config) ExtractTarGz(gzipStream io.Reader) (error, string) {
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		return err, ""
	}

	tarReader := tar.NewReader(uncompressedStream)

	for true {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return err, ""
		}

		switch header.Typeflag {
		case tar.TypeDir:
			break;
		case tar.TypeReg:
			if strings.Contains(header.Name, "mmdb") || strings.Contains(header.Name, "BIN") {
				outFile, err := os.Create(c.File)
				if err != nil {
					return err, ""
				}
				defer outFile.Close()
				if _, err := io.Copy(outFile, tarReader); err != nil {
					return err, ""
				}else{
					return nil, c.File
				}
			}
		default:
			log.Fatalf(
				"ExtractTarGz: uknown type: %b in %s", header.Typeflag, header.Name)
		}
	}

	return nil, ""
}
