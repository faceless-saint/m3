package net

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"io/ioutil"
	"math"
	"os"
)

// NewHash returns a new hash implementing the chosen algorithm.
// Supported algorithms: "sha512", "sha256", "sha1", "md5", "git"
func NewHash(h string) (hash.Hash, error) {
	switch h {
	case "sha512":
		return sha512.New(), nil
	case "sha256":
		return sha256.New(), nil
	case "sha1":
		return sha1.New(), nil
	case "md5":
		return md5.New(), nil
	case "git":
		return &GitHash{}, nil
	default:
		return nil, fmt.Errorf("error: unsupported hash type %s", h)
	}
}

// DefualtHash returns the default hash using the SHA256 algorithm.
func DefaultHash() hash.Hash { return sha256.New() }

// ByteChecksum returns the hex-encoded checksum of the data bytes.
func ByteChecksum(data []byte, h hash.Hash) string {
	h.Write(data)
	checksum := fmt.Sprintf("%x", h.Sum([]byte{}))
	h.Reset()
	return checksum
}

// StringChecksum returns the hex-encoded checksum of the string.
func StringChecksum(str string, h hash.Hash) string {
	return ByteChecksum([]byte(str), h)
}

// FileChecksum returns the hex-encoded checksum of the file.
func FileChecksum(file string, h hash.Hash) (string, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}
	return ByteChecksum(data, h), nil
}

// Digest returns a short digest consisting of the first 'n' characters
// of the given string's checksum. The library default hash is used.
func Digest(str string, n int) string {
	return StringChecksum(str, DefaultHash())[:n]
}

// VerifyBytes compares the checksum of the input data against the
// reference checksum, returning true iff they match.
func VerifyBytes(data []byte, checksum string, h hash.Hash) bool {
	return ByteChecksum(data, h) == checksum
}

// VerifyFile validates the given file against a reference checksum. The
// file is deleted if the checksums do not match, unless the reference
// checksum is empty.
func VerifyFile(file, checksum string, h hash.Hash) error {
	sum, err := FileChecksum(file, h)
	if err != nil {
		return err
	}
	if len(checksum) > 0 {
		if sum != checksum {
			err := os.RemoveAll(file)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// ByteCountToString returns a human readable representation of the
// given raw byte count using SI units to compactly display the size.
func ByteCountToString(bytes uint64) string {
	var unit uint64
	unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	exp := uint64(math.Log2(float64(bytes)) / math.Log2(float64(unit)))
	char := string("kMGTPE"[exp-1])
	return fmt.Sprintf("%7.2f %sB", float64(bytes)/math.Pow(float64(unit), float64(exp)), char)
}

// GitHash is an implementation of hash.Hash for Git blob checksums. It
// corresponds to the NewHash function's "git" hash type.
type GitHash struct {
	size int
	data []byte
}

func (this *GitHash) Write(data []byte) (int, error) {
	this.data = append(this.data, data...)
	this.size = len(this.data)
	return len(data), nil
}
func (this *GitHash) Sum(data []byte) []byte {
	data = append(data, this.data...)
	prefix := fmt.Sprintf("blob %d\x00", len(data))
	raw := sha1.Sum(append([]byte(prefix), data...))
	return raw[:]
}
func (this *GitHash) Reset()         { this.data = []byte{}; this.size = 0 }
func (this *GitHash) Size() int      { return this.size }
func (this *GitHash) BlockSize() int { return 64 }
