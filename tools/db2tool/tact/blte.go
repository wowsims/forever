// Go translation of TACTSharp's BLTE decoder (https://github.com/wowdev/TACTSharp,
// v0.0.13-alpha, commit d0ab516eb98b5db35682467b6e4977d88955046d).
// Copyright (c) 2024 Martin Benjamins. MIT License — see tools/db2tool/NOTICES.md.
//
// 'E' (encrypted) chunks are Salsa20-decrypted when the KeyStore holds the
// chunk's key and left zero-filled otherwise, exactly what the WDC layer's
// encrypted-section skip expects. 'F' never occurs.

package tact

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
)

// Decodes a BLTE blob. totalDecompSize may be 0 (computed from chunk infos
// for multi-chunk files; required for single-chunk non-'N').
func blteDecode(data []byte, totalDecompSize uint64, keys *KeyStore) ([]byte, error) {
	if len(data) < 8 || data[0] != 'B' || data[1] != 'L' || data[2] != 'T' || data[3] != 'E' {
		return nil, fmt.Errorf("invalid BLTE header")
	}
	const fixedHeaderSize = 8
	headerSize := int(be32(data[4:]))

	if headerSize == 0 {
		mode := data[fixedHeaderSize]
		if mode != 'N' && totalDecompSize == 0 {
			return nil, fmt.Errorf("totalDecompSize must be set for single non-normal BLTE block")
		}
		if mode == 'N' && totalDecompSize == 0 {
			totalDecompSize = uint64(len(data) - fixedHeaderSize - 1)
		}
		out := make([]byte, totalDecompSize)
		if err := handleDataBlock(mode, data[fixedHeaderSize+1:], out, 0, keys); err != nil {
			return nil, err
		}
		return out, nil
	}

	if data[fixedHeaderSize] != 0xF {
		return nil, fmt.Errorf("unexpected BLTE table format 0x%x", data[fixedHeaderSize])
	}
	const blockInfoSize = 24
	chunkCount := int(data[fixedHeaderSize+1])<<16 | int(data[fixedHeaderSize+2])<<8 | int(data[fixedHeaderSize+3])
	infoStart := fixedHeaderSize + 4

	if totalDecompSize == 0 {
		o := infoStart + 4
		for range chunkCount {
			totalDecompSize += uint64(be32(data[o:]))
			o += blockInfoSize
		}
	}

	out := make([]byte, totalDecompSize)
	infoOffset := infoStart
	compOffset := headerSize
	decompOffset := 0

	for chunk := range chunkCount {
		compSize := int(be32(data[infoOffset:]))
		decompSize := int(be32(data[infoOffset+4:]))
		if compOffset+compSize > len(data) || decompOffset+decompSize > len(out) {
			return nil, fmt.Errorf("BLTE chunk %d out of bounds", chunk)
		}
		if err := handleDataBlock(data[compOffset], data[compOffset+1:compOffset+compSize], out[decompOffset:decompOffset+decompSize], chunk, keys); err != nil {
			return nil, fmt.Errorf("BLTE chunk %d: %w", chunk, err)
		}
		infoOffset += blockInfoSize
		compOffset += compSize
		decompOffset += decompSize
	}
	return out, nil
}

func handleDataBlock(mode byte, compData, out []byte, chunkIndex int, keys *KeyStore) error {
	switch mode {
	case 'N':
		copy(out, compData)
		return nil
	case 'Z':
		zr, err := zlib.NewReader(bytes.NewReader(compData))
		if err != nil {
			return err
		}
		defer zr.Close()
		_, err = io.ReadFull(zr, out)
		return err
	case 'E':
		// Layout: keyNameSize(1) keyName(8, LE) ivSize(1) iv(ivSize) cipher(1)
		// payload. The nonce is the IV padded to 8 bytes with the chunk index
		// folded into its first 4. The decrypted payload is itself a chunk
		// (mode byte + data), so it goes back through this switch.
		if len(compData) < 1 || compData[0] != 8 || len(compData) < 10 {
			return fmt.Errorf("malformed encrypted BLTE chunk header")
		}
		keyName := binary.LittleEndian.Uint64(compData[1:9])
		ivSize := int(compData[9])
		if ivSize > 8 || len(compData) < 10+ivSize+1 {
			return fmt.Errorf("malformed encrypted BLTE chunk IV")
		}
		key, ok := keys.lookup(keyName)
		if !ok {
			// No key: leave the output range zero-filled.
			return nil
		}
		cipher := compData[10+ivSize]
		if cipher != 'S' {
			return fmt.Errorf("BLTE chunk cipher %q not supported (only Salsa20 'S')", cipher)
		}
		var nonce [8]byte
		copy(nonce[:], compData[10:10+ivSize])
		for i := range 4 {
			nonce[i] ^= byte(chunkIndex >> (8 * i))
		}
		plain := salsa20XOR(key, nonce, compData[11+ivSize:])
		if len(plain) == 0 {
			return fmt.Errorf("empty encrypted BLTE chunk")
		}
		return handleDataBlock(plain[0], plain[1:], out, chunkIndex, keys)
	case 'F':
		return fmt.Errorf("BLTE frame ('F') decompression not implemented (never occurs in this data)")
	default:
		return fmt.Errorf("invalid BLTE chunk mode %q", mode)
	}
}

func be32(b []byte) uint32 {
	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
}
