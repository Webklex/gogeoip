package db

import (
	"archive/tar"
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

func (d *DB) processFile() (error, string) {
	f, err := os.Open(d.dbArchive)
	if err != nil {
		return err, ""
	}
	defer f.Close()

	return d.ExtractTarGz(f)
}

func (d *DB) ExtractTarGz(gzipStream io.Reader) (error, string) {
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		log.Fatal("ExtractTarGz: NewReader failed")
		return err, ""
	}

	tarReader := tar.NewReader(uncompressedStream)

	for true {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("ExtractTarGz: Next() failed: %s", err.Error())
		}

		switch header.Typeflag {
		case tar.TypeDir:
			break;
		case tar.TypeReg:
			if strings.Contains(header.Name, "mmdb") {
				outFile, err := os.Create(d.File)
				if err != nil {
					log.Fatalf("ExtractTarGz: Create() failed: %s", err.Error())
				}
				defer outFile.Close()
				if _, err := io.Copy(outFile, tarReader); err != nil {
					log.Fatalf("ExtractTarGz: Copy() failed: %s", err.Error())
				}else{
					return nil, d.File
				}
			}
		default:
			log.Fatalf(
				"ExtractTarGz: uknown type: %b in %s", header.Typeflag, header.Name)
		}
	}

	return nil, ""
}
