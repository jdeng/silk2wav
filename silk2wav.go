package silk2wav

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"unsafe"
)

// #cgo CFLAGS: -I ./interface -I ./src
// #include "silk.inl"
import "C"

func decode(buf []byte) []byte {
	const MAXSIZE = 20 * 48 * 5 * 2 * 2
	var b bytes.Buffer
	out := make([]byte, MAXSIZE)

	d := C.new_decoder()
	for {
		if len(buf) < 2 {
			break
		}
		plen := int(uint16(buf[0]) + (uint16(buf[1]) << 8))

		if len(buf) < 2+plen {
			log.Printf("%d bytes expected but only %d available\n", int(plen), len(buf))
			break
		}
		payload := buf[2 : 2+plen]

		buf = buf[2+plen:]

		olen := C.decode_frame(d, C.int(SAMPLE_RATE), unsafe.Pointer(&payload[0]), C.int(plen), unsafe.Pointer(&out[0]), C.int(len(out)))
		if olen < 0 {
			log.Printf("Failed to decode %d bytes\n", plen)
			break
		}
		b.Write(out[:olen])
	}

	if d != nil {
		C.free_decoder(d)
	}

	return b.Bytes()
}

type wavHeader struct {
	ChunkID       [4]byte
	ChunkSize     uint32
	Format        [4]byte
	Subchunk1ID   [4]byte
	Subchunk1Size uint32
	AudioFormat   uint16
	NumChannels   uint16
	SampleRate    uint32
	ByteRate      uint32
	BlockAlign    uint16
	BitsPerSample uint16
	Subchunk2ID   [4]byte
	Subchunk2Size uint32
}

const SAMPLE_RATE = 24000

func Convert(buf []byte) ([]byte, error) {
	const TAG = "#!SILK_V3"
	if len(buf) < 1+len(TAG) || buf[0] != 0x02 || bytes.Compare(buf[1:1+len(TAG)], []byte(TAG)) != 0 {
		return nil, fmt.Errorf("Invalid header")
	}

	out := decode(buf[1+len(TAG):])
	if out == nil {
		return nil, fmt.Errorf("Failed to decode")
	}

	datalen := uint32(len(out))
	hdr := wavHeader{
		ChunkID:       [4]byte{'R', 'I', 'F', 'F'},
		ChunkSize:     36 + datalen,
		Format:        [4]byte{'W', 'A', 'V', 'E'},
		Subchunk1ID:   [4]byte{'f', 'm', 't', ' '},
		Subchunk1Size: 16,
		AudioFormat:   1, //1 = PCM not compressed
		NumChannels:   1,
		SampleRate:    SAMPLE_RATE,
		ByteRate:      2 * SAMPLE_RATE,
		BlockAlign:    2,
		BitsPerSample: 16,
		Subchunk2ID:   [4]byte{'d', 'a', 't', 'a'},
		Subchunk2Size: datalen,
	}

	var b bytes.Buffer
	binary.Write(&b, binary.LittleEndian, hdr)
	b.Write(out)
	return b.Bytes(), nil
}
