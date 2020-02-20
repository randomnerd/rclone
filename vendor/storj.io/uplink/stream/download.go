// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package stream

import (
	"context"
	"io"

	"storj.io/common/storj"
	"storj.io/uplink/metainfo/kvmetainfo"
	"storj.io/uplink/storage/streams"
)

// Download implements Reader, Seeker and Closer for reading from stream.
type Download struct {
	ctx     context.Context
	stream  kvmetainfo.ReadOnlyStream
	streams streams.Store
	reader  io.ReadCloser
	offset  int64
	limit   int64
	closed  bool
}

// NewDownload creates new stream download.
func NewDownload(ctx context.Context, stream kvmetainfo.ReadOnlyStream, streams streams.Store) *Download {
	return &Download{
		ctx:     ctx,
		stream:  stream,
		streams: streams,
		limit:   -1,
	}
}

// NewDownloadRange creates new stream range download with range from offset to offset+limit.
func NewDownloadRange(ctx context.Context, stream kvmetainfo.ReadOnlyStream, streams streams.Store, offset, limit int64) *Download {
	return &Download{
		ctx:     ctx,
		stream:  stream,
		streams: streams,
		offset:  offset,
		limit:   limit,
	}
}

// Read reads up to len(data) bytes into data.
//
// If this is the first call it will read from the beginning of the stream.
// Use Seek to change the current offset for the next Read call.
//
// See io.Reader for more details.
func (download *Download) Read(data []byte) (n int, err error) {
	if download.closed {
		return 0, Error.New("already closed")
	}

	if download.reader == nil {
		err = download.resetReader(download.offset)
		if err != nil {
			return 0, err
		}
	}

	if download.limit == 0 {
		return 0, io.EOF
	}
	if download.limit > 0 && download.limit < int64(len(data)) {
		data = data[:download.limit]
	}
	n, err = download.reader.Read(data)
	if download.limit >= 0 {
		download.limit -= int64(n)
	}
	download.offset += int64(n)

	return n, err
}

// Close closes the stream and releases the underlying resources.
func (download *Download) Close() error {
	if download.closed {
		return Error.New("already closed")
	}

	download.closed = true

	if download.reader == nil {
		return nil
	}

	return download.reader.Close()
}

func (download *Download) resetReader(offset int64) error {
	if download.reader != nil {
		err := download.reader.Close()
		if err != nil {
			return err
		}
	}

	obj := download.stream.Info()

	rr, err := download.streams.Get(download.ctx, storj.JoinPaths(obj.Bucket.Name, obj.Path), obj, obj.Bucket.PathCipher)
	if err != nil {
		return err
	}

	download.reader, err = rr.Range(download.ctx, offset, obj.Size-offset)
	if err != nil {
		return err
	}

	download.offset = offset

	return nil
}
