// Package storj provides an interface to Storj object storage.
package storj

import (
	"context"
	"io"
	"path"
	"time"

	"github.com/pkg/errors"
	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/hash"
	"storj.io/storj/lib/uplink"
	"storj.io/storj/pkg/storj"
)

// Object describes a storj object
type Object struct {
	fs *Fs

	relative string
	isPrefix bool

	created  time.Time
	modified time.Time

	size     int64
	checksum []byte
}

// Check the interfaces are satisfied.
var _ fs.Object = &Object{}

// NewObjectFromUplink creates a new object from a Storj uplink object.
func NewObjectFromUplink(f *Fs, object *uplink.Object) *Object {
	// Attempt to use the modified time from the metadata. Otherwise
	// fallback to the server time.
	modified := object.Meta.Modified

	if modifiedStr, ok := object.Meta.Metadata["Modified"]; ok {
		var err error

		modified, err = time.Parse(time.RFC3339Nano, modifiedStr)
		if err != nil {
			modified = object.Meta.Modified
		}
	}

	return &Object{
		fs: f,

		relative: object.Meta.Path,
		isPrefix: object.Meta.IsPrefix,

		created:  object.Meta.Created,
		modified: modified,

		size:     object.Meta.Size,
		checksum: object.Meta.Checksum,
	}
}

// NewObjectFromStorj creates a new object from a Storj storj object.
func NewObjectFromStorj(f *Fs, object *storj.Object) *Object {
	// Attempt to use the modified time from the metadata. Otherwise
	// fallback to the server time.
	modified := object.Modified

	if modifiedStr, ok := object.Metadata["Modified"]; ok {
		var err error

		modified, err = time.Parse(time.RFC3339Nano, modifiedStr)
		if err != nil {
			modified = object.Modified
		}
	}

	return &Object{
		fs: f,

		relative: object.Path,
		isPrefix: object.IsPrefix,

		created:  object.Created,
		modified: modified,

		size:     object.Size,
		checksum: object.Checksum,
	}
}

// absolute computes the absolute bucket name and path for the object
// considering the underlying filesystem root and the object's relative path.
func (o *Object) absolute() (bucketName, bucketPath string) {
	return o.fs.absolute(o.relative)
}

// setRelative updates the object's relative path and returns the object.
func (o *Object) setRelative(relative string) *Object {
	o.relative = relative

	return o
}

// String returns a description of the Object
func (o *Object) String() string {
	if o == nil {
		return "<nil>"
	}

	return o.relative
}

// Remote returns the remote path
func (o *Object) Remote() string {
	// The relative path can end up empty if the root of the filesystem
	// *is* an object. In this case rclone needs to see the filename.
	if o.relative == "" {
		_, filename := path.Split(path.Join(o.absolute()))

		return filename
	}

	return o.relative
}

// ModTime returns the modification date of the file
// It should return a best guess if one isn't available
func (o *Object) ModTime(ctx context.Context) time.Time {
	return o.modified
}

// Size returns the size of the file
func (o *Object) Size() int64 {
	return o.size
}

// Fs returns read only access to the Fs that this object is part of
func (o *Object) Fs() fs.Info {
	return o.fs
}

// Hash returns the selected checksum of the file
// If no checksum is available it returns ""
func (o *Object) Hash(ctx context.Context, ty hash.Type) (string, error) {
	fs.Debugf(o, "%s", ty)

	return "", hash.ErrUnsupported
}

// Storable says whether this object can be stored
func (o *Object) Storable() bool {
	return true
}

// SetModTime sets the metadata on the object to set the modification date
func (o *Object) SetModTime(ctx context.Context, t time.Time) error {
	fs.Debugf(o, "touch -d %q sj://%s", t, path.Join(o.absolute()))

	return fs.ErrorCantSetModTime
}

// Open opens the file for read.  Call Close() on the returned io.ReadCloser
func (o *Object) Open(ctx context.Context, options ...fs.OpenOption) (io.ReadCloser, error) {
	fs.Infof(o, "cat sj://%s # %+v", path.Join(o.absolute()), options)

	bucketName, bucketPath := o.absolute()

	bucket, err := o.fs.project.OpenBucket(ctx, bucketName, o.fs.scope.EncryptionAccess)
	if err != nil {
		return nil, err
	}
	defer bucket.Close()

	object, err := bucket.OpenObject(ctx, bucketPath)
	if err != nil {
		return nil, err
	}
	defer object.Close()

	var (
		offset int64 = 0
		length int64 = -1
	)

	for _, option := range options {
		switch opt := option.(type) {
		case *fs.RangeOption:
			s := opt.Start >= 0
			e := opt.End >= 0

			switch {
			case s && e:
				offset = opt.Start
				length = (opt.End + 1) - opt.Start
			case s && !e:
				offset = opt.Start
			case !s && e:
				offset = object.Meta.Size - opt.End
				length = opt.End
			}
		case *fs.SeekOption:
			offset = opt.Offset
		default:
			if option.Mandatory() {
				fs.Errorf(o, "Unsupported mandatory option: %v", option)

				return nil, errors.New("unsupported mandatory option")
			}
		}
	}

	fs.Debugf(o, "range %d + %d", offset, length)

	return object.DownloadRange(ctx, offset, length)
}

// Update in to the object with the modTime given of the given size
//
// When called from outside a Fs by rclone, src.Size() will always be >= 0.
// But for unknown-sized objects (indicated by src.Size() == -1), Upload should either
// return an error or update the object properly (rather than e.g. calling panic).
func (o *Object) Update(ctx context.Context, in io.Reader, src fs.ObjectInfo, options ...fs.OpenOption) error {
	fs.Debugf(o, "cp input ./%s %+v", src.Remote(), options)

	oNew, err := o.fs.Put(ctx, in, src, options...)

	if err == nil {
		*o = *(oNew.(*Object))
	}

	return err
}

// Remove this object.
func (o *Object) Remove(ctx context.Context) error {
	fs.Infof(o, "rm sj://%s", path.Join(o.absolute()))

	bucketName, bucketPath := o.absolute()

	u, p, err := o.fs.connect(ctx)
	if err != nil {
		return err
	}
	defer u.Close()
	defer p.Close()

	bucket, err := p.OpenBucket(ctx, bucketName, o.fs.scope.EncryptionAccess)
	if err != nil {
		return err
	}
	defer bucket.Close()

	err = bucket.DeleteObject(ctx, bucketPath)

	return err
}
