package params

import (
	"bytes"
	"compress/gzip"
	"fmt"
)

// Compress data using gzip
func compress(data []byte) ([]byte, error) {
	var b bytes.Buffer
	// writer with compression
	w, err := gzip.NewWriterLevel(&b, gzip.BestCompression)
	if err != nil {
		return nil, fmt.Errorf("Failed init compress writer: %v", err)
	}

	// compressing
	_, err = w.Write(data)
	if err != nil {
		return nil, fmt.Errorf("Failed write data to compress temporary buffer: %v", err)
	}

	err = w.Close()
	if err != nil {
		return nil, fmt.Errorf("Failed compress data: %v", err)
	}

	return b.Bytes(), nil
}

// Decompress data using gzip
func decompress(data []byte) ([]byte, error) {
	// reader for compressed data - readed data is decompressed
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("Failed init decompress reader: %v", err)
	}

	defer r.Close()

	var b bytes.Buffer
	// read and decompress
	_, err = b.ReadFrom(r)
	if err != nil {
		return nil, fmt.Errorf("Failed decompress data: %v", err)
	}

	return b.Bytes(), nil
}
