package ffmpeg

import "time"

type Video interface {
	Record(time time.Time, frame []byte) error
	Build(start time.Time, end time.Time) ([]byte, error)
}
