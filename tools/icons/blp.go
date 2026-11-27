package main

import (
	"encoding/binary"
	"fmt"
	"image"
)

// The fixed BLP2 header: magic, version, encoding, alpha size, alpha type, mip flag, width, height, 16 mip
// offsets and 16 mip sizes, then a 256-entry BGRA palette that every file carries.
const (
	blpHeaderSize  = 148
	blpPaletteSize = 1024
)

// The largest mip of a BLP2 texture, the format of the game files' icons. It reads the three encodings
// icons ship in: a palette with 0, 1, 4 or 8 alpha bits per pixel, DXT1/3/5 blocks, and raw BGRA.
func decodeBLP(data []byte) (*image.NRGBA, error) {
	if len(data) < blpHeaderSize+blpPaletteSize || string(data[:4]) != "BLP2" {
		return nil, fmt.Errorf("not a BLP2 file")
	}
	encoding, alphaSize, alphaType := data[8], data[9], data[10]
	w := int(binary.LittleEndian.Uint32(data[12:]))
	h := int(binary.LittleEndian.Uint32(data[16:]))
	offset := int(binary.LittleEndian.Uint32(data[20:]))
	size := int(binary.LittleEndian.Uint32(data[84:]))
	if w <= 0 || h <= 0 || w > 4096 || h > 4096 || offset+size > len(data) {
		return nil, fmt.Errorf("bad header: %dx%d, mip at %d+%d of %d bytes", w, h, offset, size, len(data))
	}
	mip := data[offset : offset+size]
	img := image.NewNRGBA(image.Rect(0, 0, w, h))

	switch encoding {
	case 1:
		if err := decodePalette(img, mip, data[blpHeaderSize:blpHeaderSize+blpPaletteSize], alphaSize); err != nil {
			return nil, err
		}
	case 2:
		if err := decodeDXT(img, mip, alphaSize, alphaType); err != nil {
			return nil, err
		}
	case 3:
		if len(mip) < w*h*4 {
			return nil, fmt.Errorf("BGRA mip holds %d bytes, want %d", len(mip), w*h*4)
		}
		for i := range w * h {
			setPixel(img, i, mip[i*4+2], mip[i*4+1], mip[i*4], mip[i*4+3])
		}
	default:
		return nil, fmt.Errorf("unsupported encoding %d", encoding)
	}
	return img, nil
}

func setPixel(img *image.NRGBA, i int, r, g, b, a uint8) {
	copy(img.Pix[i*4:], []uint8{r, g, b, a})
}

// The i-th 4-bit alpha of a run packed low nibble first, scaled to 0-255.
func alpha4(b []byte, i int) uint8 {
	return (b[i/2] >> (4 * (i % 2)) & 0xF) * 17
}

// One palette index per pixel, then the alpha bits packed low bit first.
func decodePalette(img *image.NRGBA, mip, palette []byte, alphaSize uint8) error {
	n := img.Rect.Dx() * img.Rect.Dy()
	if want := n + (n*int(alphaSize)+7)/8; len(mip) < want {
		return fmt.Errorf("palette mip holds %d bytes, want %d", len(mip), want)
	}
	alpha := mip[n:]
	for i := range n {
		p := palette[int(mip[i])*4:]
		a := uint8(255)
		switch alphaSize {
		case 1:
			a = (alpha[i/8] >> (i % 8) & 1) * 255
		case 4:
			a = alpha4(alpha, i)
		case 8:
			a = alpha[i]
		}
		setPixel(img, i, p[2], p[1], p[0], a)
	}
	return nil
}

// DXT1 for 0 or 1 alpha bits; with 8, alpha type 7 is DXT5 and any other DXT3.
func decodeDXT(img *image.NRGBA, mip []byte, alphaSize, alphaType uint8) error {
	w, h := img.Rect.Dx(), img.Rect.Dy()
	blockSize := 16
	if alphaSize <= 1 {
		blockSize = 8
	}
	bw, bh := (w+3)/4, (h+3)/4
	if len(mip) < bw*bh*blockSize {
		return fmt.Errorf("DXT mip holds %d bytes, want %d", len(mip), bw*bh*blockSize)
	}
	var px [16][4]uint8
	for by := range bh {
		for bx := range bw {
			block := mip[(by*bw+bx)*blockSize:]
			if blockSize == 8 {
				decodeColorBlock(block, &px, true)
			} else {
				decodeColorBlock(block[8:], &px, false)
				if alphaType == 7 {
					decodeDXT5Alpha(block, &px)
				} else {
					for i := range 16 {
						px[i][3] = alpha4(block, i)
					}
				}
			}
			for i := range 16 {
				x, y := bx*4+i%4, by*4+i/4
				if x < w && y < h {
					setPixel(img, y*w+x, px[i][0], px[i][1], px[i][2], px[i][3])
				}
			}
		}
	}
	return nil
}

// Two RGB565 endpoints and 2-bit indices. DXT1 with the first endpoint not above the second has three
// colours and transparent black.
func decodeColorBlock(block []byte, px *[16][4]uint8, dxt1 bool) {
	c0, c1 := binary.LittleEndian.Uint16(block), binary.LittleEndian.Uint16(block[2:])
	four := !dxt1 || c0 > c1
	var colors [4][4]uint8
	colors[0], colors[1] = rgb565(c0), rgb565(c1)
	for ch := range 3 {
		a, b := int(colors[0][ch]), int(colors[1][ch])
		if four {
			colors[2][ch] = uint8((2*a + b) / 3)
			colors[3][ch] = uint8((a + 2*b) / 3)
		} else {
			colors[2][ch] = uint8((a + b) / 2)
		}
	}
	colors[2][3] = 255
	if four {
		colors[3][3] = 255
	}
	indices := binary.LittleEndian.Uint32(block[4:])
	for i := range 16 {
		px[i] = colors[indices>>(2*i)&3]
	}
}

func rgb565(c uint16) [4]uint8 {
	r, g, b := uint8(c>>11&0x1F), uint8(c>>5&0x3F), uint8(c&0x1F)
	return [4]uint8{r<<3 | r>>2, g<<2 | g>>4, b<<3 | b>>2, 255}
}

// Two alpha endpoints and 3-bit indices; the first endpoint not above the second leaves room for 0 and 255.
func decodeDXT5Alpha(block []byte, px *[16][4]uint8) {
	a0, a1 := int(block[0]), int(block[1])
	alphas := [8]int{a0, a1}
	steps := 7
	if a0 <= a1 {
		steps = 5
		alphas[6], alphas[7] = 0, 255
	}
	for i := 1; i < steps; i++ {
		alphas[i+1] = ((steps-i)*a0 + i*a1) / steps
	}
	bits := binary.LittleEndian.Uint64(block) >> 16 // the 48 index bits after the two endpoints
	for i := range 16 {
		px[i][3] = uint8(alphas[bits>>(3*i)&7])
	}
}
