package mjpeg

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"time"
)

var (
	// ErrTooLarge reports if more frames cannot be added,
	// else the video file would get corrupted.
	ErrTooLarge = errors.New("video file too large")

	// errImproperUse signals improper state (due to a previous error).
	errImproperState = errors.New("improper State")
)

// AviWriter is an *.avi video writer.
// The video codec is MJPEG.
type AviWriter interface {
	// AddFrame adds a frame from a JPEG encoded data slice.
	// AddFrame(jpegData []byte) error

	// Close finalizes and closes the avi file.
	// Close() error
}

// aviWriter is the AviWriter implementation.
type aviWriter struct {
	// width is the width of the video
	width int32
	// height is the height of the video
	height int32
	// fps is the frames/second (the "speed") of the video
	fps int32

	// idxFile is the name of the index file
	optionFile string
	// idxf is the index file descriptor
	optionf *os.File

	// dataFile is the name of the file to write the result to
	dataFile string
	// dataf is the avi file descriptor
	dataf *os.File

	// aviFile is the name of the file to write the result to
	aviFile string
	// avif is the avi file descriptor
	avif *os.File

	// idxFile is the name of the index file
	idxFile string
	// idxf is the index file descriptor
	idxf *os.File

	// writeErr holds the last encountered write error (to avif)
	err error

	// lengthFields contains the file positions of the length fields
	// that are filled later; used as a stack (LIFO)
	lengthFields []int64

	// Position of the frames count fields
	framesCountFieldPos, framesCountFieldPos2 int64
	// Position of the MOVI chunk
	moviPos int64

	// frames is the number of frames written to the AVI file
	frames int

	// General buffers used to write int values.
	buf4, buf2 []byte
}

type Options struct {
	LengthFields         []int64 `json:"lengthFields"`
	FramesCountFieldPos  int64   `json:"framesCountFieldPos"`
	FramesCountFieldPos2 int64   `json:"framesCountFieldPos2"`
	MoviPos              int64   `json:"moviPos"`
	Frames               int     `json:"frames"`
	Buf4                 []byte  `json:"buf4"`
	Buf2                 []byte  `json:"buf2"`
}

// Comment
func In(item string, items []string) bool {
	for i := range items {
		if items[i] == item {
			return true
		}
	}
	return false
}

// New returns a new AviWriter.
// The Close() method of the AviWriter must be called to finalize the video file.
func New(storage string, jpegs [][]byte, width, height, fps int32, aviFile string) (awr AviWriter, err error) {
	aw := &aviWriter{
		dataFile:     aviFile,
		width:        width,
		height:       height,
		fps:          fps,
		optionFile:   aviFile + ".__options__",
		aviFile:      aviFile + ".__data__",
		idxFile:      aviFile + ".__idx__",
		lengthFields: make([]int64, 0, 5),
		buf4:         make([]byte, 4),
		buf2:         make([]byte, 2),
	}

	defer func() {
		if err == nil {
			return
		}
		logErr := func(e error) {
			if e != nil {
				log.Printf("Error: %v\n", e)
			}
		}
		if aw.dataf != nil {
			logErr(aw.dataf.Close())
		}
		if aw.idxf != nil {
			logErr(aw.idxf.Close())
		}
	}()

	// Options Cache
	aw.optionf, err = os.Create(storage + "/" + aw.optionFile)
	if err != nil {
		return nil, err
	}

	// Full Video File
	aw.dataf, err = os.Create(storage + "/" + aw.dataFile)
	if err != nil {
		return nil, err
	}

	// Video Data Cache
	aw.avif, err = os.Create(storage + "/" + aw.aviFile)
	if err != nil {
		return nil, err
	}

	// Index Cache
	aw.idxf, err = os.Create(storage + "/" + aw.idxFile)
	if err != nil {
		return nil, err
	}

	dir, err := os.ReadDir(storage)

	// Delete older file
	if err == nil {
		items := []string{aw.optionFile, aw.dataFile, aw.aviFile, aw.idxFile}

		for _, entry := range dir {
			if entry.IsDir() || path.Ext(entry.Name()) == ".avi" || In(entry.Name(), items) {
				continue
			}

			os.Remove(storage + "/" + entry.Name())
		}
	}

	/*************************************** HEADER ***************************************/

	// wint16 := aw.writeInt16

	// Write AVI header
	WriteStr(aw.dataf, "RIFF")                                         // RIFF type
	aw.lengthFields, err = WriteLengthField(aw.dataf, aw.lengthFields) // File length (remaining bytes after this field) (nesting level 0)

	if err != nil {
		return nil, err
	}

	WriteStr(aw.dataf, "AVI ") // AVI signature
	WriteStr(aw.dataf, "LIST") // LIST chunk: data encoding

	aw.lengthFields, err = WriteLengthField(aw.dataf, aw.lengthFields) // Chunk length (nesting level 1)

	if err != nil {
		return nil, err
	}

	WriteStr(aw.dataf, "hdrl")        // LIST chunk type
	WriteStr(aw.dataf, "avih")        // avih sub-chunk
	WriteInt32(aw.dataf, 0x38)        // Sub-chunk length excluding the first 8 bytes of avih signature and size
	WriteInt32(aw.dataf, 1000000/fps) // Frame delay time in microsec
	WriteInt32(aw.dataf, 0)           // dwMaxBytesPerSec (maximum data rate of the file in bytes per second)
	WriteInt32(aw.dataf, 0)           // Reserved
	WriteInt32(aw.dataf, 0x10)        // dwFlags, 0x10 bit: AVIF_HASINDEX (the AVI file has an index chunk at the end of the file - for good performance); Windows Media Player can't even play it if index is missing!
	aw.framesCountFieldPos, err = CurrentPos(aw.dataf)

	if err != nil {
		return nil, err
	}

	WriteInt32(aw.dataf, 0)      // Number of frames
	WriteInt32(aw.dataf, 0)      // Initial frame for non-interleaved files; non interleaved files should set this to 0
	WriteInt32(aw.dataf, 1)      // Number of streams in the video; here 1 video, no audio
	WriteInt32(aw.dataf, 0)      // dwSuggestedBufferSize
	WriteInt32(aw.dataf, width)  // Image width in pixels
	WriteInt32(aw.dataf, height) // Image height in pixels
	WriteInt32(aw.dataf, 0)      // Reserved
	WriteInt32(aw.dataf, 0)
	WriteInt32(aw.dataf, 0)
	WriteInt32(aw.dataf, 0)

	// Write stream information
	WriteStr(aw.dataf, "LIST") // LIST chunk: stream headers

	aw.lengthFields, err = WriteLengthField(aw.dataf, aw.lengthFields) // Chunk size (nesting level 2)

	if err != nil {
		return nil, err
	}

	WriteStr(aw.dataf, "strl") // LIST chunk type: stream list
	WriteStr(aw.dataf, "strh") // Stream header
	WriteInt32(aw.dataf, 56)   // Length of the strh sub-chunk
	WriteStr(aw.dataf, "vids") // fccType - type of data stream - here 'vids' for video stream
	WriteStr(aw.dataf, "MJPG") // MJPG for Motion JPEG
	WriteInt32(aw.dataf, 0)    // dwFlags
	WriteInt32(aw.dataf, 0)    // wPriority, wLanguage
	WriteInt32(aw.dataf, 0)    // dwInitialFrames
	WriteInt32(aw.dataf, 1)    // dwScale
	WriteInt32(aw.dataf, fps)  // dwRate, Frame rate for video streams (the actual FPS is calculated by dividing this by dwScale)
	WriteInt32(aw.dataf, 0)    // usually zero

	aw.framesCountFieldPos2, err = CurrentPos(aw.dataf)

	if err != nil {
		return nil, err
	}

	WriteInt32(aw.dataf, 0)  // dwLength, playing time of AVI file as defined by scale and rate (set equal to the number of frames)
	WriteInt32(aw.dataf, 0)  // dwSuggestedBufferSize for reading the stream (typically, this contains a value corresponding to the largest chunk in a stream)
	WriteInt32(aw.dataf, -1) // wint32, encoding quality given by an integer between (0 and 10,000.  If set to -1, drivers use the default quality value)
	WriteInt32(aw.dataf, 0)  // dwSampleSize, 0 means that each frame is in its own chunk
	WriteInt32(aw.dataf, 0)  // left of rcFrame if stream has a different size than dwWidth*dwHeight(unused)
	WriteInt32(aw.dataf, 0)  //   ..top
	WriteInt32(aw.dataf, 0)  //   ..right
	WriteInt32(aw.dataf, 0)  //   ..bottom
	// end of 'strh' chunk, stream format follows
	WriteStr(aw.dataf, "strf") // stream format chunk

	aw.lengthFields, err = WriteLengthField(aw.dataf, aw.lengthFields) // Chunk size (nesting level 3)

	if err != nil {
		return nil, err
	}

	WriteInt32(aw.dataf, 40)     // biSize, write header size of BITMAPINFO header structure; applications should use this size to determine which BITMAPINFO header structure is being used, this size includes this biSize field
	WriteInt32(aw.dataf, width)  // biWidth, width in pixels
	WriteInt32(aw.dataf, height) // biWidth, height in pixels (may be negative for uncompressed video to indicate vertical flip)
	writeInt16(aw.dataf, 1)      // biPlanes, number of color planes in which the data is stored
	writeInt16(aw.dataf, 24)

	// biBitCount, number of bits per pixel #
	WriteStr(aw.dataf, "MJPG")                                            // biCompression, type of compression used (uncompressed: NO_COMPRESSION=0)
	WriteInt32(aw.dataf, width*height*3)                                  // biSizeImage (buffer size for decompressed mage) may be 0 for uncompressed data
	WriteInt32(aw.dataf, 0)                                               // biXPelsPerMeter, horizontal resolution in pixels per meter
	WriteInt32(aw.dataf, 0)                                               // biYPelsPerMeter, vertical resolution in pixels per meter
	WriteInt32(aw.dataf, 0)                                               // biClrUsed (color table size; for 8-bit only)
	WriteInt32(aw.dataf, 0)                                               // biClrImportant, specifies that the first x colors of the color table (0: all the colors are important, or, rather, their relative importance has not been computed)
	aw.lengthFields, err = FinalizeLengthField(aw.dataf, aw.lengthFields) // 'strf' chunk finished (nesting level 3)

	if err != nil {
		return nil, err
	}

	WriteStr(aw.dataf, "strn") // Use 'strn' to provide a zero terminated text string describing the stream

	name := fmt.Sprintf("classify at %s", time.Now().Format("2025-01-01 00:00:00 GTM"))

	// Name must be 0-terminated and stream name length (the length of the chunk) must be even
	if len(name)&0x01 == 0 {
		name = name + " \000" // padding space plus terminating 0
	} else {
		name = name + "\000" // terminating 0
	}
	WriteInt32(aw.dataf, int32(len(name))) // Length of the strn sub-CHUNK (must be even)
	WriteStr(aw.dataf, name)
	aw.lengthFields, err = FinalizeLengthField(aw.dataf, aw.lengthFields) // LIST 'strl' finished (nesting level 2)
	aw.lengthFields, err = FinalizeLengthField(aw.dataf, aw.lengthFields) // LIST 'hdrl' finished (nesting level 1)

	WriteStr(aw.dataf, "LIST") // The second LIST chunk, which contains the actual data

	aw.lengthFields, err = WriteLengthField(aw.dataf, aw.lengthFields) // Chunk length (nesting level 1)

	if err != nil {
		return nil, err
	}

	aw.moviPos, err = CurrentPos(aw.dataf)

	if err != nil {
		return nil, err
	}

	WriteStr(aw.dataf, "movi") // LIST chunk type: 'movi'

	if aw.err != nil {
		return nil, aw.err
	}

	/*************************************** DATA ***************************************/

	for _, jpegData := range jpegs {
		framePos, err := CurrentPos(aw.dataf)

		if err != nil {
			continue // TODO bad things happing...
		}

		// Pointers in AVI are 32 bit. Do not write beyond that else the whole AVI file will be corrupted (not playable).
		// Index entry size: 16 bytes (for each frame)
		if framePos+int64(len(jpegData))+int64(aw.frames*16) > 4200000000 { // 2^32 = 4 294 967 296
			return nil, ErrTooLarge
		}

		aw.frames++

		WriteInt32(aw.dataf, 0x63643030) // "00dc" compressed frame

		// Chunk length (nesting level 2)
		if aw.lengthFields, err = WriteLengthField(aw.dataf, aw.lengthFields); err != nil {
			continue // TODO bad things happing...
		}

		err = Write(aw.dataf, jpegData)

		if err != nil {
			continue // TODO bad things happing...
		}

		aw.lengthFields, err = FinalizeLengthField(aw.dataf, aw.lengthFields) // "00dc" chunk finished (nesting level 2)

		if err != nil {
			continue // TODO bad things happing...
		}

		// Write index data
		WriteInt32(aw.idxf, 0x63643030)                 // "00dc" compressed frame
		WriteInt32(aw.idxf, 0x10)                       // flags: select AVIIF_KEYFRAME (The flag indicates key frames in the video sequence. Key frames do not need previous video information to be decompressed.)
		WriteInt32(aw.idxf, int32(framePos-aw.moviPos)) // offset to the chunk, offset can be relative to file start or 'movi'
		WriteInt32(aw.idxf, int32(len(jpegData)))       // length of the chunk
	}

	/*************************************** INDEX ***************************************/

	defer func() {
		aw.dataf.Close()
		aw.idxf.Close()
		// os.Remove(aw.idxFile)
	}()

	// aw.finalizeLengthField()
	aw.lengthFields, err = FinalizeLengthField(aw.dataf, aw.lengthFields) // LIST 'movi' finished (nesting level 1)

	// Write index
	WriteStr(aw.dataf, "idx1") // idx1 chunk
	var idxLength int64

	if aw.err == nil {
		idxLength, aw.err = Seek(aw.idxf, 0, 1) // Seek relative to current pos
	}

	WriteInt32(aw.dataf, int32(idxLength)) // Chunk length (we know its size, no need to use writeLengthField() and finalizeLengthField() pair)

	// Copy temporary index data
	if aw.err == nil { // TOD RM
		_, aw.err = Seek(aw.idxf, 0, 0)
	}

	if aw.err == nil { // TODO RM
		_, aw.err = io.Copy(aw.dataf, aw.idxf)
	}

	pos, err := CurrentPos(aw.dataf)

	if err != nil {
		return nil, err
	}

	Seek(aw.dataf, aw.framesCountFieldPos, 0)
	WriteInt32(aw.dataf, int32(aw.frames))
	Seek(aw.dataf, aw.framesCountFieldPos2, 0)
	WriteInt32(aw.dataf, int32(aw.frames))
	Seek(aw.dataf, pos, 0)

	aw.lengthFields, err = FinalizeLengthField(aw.dataf, aw.lengthFields) // 'RIFF' File finished (nesting level 0)

	if err != nil {
		return nil, aw.err
	}

	return aw, nil
}
