package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// Sha256 computes the SHA256 hash of content and returns it as a hex string.
func Sha256(content []byte) string {
	hash := sha256.Sum256(content)
	return hex.EncodeToString(hash[:])
}

// CompressZlib is a placeholder for Zlib compression.
// TODO: Implement actual Zlib compression if ContentAddressable.Serialize is used.
func CompressZlib(data []byte) []byte {
	// fmt.Println("Warning: CompressZlib is a placeholder and does not actually compress.")
	return data
}

// DecompressZlib is a placeholder for Zlib decompression.
// TODO: Implement actual Zlib decompression if ContentAddressable.Deserialize is used.
func DecompressZlib(data []byte) []byte {
	// fmt.Println("Warning: DecompressZlib is a placeholder and does not actually decompress.")
	return data
}

const NULL_BYTE byte = 0x00
const CONTENT_ADDRESSABLE_HEADER_SEPARATOR = NULL_BYTE

type ContentAddressable struct {
	address *ContentAddress
	content []byte
}

type ContentAddress struct {
	contentType string
	hash        string
}

func NewContentAddress(contentType string, hash string) *ContentAddress {
	return &ContentAddress{
		contentType: contentType,
		hash:        hash,
	}
}

func (ca *ContentAddress) ContentType() string {
	return ca.contentType
}

func (ca *ContentAddress) Hash() string {
	return ca.hash
}

func NewContentAddressable(contentType string, content []byte) *ContentAddressable {
	return &ContentAddressable{
		address: &ContentAddress{
			contentType: contentType,
			hash:        Sha256(content),
		},
		content: content,
	}
}

func (ca *ContentAddressable) Address() *ContentAddress {
	return ca.address
}

func (ca *ContentAddressable) Content() []byte {
	return ca.content
}

func (ca *ContentAddressable) Serialize() []byte {
	var header = fmt.Sprintf("%s %d", ca.address.contentType, len(ca.content))
	var headerBytes = []byte(header)
	var contentWithHeader = append(append(headerBytes, CONTENT_ADDRESSABLE_HEADER_SEPARATOR), ca.content...)

	return CompressZlib(contentWithHeader)
}

func DeserializeContentAddressable(serialized []byte) (*ContentAddressable, error) {
	var contentWithHeader = DecompressZlib(serialized)
	var headerAndContent = bytes.SplitN(contentWithHeader, []byte{CONTENT_ADDRESSABLE_HEADER_SEPARATOR}, 2)
	if len(headerAndContent) != 2 {
		return nil, fmt.Errorf("invalid content addressable format")
	}

	var header = string(headerAndContent[0])
	var content = headerAndContent[1]

	var contentType string
	var contentLength int
	_, err := fmt.Sscanf(header, "%s %d", &contentType, &contentLength)
	if err != nil {
		return nil, fmt.Errorf("invalid content addressable header: %v", err)
	}

	if contentLength != len(content) {
		return nil, fmt.Errorf("invalid content addressable length")
	}

	// Re-hash the deserialized content to ensure integrity and that the hash field
	// in the ContentAddress is correct, rather than trusting the serialized hash.
	// The ContentAddress.hash effectively serves as an expected hash if it were
	// pre-computed and passed in, but NewContentAddressable recalculates it anyway.
	// For Deserialize, we should trust the content and recalculate its hash.
	return &ContentAddressable{
		address: &ContentAddress{
			contentType: contentType,
			hash:        Sha256(content), // Recalculate hash from deserialized content
		},
		content: content,
	}, nil
}
