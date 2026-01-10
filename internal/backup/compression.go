package backup

import (
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"os"
)

// Compressor handles compression of backup data with checksum calculation.
type Compressor struct {
	compression string
	level       int
}

// NewCompressor creates a new Compressor.
func NewCompressor(compression string) *Compressor {
	return &Compressor{
		compression: compression,
		level:       gzip.DefaultCompression,
	}
}

// NewCompressorWithLevel creates a new Compressor with specific compression level.
func NewCompressorWithLevel(compression string, level int) *Compressor {
	return &Compressor{
		compression: compression,
		level:       level,
	}
}

// CompressResult holds the result of compression operation.
type CompressResult struct {
	BytesRead    int64
	BytesWritten int64
	Checksum     string
}

// Compress compresses data from reader to writer, calculating checksum during compression.
// Returns the number of bytes read, bytes written, and the SHA-256 checksum.
func (c *Compressor) Compress(reader io.Reader, writer io.Writer) (*CompressResult, error) {
	var bytesRead int64
	var bytesWritten int64

	// Create hash for checksum calculation
	hasher := sha256.New()

	// Create a multi-writer to calculate checksum while writing
	checksumReader := io.TeeReader(reader, hasher)

	switch c.compression {
	case CompressionGzip:
		result, err := c.compressGzip(checksumReader, writer)
		if err != nil {
			return nil, err
		}
		bytesRead = result.BytesRead
		bytesWritten = result.BytesWritten

	case CompressionNone:
		var err error
		bytesWritten, err = io.Copy(writer, checksumReader)
		if err != nil {
			return nil, WrapCompressionError("", "failed to copy data", err)
		}
		bytesRead = bytesWritten

	default:
		return nil, &CompressionError{
			Message: fmt.Sprintf("unsupported compression: %s", c.compression),
		}
	}

	// Calculate final checksum
	checksum := fmt.Sprintf("sha256:%x", hasher.Sum(nil))

	return &CompressResult{
		BytesRead:    bytesRead,
		BytesWritten: bytesWritten,
		Checksum:     checksum,
	}, nil
}

// compressGzip compresses data using gzip.
func (c *Compressor) compressGzip(reader io.Reader, writer io.Writer) (*CompressResult, error) {
	gzWriter, err := gzip.NewWriterLevel(writer, c.level)
	if err != nil {
		return nil, WrapCompressionError("", "failed to create gzip writer", err)
	}

	bytesRead, err := io.Copy(gzWriter, reader)
	if err != nil {
		gzWriter.Close()
		return nil, WrapCompressionError("", "failed to compress data", err)
	}

	if err := gzWriter.Close(); err != nil {
		return nil, WrapCompressionError("", "failed to close gzip writer", err)
	}

	// Get bytes written from the underlying writer if it's a counting writer
	// For now, we'll estimate based on file size after writing
	return &CompressResult{
		BytesRead:    bytesRead,
		BytesWritten: 0, // Will be calculated from file size
	}, nil
}

// CompressFile compresses a source file to a destination file with checksum.
func (c *Compressor) CompressFile(srcPath, dstPath string) (*CompressResult, error) {
	// Open source file
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return nil, WrapCompressionError(srcPath, "failed to open source file", err)
	}
	defer srcFile.Close()

	// Create destination file
	dstFile, err := os.Create(dstPath)
	if err != nil {
		return nil, WrapCompressionError(dstPath, "failed to create destination file", err)
	}
	defer dstFile.Close()

	// Compress
	result, err := c.Compress(srcFile, dstFile)
	if err != nil {
		return nil, err
	}

	// Get actual bytes written from file size
	fileInfo, err := dstFile.Stat()
	if err != nil {
		return nil, WrapCompressionError(dstPath, "failed to stat compressed file", err)
	}
	result.BytesWritten = fileInfo.Size()

	return result, nil
}

// StreamCompress compresses data from reader to a file, calculating checksum.
// This is the main method used for mysqldump streaming.
func (c *Compressor) StreamCompress(reader io.Reader, outputPath string) (*CompressResult, error) {
	// Create output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return nil, WrapCompressionError(outputPath, "failed to create output file", err)
	}
	defer outFile.Close()

	// Compress with checksum
	result, err := c.Compress(reader, outFile)
	if err != nil {
		return nil, err
	}

	// Get actual bytes written from file size
	fileInfo, err := outFile.Stat()
	if err != nil {
		return nil, WrapCompressionError(outputPath, "failed to stat compressed file", err)
	}
	result.BytesWritten = fileInfo.Size()

	return result, nil
}

// Decompressor handles decompression of backup data.
type Decompressor struct {
	compression string
}

// NewDecompressor creates a new Decompressor.
func NewDecompressor(compression string) *Decompressor {
	return &Decompressor{
		compression: compression,
	}
}

// Decompress decompresses data from reader to writer.
func (d *Decompressor) Decompress(reader io.Reader, writer io.Writer) (int64, error) {
	switch d.compression {
	case CompressionGzip:
		return d.decompressGzip(reader, writer)

	case CompressionNone:
		bytesWritten, err := io.Copy(writer, reader)
		if err != nil {
			return 0, WrapCompressionError("", "failed to copy data", err)
		}
		return bytesWritten, nil

	default:
		return 0, &CompressionError{
			Message: fmt.Sprintf("unsupported compression: %s", d.compression),
		}
	}
}

// decompressGzip decompresses gzip data.
func (d *Decompressor) decompressGzip(reader io.Reader, writer io.Writer) (int64, error) {
	gzReader, err := gzip.NewReader(reader)
	if err != nil {
		return 0, WrapCompressionError("", "failed to create gzip reader", err)
	}
	defer gzReader.Close()

	bytesWritten, err := io.Copy(writer, gzReader)
	if err != nil {
		return 0, WrapCompressionError("", "failed to decompress data", err)
	}

	return bytesWritten, nil
}

// DecompressFile decompresses a source file to a destination file.
func (d *Decompressor) DecompressFile(srcPath, dstPath string) (int64, error) {
	// Open source file
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return 0, WrapCompressionError(srcPath, "failed to open source file", err)
	}
	defer srcFile.Close()

	// Create destination file
	dstFile, err := os.Create(dstPath)
	if err != nil {
		return 0, WrapCompressionError(dstPath, "failed to create destination file", err)
	}
	defer dstFile.Close()

	// Decompress
	return d.Decompress(srcFile, dstFile)
}

// DecompressToReader decompresses data from a reader and returns a reader for the decompressed data.
// The returned reader must be closed by the caller.
func (d *Decompressor) DecompressToReader(reader io.Reader) (io.ReadCloser, error) {
	switch d.compression {
	case CompressionGzip:
		gzReader, err := gzip.NewReader(reader)
		if err != nil {
			return nil, WrapCompressionError("", "failed to create gzip reader", err)
		}
		return gzReader, nil

	case CompressionNone:
		// Return a no-op closer that just closes the reader if it's a ReadCloser
		return io.NopCloser(reader), nil

	default:
		return nil, &CompressionError{
			Message: fmt.Sprintf("unsupported compression: %s", d.compression),
		}
	}
}

// VerifyChecksum verifies the checksum of a compressed file.
func VerifyChecksum(filePath, expectedChecksum string) (bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return false, fmt.Errorf("failed to calculate checksum: %w", err)
	}

	actualChecksum := fmt.Sprintf("sha256:%x", hasher.Sum(nil))
	return actualChecksum == expectedChecksum, nil
}

// CountingWriter wraps a writer and counts bytes written.
type CountingWriter struct {
	writer       io.Writer
	bytesWritten int64
}

// NewCountingWriter creates a new CountingWriter.
func NewCountingWriter(writer io.Writer) *CountingWriter {
	return &CountingWriter{
		writer: writer,
	}
}

// Write writes data and counts bytes.
func (w *CountingWriter) Write(p []byte) (n int, err error) {
	n, err = w.writer.Write(p)
	w.bytesWritten += int64(n)
	return
}

// BytesWritten returns the number of bytes written.
func (w *CountingWriter) BytesWritten() int64 {
	return w.bytesWritten
}

// MultiWriter creates a writer that writes to multiple writers and calculates checksum.
type ChecksumMultiWriter struct {
	writers []io.Writer
	hasher  hash.Hash
}

// NewChecksumMultiWriter creates a new ChecksumMultiWriter.
func NewChecksumMultiWriter(writers ...io.Writer) *ChecksumMultiWriter {
	return &ChecksumMultiWriter{
		writers: writers,
		hasher:  sha256.New(),
	}
}

// Write writes to all writers and updates checksum.
func (w *ChecksumMultiWriter) Write(p []byte) (n int, err error) {
	// Write to hasher first
	w.hasher.Write(p)

	// Write to all writers
	for _, writer := range w.writers {
		n, err = writer.Write(p)
		if err != nil {
			return
		}
		if n != len(p) {
			err = io.ErrShortWrite
			return
		}
	}
	return len(p), nil
}

// Checksum returns the calculated checksum.
func (w *ChecksumMultiWriter) Checksum() string {
	return fmt.Sprintf("sha256:%x", w.hasher.Sum(nil))
}
