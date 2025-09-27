package model

import (
	"context"
	"io"
)

type streamReader interface {
	Read(ctx context.Context, src io.Reader, emit func(StreamChunk)) (streamState, error)
}

type streamState interface {
	Finalize(out chan<- StreamChunk)
	ResponseID() string
}
