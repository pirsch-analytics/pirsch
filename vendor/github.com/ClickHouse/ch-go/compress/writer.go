package compress

import (
	"encoding/binary"
	"math"

	"github.com/go-faster/city"
	"github.com/go-faster/errors"
	"github.com/klauspost/compress/zstd"
	"github.com/pierrec/lz4/v4"
)

const (
	CompressionLevelLZ4HCDefault Level = 9
	CompressionLevelLZ4HCMax     Level = 12
)

// Writer encodes compressed blocks.
type Writer struct {
	Data []byte

	lz4   *lz4.Compressor
	lz4hc *lz4.CompressorHC
	zstd  *zstd.Encoder
}

// Compress buf into Data.
func (w *Writer) Compress(m Method, buf []byte) error {
	maxSize := lz4.CompressBlockBound(len(buf))
	w.Data = append(w.Data[:0], make([]byte, maxSize+headerSize)...)
	_ = w.Data[:headerSize]
	w.Data[hMethod] = byte(methodTable[m])

	var n int

	switch m {
	case LZ4:
		compressedSize, err := w.lz4.CompressBlock(buf, w.Data[headerSize:])
		if err != nil {
			return errors.Wrap(err, "block")
		}
		n = compressedSize
	case LZ4HC:
		compressedSize, err := w.lz4hc.CompressBlock(buf, w.Data[headerSize:])
		if err != nil {
			return errors.Wrap(err, "block")
		}
		n = compressedSize
	case ZSTD:
		w.Data = w.zstd.EncodeAll(buf, w.Data[:headerSize])
		n = len(w.Data) - headerSize
	case None:
		n = copy(w.Data[headerSize:], buf)
	}

	w.Data = w.Data[:n+headerSize]

	binary.LittleEndian.PutUint32(w.Data[hRawSize:], uint32(n+compressHeaderSize))
	binary.LittleEndian.PutUint32(w.Data[hDataSize:], uint32(len(buf)))
	h := city.CH128(w.Data[hMethod:])
	binary.LittleEndian.PutUint64(w.Data[0:8], h.Low)
	binary.LittleEndian.PutUint64(w.Data[8:16], h.High)

	return nil
}

func NewWriterWithLevel(l Level) *Writer {
	w, err := zstd.NewWriter(nil,
		zstd.WithEncoderLevel(zstd.SpeedDefault),
		zstd.WithEncoderConcurrency(1),
		zstd.WithLowerEncoderMem(true),
	)
	if err != nil {
		panic(err)
	}

	// handle level for LZ4HC
	levelLZ4HC := l
	if levelLZ4HC == 0 {
		levelLZ4HC = CompressionLevelLZ4HCDefault
	} else {
		levelLZ4HC = Level(math.Min(float64(levelLZ4HC), float64(CompressionLevelLZ4HCMax)))
	}

	return &Writer{
		lz4:   &lz4.Compressor{},
		lz4hc: &lz4.CompressorHC{Level: lz4.CompressionLevel(1 << (8 + levelLZ4HC))},
		zstd:  w,
	}
}

func NewWriter() *Writer {
	return NewWriterWithLevel(0)
}
