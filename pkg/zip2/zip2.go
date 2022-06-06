package zip2

import (
	"archive/zip"
	"bytes"
	"context"
	"io/ioutil"
)

// FileInfo file information
type FileInfo struct {
	Name string
	Body []byte
}

// PackagingToBuffer Compress information to zip io stream
func PackagingToBuffer(ctx context.Context, infos []FileInfo) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	wZip := zip.NewWriter(buf)
	defer wZip.Close()
	for _, info := range infos {
		f, err := wZip.Create(info.Name)
		if err != nil {
			return nil, err
		}
		_, err = f.Write(info.Body)
		if err != nil {
			return nil, err
		}
	}

	return buf, nil
}

// DecompressToBytes Decompress zip information to bytes
func DecompressToBytes(ctx context.Context, zipBytes []byte) (map[string][]byte, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		return nil, err
	}
	result := map[string][]byte{}
	for _, file := range zipReader.File {

		io, err := file.Open()
		if err != nil {
			return nil, err
		}
		fileBytes, err := ioutil.ReadAll(io)
		if err != nil {
			return nil, err
		}
		result[file.Name] = fileBytes
	}
	return result, nil
}
