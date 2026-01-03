package backup

import (
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCompressor(t *testing.T) {
	compressor := NewCompressor(CompressionGzip)
	assert.NotNil(t, compressor)
	assert.Equal(t, CompressionGzip, compressor.compression)
}

func TestCompressGzip(t *testing.T) {
	compressor := NewCompressor(CompressionGzip)

	input := bytes.NewReader([]byte("Hello, World! This is a test of compression."))
	var output bytes.Buffer

	result, err := compressor.Compress(input, &output)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Greater(t, result.BytesRead, int64(0))
	assert.True(t, len(result.Checksum) > 0)
	assert.True(t, output.Len() > 0)

	// Output should be compressed (though small strings might not compress well)
	// Just check that we got output
	assert.Greater(t, output.Len(), 0)
}

func TestCompressNone(t *testing.T) {
	compressor := NewCompressor(CompressionNone)

	inputData := []byte("Hello, World!")
	input := bytes.NewReader(inputData)
	var output bytes.Buffer

	result, err := compressor.Compress(input, &output)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(len(inputData)), result.BytesRead)
	assert.Equal(t, inputData, output.Bytes())
}

func TestCompressUnsupported(t *testing.T) {
	compressor := NewCompressor("unsupported")

	input := bytes.NewReader([]byte("test"))
	var output bytes.Buffer

	_, err := compressor.Compress(input, &output)
	assert.Error(t, err)
	assert.True(t, IsCompressionError(err))
}

func TestCompressFile(t *testing.T) {
	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "source.txt")
	dstPath := filepath.Join(tmpDir, "compressed.gz")

	// Create source file
	content := []byte("Hello, World! This is a test of file compression.")
	err := os.WriteFile(srcPath, content, 0644)
	require.NoError(t, err)

	// Compress
	compressor := NewCompressor(CompressionGzip)
	result, err := compressor.CompressFile(srcPath, dstPath)
	require.NoError(t, err)

	assert.Equal(t, int64(len(content)), result.BytesRead)
	assert.Greater(t, result.BytesWritten, int64(0))
	assert.NotEmpty(t, result.Checksum)

	// Compressed file should exist
	assert.True(t, PathExists(dstPath))

	// Verify it's valid gzip
	file, err := os.Open(dstPath)
	require.NoError(t, err)
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	require.NoError(t, err)
	defer gzReader.Close()

	decompressed, err := io.ReadAll(gzReader)
	require.NoError(t, err)
	assert.Equal(t, content, decompressed)
}

func TestStreamCompress(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output.gz")

	content := []byte("This is test data for stream compression.")
	reader := bytes.NewReader(content)

	compressor := NewCompressor(CompressionGzip)
	result, err := compressor.StreamCompress(reader, outputPath)
	require.NoError(t, err)

	assert.Equal(t, int64(len(content)), result.BytesRead)
	assert.Greater(t, result.BytesWritten, int64(0))
	assert.NotEmpty(t, result.Checksum)

	// File should exist and be valid gzip
	file, err := os.Open(outputPath)
	require.NoError(t, err)
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	require.NoError(t, err)
	defer gzReader.Close()

	decompressed, err := io.ReadAll(gzReader)
	require.NoError(t, err)
	assert.Equal(t, content, decompressed)
}

func TestDecompressor(t *testing.T) {
	// First compress some data
	compressor := NewCompressor(CompressionGzip)
	content := []byte("Test data for decompression")
	var compressed bytes.Buffer

	_, err := compressor.Compress(bytes.NewReader(content), &compressed)
	require.NoError(t, err)

	// Now decompress it
	decompressor := NewDecompressor(CompressionGzip)
	var decompressed bytes.Buffer

	bytesWritten, err := decompressor.Decompress(&compressed, &decompressed)
	require.NoError(t, err)
	assert.Equal(t, int64(len(content)), bytesWritten)
	assert.Equal(t, content, decompressed.Bytes())
}

func TestDecompressFile(t *testing.T) {
	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "source.txt")
	compressedPath := filepath.Join(tmpDir, "compressed.gz")
	decompressedPath := filepath.Join(tmpDir, "decompressed.txt")

	// Create and compress a file
	content := []byte("Test data for file decompression")
	err := os.WriteFile(srcPath, content, 0644)
	require.NoError(t, err)

	compressor := NewCompressor(CompressionGzip)
	_, err = compressor.CompressFile(srcPath, compressedPath)
	require.NoError(t, err)

	// Decompress it
	decompressor := NewDecompressor(CompressionGzip)
	bytesWritten, err := decompressor.DecompressFile(compressedPath, decompressedPath)
	require.NoError(t, err)
	assert.Equal(t, int64(len(content)), bytesWritten)

	// Verify decompressed content matches original
	decompressedContent, err := os.ReadFile(decompressedPath)
	require.NoError(t, err)
	assert.Equal(t, content, decompressedContent)
}

func TestVerifyChecksum(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")

	// Create file
	content := []byte("Test content for checksum verification")
	err := os.WriteFile(filePath, content, 0644)
	require.NoError(t, err)

	// Calculate checksum
	checksum, err := CalculateChecksum(filePath)
	require.NoError(t, err)

	// Verify with correct checksum
	valid, err := VerifyChecksum(filePath, checksum)
	require.NoError(t, err)
	assert.True(t, valid)

	// Verify with incorrect checksum
	valid, err = VerifyChecksum(filePath, "sha256:incorrect")
	require.NoError(t, err)
	assert.False(t, valid)
}

func TestCountingWriter(t *testing.T) {
	var buf bytes.Buffer
	counter := NewCountingWriter(&buf)

	// Write some data
	data1 := []byte("Hello, ")
	n, err := counter.Write(data1)
	require.NoError(t, err)
	assert.Equal(t, len(data1), n)
	assert.Equal(t, int64(len(data1)), counter.BytesWritten())

	// Write more data
	data2 := []byte("World!")
	n, err = counter.Write(data2)
	require.NoError(t, err)
	assert.Equal(t, len(data2), n)
	assert.Equal(t, int64(len(data1)+len(data2)), counter.BytesWritten())

	// Verify actual content
	assert.Equal(t, "Hello, World!", buf.String())
}

func TestChecksumMultiWriter(t *testing.T) {
	var buf1 bytes.Buffer
	var buf2 bytes.Buffer

	multiWriter := NewChecksumMultiWriter(&buf1, &buf2)

	// Write data
	data := []byte("Test data for multi-writer")
	n, err := multiWriter.Write(data)
	require.NoError(t, err)
	assert.Equal(t, len(data), n)

	// Both buffers should have the data
	assert.Equal(t, data, buf1.Bytes())
	assert.Equal(t, data, buf2.Bytes())

	// Checksum should be calculated
	checksum := multiWriter.Checksum()
	assert.True(t, len(checksum) > 0)
	assert.Contains(t, checksum, "sha256:")
}

func TestCompressDecompressRoundTrip(t *testing.T) {
	originalData := []byte("This is a test of compression and decompression round trip.")

	// Compress
	compressor := NewCompressor(CompressionGzip)
	var compressed bytes.Buffer
	result, err := compressor.Compress(bytes.NewReader(originalData), &compressed)
	require.NoError(t, err)
	assert.Equal(t, int64(len(originalData)), result.BytesRead)

	// Decompress
	decompressor := NewDecompressor(CompressionGzip)
	var decompressed bytes.Buffer
	_, err = decompressor.Decompress(&compressed, &decompressed)
	require.NoError(t, err)

	// Should match original
	assert.Equal(t, originalData, decompressed.Bytes())
}

func TestCompressLargeData(t *testing.T) {
	// Create a large buffer (10MB of repeated text)
	largeData := bytes.Repeat([]byte("This is a test line that will be repeated many times.\n"), 200000)

	compressor := NewCompressor(CompressionGzip)
	var compressed bytes.Buffer

	result, err := compressor.Compress(bytes.NewReader(largeData), &compressed)
	require.NoError(t, err)
	assert.Equal(t, int64(len(largeData)), result.BytesRead)

	// Compressed should be significantly smaller for repeated data
	compressionRatio := float64(compressed.Len()) / float64(len(largeData))
	assert.Less(t, compressionRatio, 0.1) // Should compress to less than 10%
}
