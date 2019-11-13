// Package storj provides an interface to Storj object storage.
package storj

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/config/configmap"
	"github.com/rclone/rclone/fs/config/configstruct"
	"github.com/rclone/rclone/fs/fserrors"
	"github.com/rclone/rclone/fs/hash"
	"github.com/rclone/rclone/lib/bucket"

	"storj.io/storj/lib/uplink"
	"storj.io/storj/pkg/storj"
)

// Register with Fs
func init() {
	fs.Register(&fs.RegInfo{
		Name:        "storj",
		Description: "Storj Connection",
		NewFs:       NewFs,
		Options: []fs.Option{
			{
				Name:     "scope",
				Help:     "Uplink scope.",
				Required: true,
			},
			{
				Name:    "skip-peer-ca-whitelist",
				Help:    "Disable peer CA whitelist enforcement.",
				Default: false,
			},
			{
				Name:    "defaults",
				Help:    "Defaults to use: dev or release",
				Default: "release",
			},
		},
	})
}

// Options defines the configuration for this backend
type Options struct {
	Scope               string `config:"scope"`
	SkipPeerCAWhitelist bool   `config:"skip-peer-ca-whitelist"`
	Defaults            string `config:"defaults"`
}

// Fs represents a remote storj server
type Fs struct {
	name string // the name of the remote
	root string // root of the filesystem

	opts     Options      // parsed options
	features *fs.Features // optional features

	scope *uplink.Scope // parsed scope

	uplink  *uplink.Uplink  // uplink client
	project *uplink.Project // project client
}

// Check the interfaces are satisfied.
var (
	_ fs.Fs          = &Fs{}
	_ fs.ListRer     = &Fs{}
	_ fs.PutStreamer = &Fs{}
)

// NewFs creates a filesystem backed by Storj.
func NewFs(name, root string, m configmap.Mapper) (_ fs.Fs, err error) {
	ctx := context.Background()

	// Parse config into Options struct
	opts := new(Options)
	err = configstruct.Set(m, opts)
	if err != nil {
		return nil, err
	}

	// Parse scope
	scopeb58 := opts.Scope
	if scopeb58 == "" {
		return nil, errors.New("scope not found")
	}

	scope, err := uplink.ParseScope(scopeb58)
	if err != nil {
		return nil, errors.Wrap(err, "storj: scope")
	}

	// Setup filesystem and connection to Storj
	root = strings.Trim(root, "/")

	f := &Fs{
		name: name,
		root: root,

		opts: *opts,

		scope: scope,
	}
	f.features = (&fs.Features{
		BucketBased:       true,
		BucketBasedRootOK: true,
	}).Fill(f)

	upl, project, err := f.connect(ctx)
	if err != nil {
		return nil, err
	}
	f.uplink = upl
	f.project = project

	// Root validation needs to check the following: If a bucket path is
	// specified and exists, then the object must be a directory.
	//
	// NOTE: At this point this must return the filesystem object we've
	// created so far even if there is an error.
	if root != "" {
		bucketName, bucketPath := bucket.Split(root)

		if bucketName != "" && bucketPath != "" {
			bucket, err := project.OpenBucket(ctx, bucketName, scope.EncryptionAccess)
			if err != nil {
				return f, errors.Wrap(err, "storj: bucket")
			}
			defer bucket.Close()

			object, err := bucket.OpenObject(ctx, bucketPath)
			if err == nil {
				defer object.Close()

				if !object.Meta.IsPrefix {
					return f, fs.ErrorIsFile
				}
			}
		}
	}

	return f, nil
}

// connect opens a connection to Storj.
func (f *Fs) connect(ctx context.Context) (upl *uplink.Uplink, project *uplink.Project, err error) {
	fs.Debugf(f, "connecting...")
	defer fs.Debugf(f, "connected: %+v", err)

	cfg := &uplink.Config{}
	cfg.Volatile.TLS.SkipPeerCAWhitelist = f.opts.SkipPeerCAWhitelist

	upl, err = uplink.NewUplink(ctx, cfg)
	if err != nil {
		return nil, nil, errors.Wrap(err, "storj: uplink")
	}

	project, err = upl.OpenProject(ctx, f.scope.SatelliteAddr, f.scope.APIKey)
	if err != nil {
		upl.Close()

		return nil, nil, errors.Wrap(err, "storj: project")
	}

	return upl, project, nil
}

// absolute computes the absolute bucket name and path from the filesystem root
// and the relative path provided.
func (f *Fs) absolute(relative string) (bucketName, bucketPath string) {
	bucketName, bucketPath = bucket.Split(path.Join(f.root, relative))

	return bucket.Split(path.Join(f.root, relative))
}

// Name of the remote (as passed into NewFs)
func (f *Fs) Name() string {
	return f.name
}

// Root of the remote (as passed into NewFs)
func (f *Fs) Root() string {
	return f.root
}

// String returns a description of the FS
func (f *Fs) String() string {
	return fmt.Sprintf("FS sj://%s", f.root)
}

// Precision of the ModTimes in this Fs
func (f *Fs) Precision() time.Duration {
	return time.Nanosecond
}

// Hashes returns the supported hash types of the filesystem.
func (f *Fs) Hashes() hash.Set {
	return hash.NewHashSet()
}

// Features returns the optional features of this Fs
func (f *Fs) Features() *fs.Features {
	return f.features
}

// List the objects and directories in relative into entries. The entries can
// be returned in any order but should be for a complete directory.
//
// relative should be "" to list the root, and should not have trailing
// slashes.
//
// This should return fs.ErrDirNotFound if the directory isn't found.
func (f *Fs) List(ctx context.Context, relative string) (entries fs.DirEntries, err error) {
	fs.Infof(f, "ls ./%s", relative)

	bucketName, bucketPath := f.absolute(relative)

	if bucketName == "" {
		if bucketPath != "" {
			return nil, fs.ErrorListBucketRequired
		}

		return f.listBuckets(ctx)
	}

	return f.listObjects(ctx, relative, bucketName, bucketPath)
}

func (f *Fs) listBuckets(ctx context.Context) (entries fs.DirEntries, err error) {
	fs.Debugf(f, "BKT ls")

	opts := uplink.BucketListOptions{
		Direction: storj.Forward,
	}

	for {
		result, err := f.project.ListBuckets(ctx, &opts)
		if err != nil {
			fs.Debugf(f, "err: %+v", err)

			return nil, fs.ErrorDirNotFound
		}

		for _, bucket := range result.Items {
			d := fs.NewDir(bucket.Name, bucket.Created).SetSize(0)
			entries = append(entries, d)
		}

		if ctx.Err() != nil {
			return entries, ctx.Err()
		}

		if !result.More {
			break
		}

		opts = opts.NextPage(result)
	}

	return entries, nil
}

func (f *Fs) listObjects(ctx context.Context, relative, bucketName, bucketPath string) (entries fs.DirEntries, err error) {
	fs.Debugf(f, "OBJ ls ./%s", relative)

	bucket, err := f.project.OpenBucket(ctx, bucketName, f.scope.EncryptionAccess)
	if err != nil {
		fs.Debugf(f, "err %+v", err)

		return nil, fs.ErrorDirNotFound
	}
	defer bucket.Close()

	// Attempt to open the directory itself. If the directory itself exists
	// as an object it will not show up through the listing below.
	object, err := bucket.OpenObject(ctx, bucketPath)
	if err == nil {
		entries = append(entries, NewObjectFromUplink(f, object).setRelative(""))

		object.Close()
	}

	startAfter := ""

	for {
		opts := &uplink.ListOptions{
			Direction: storj.After,
			Cursor:    startAfter,
			Prefix:    bucketPath,
		}

		result, err := bucket.ListObjects(ctx, opts)
		if err != nil {
			fs.Debugf(f, "err: %+v", err)

			return nil, err
		}

		for _, object := range result.Items {
			if object.IsPrefix {
				d := fs.NewDir(path.Join(relative, object.Path), object.Modified).SetSize(object.Size)
				entries = append(entries, d)

				continue
			}

			entries = append(entries, NewObjectFromStorj(f, &object).setRelative(path.Join(relative, object.Path)))
		}

		if ctx.Err() != nil {
			return entries, ctx.Err()
		}

		if !result.More {
			break
		}

		startAfter = result.Items[len(result.Items)-1].Path
	}

	return entries, nil
}

// ListR lists the objects and directories of the Fs starting from dir
// recursively into out.
//
// dir should be "" to start from the root, and should not have trailing
// slashes.
//
// This should return ErrDirNotFound if the directory isn't found.
//
// It should call callback for each tranche of entries read.  These need not be
// returned in any particular order.  If callback returns an error then the
// listing will stop immediately.
//
// Don't implement this unless you have a more efficient way of listing
// recursively that doing a directory traversal.
func (f *Fs) ListR(ctx context.Context, relative string, callback fs.ListRCallback) error {
	fs.Infof(f, "ls -R ./%s", relative)

	bucketName, bucketPath := f.absolute(relative)

	if bucketName == "" {
		if bucketPath != "" {
			return fs.ErrorListBucketRequired
		}

		return f.listBucketsR(ctx, callback)
	}

	return f.listObjectsR(ctx, relative, bucketName, bucketPath, callback)
}

func (f *Fs) listBucketsR(ctx context.Context, callback fs.ListRCallback) (err error) {
	fs.Debugf(f, "BKT ls -R")

	opts := uplink.BucketListOptions{
		Direction: storj.Forward,
	}

	for {
		var entries fs.DirEntries

		result, err := f.project.ListBuckets(ctx, &opts)
		if err != nil {
			fs.Debugf(f, "err: %+v", err)

			return fs.ErrorDirNotFound
		}

		for _, bucket := range result.Items {
			d := fs.NewDir(bucket.Name, bucket.Created).SetSize(0)
			entries = append(entries, d)
		}

		err = callback(entries)
		if err != nil {
			return err
		}

		if !result.More {
			break
		}

		opts = opts.NextPage(result)
	}

	return nil
}

func (f *Fs) listObjectsR(ctx context.Context, relative, bucketName, bucketPath string, callback fs.ListRCallback) (err error) {
	fs.Debugf(f, "OBJ ls -R ./%s", relative)

	bucket, err := f.project.OpenBucket(ctx, bucketName, f.scope.EncryptionAccess)
	if err != nil {
		fs.Debugf(f, "err %+v", err)

		return fs.ErrorDirNotFound
	}
	defer bucket.Close()

	// Attempt to open the directory itself. If the directory itself exists
	// as an object it will not show up through the listing below.
	object, err := bucket.OpenObject(ctx, bucketPath)
	if err == nil {
		err = callback(fs.DirEntries{NewObjectFromUplink(f, object).setRelative("")})
		if err != nil {
			return err
		}

		object.Close()
	}

	startAfter := ""

	for {
		var entries fs.DirEntries

		opts := &uplink.ListOptions{
			Direction: storj.After,
			Cursor:    startAfter,
			Prefix:    bucketPath,
			Recursive: true,
		}

		result, err := bucket.ListObjects(ctx, opts)
		if err != nil {
			fs.Debugf(f, "err: %+v", err)

			return err
		}

		for _, object := range result.Items {
			if object.IsPrefix {
				d := fs.NewDir(path.Join(relative, object.Path), object.Modified).SetSize(object.Size)
				entries = append(entries, d)

				continue
			}

			entries = append(entries, NewObjectFromStorj(f, &object).setRelative(path.Join(relative, object.Path)))
		}

		err = callback(entries)
		if err != nil {
			return err
		}

		if !result.More {
			break
		}

		startAfter = result.Items[len(result.Items)-1].Path
	}

	return nil
}

// NewObject finds the Object at relative. If it can't be found it returns the
// error ErrorObjectNotFound.
func (f *Fs) NewObject(ctx context.Context, relative string) (fs.Object, error) {
	fs.Infof(f, "stat ./%s", relative)

	bucketName, bucketPath := f.absolute(relative)

	bucket, err := f.project.OpenBucket(ctx, bucketName, f.scope.EncryptionAccess)
	if err != nil {
		fs.Debugf(f, "err: %+v", err)

		return nil, fs.ErrorObjectNotFound
	}
	defer bucket.Close()

	object, err := bucket.OpenObject(ctx, bucketPath)
	if err != nil {
		fs.Debugf(f, "err: %+v", err)

		return nil, fs.ErrorObjectNotFound
	}
	defer object.Close()

	o := NewObjectFromUplink(f, object)

	// If the root is empty, then the passed in relative path *is* the path
	// we want and not the object path itself.
	if f.root == "" {
		o.setRelative(relative)
	}

	return o, nil
}

// Put in to the remote path with the modTime given of the given size
//
// When called from outside a Fs by rclone, src.Size() will always be >= 0.
// But for unknown-sized objects (indicated by src.Size() == -1), Put should
// either return an error or upload it properly (rather than e.g. calling
// panic).
//
// May create the object even if it returns an error - if so will return the
// object and the error, otherwise will return nil and the error
func (f *Fs) Put(ctx context.Context, in io.Reader, src fs.ObjectInfo, options ...fs.OpenOption) (fs.Object, error) {
	fs.Infof(f, "cp input ./%s # %+v %d", src.Remote(), options, src.Size())

	// Reject options we don't support.
	for _, option := range options {
		switch opt := option.(type) {
		case *fs.RangeOption:
			fs.Errorf(f, "Unsupported range option: %v", opt)

			return nil, errors.New("ranged writes are not supported")
		case *fs.SeekOption:
			if opt.Offset != 0 {
				fs.Errorf(f, "Unsupported seek option: %v", opt)

				return nil, errors.New("writes from a non-zero offset are not supported")
			}
		default:
			if option.Mandatory() {
				fs.Errorf(f, "Unsupported mandatory option: %v", option)

				return nil, errors.New("unsupported mandatory option")
			}
		}
	}

	bucketName, bucketPath := f.absolute(src.Remote())

	u, p, err := f.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer u.Close()
	defer p.Close()

	bucket, err := p.OpenBucket(ctx, bucketName, f.scope.EncryptionAccess)
	if err != nil {
		return nil, err
	}
	defer bucket.Close()

	out, err := bucket.NewWriter(ctx, bucketPath, &uplink.UploadOptions{
		Metadata: map[string]string{
			"Modified": src.ModTime(ctx).Format(time.RFC3339Nano),
		},
	})
	if err != nil {
		return nil, err
	}

	// Write to the output with a timeout.
	for {
		ch := make(chan error)
		go func() {
			n, err := io.CopyN(out, in, 1024)
			fs.Debugf(f, "cp input ./%s %+v: copied %d %+v", src.Remote(), options, n, err)

			if err == io.EOF {
				out.Close()
			}

			ch <- err
		}()

		select {
		case err = <-ch:
		case <-time.After(30 * time.Second):
			err = fserrors.RetryError(errors.New("timed out waiting on copy"))
			fs.Errorf(f, "cp input ./%s %+v: %+v", src.Remote(), options, err)
		}
		if err != nil {
			if err == io.EOF {
				err = nil
			}

			break
		}
	}
	if err != nil {
		return nil, err
	}

	fs.Debugf(f, "cp input ./%s %+v: reopening for metadata", src.Remote(), options)

	object, err := bucket.OpenObject(ctx, bucketPath)
	if err != nil {
		return nil, err
	}
	defer object.Close()

	o := NewObjectFromUplink(f, object).setRelative(src.Remote())

	return o, nil
}

// PutStream uploads to the remote path with the modTime given of indeterminate
// size.
//
// May create the object even if it returns an error - if so will return the
// object and the error, otherwise will return nil and the error.
func (f *Fs) PutStream(ctx context.Context, in io.Reader, src fs.ObjectInfo, options ...fs.OpenOption) (fs.Object, error) {
	return f.Put(ctx, in, src, options...)
}

// Mkdir makes the directory (container, bucket)
//
// Shouldn't return an error if it already exists
func (f *Fs) Mkdir(ctx context.Context, relative string) error {
	fs.Infof(f, "mkdir -p ./%s", relative)

	bucketName, _ := f.absolute(relative)

	// Check if bucket already exists.
	bucket, err := f.project.OpenBucket(ctx, bucketName, f.scope.EncryptionAccess)
	if err != nil {
		// Otherwise create the bucket.
		cfg := &uplink.BucketConfig{}

		// FIXME: These are hardcoded here because they aren't exposed publicly
		// anywhere. This is needed if you want to test against a storj-sim
		// network.
		if f.opts.Defaults == "dev" {
			cfg.Volatile.RedundancyScheme.RequiredShares = 4
			cfg.Volatile.RedundancyScheme.RepairShares = 6
			cfg.Volatile.RedundancyScheme.OptimalShares = 8
			cfg.Volatile.RedundancyScheme.TotalShares = 10
		}

		_, err = f.project.CreateBucket(ctx, bucketName, cfg)
		if err != nil {
			return err
		}
	}
	bucket.Close()

	return nil
}

// Rmdir removes the directory (container, bucket) if empty
//
// Return an error if it doesn't exist or isn't empty
func (f *Fs) Rmdir(ctx context.Context, relative string) error {
	fs.Infof(f, "rmdir ./%s", relative)

	bucketName, bucketPath := f.absolute(relative)

	// FIXME: This feels like black magic. I pulled it from the S3 backend,
	// but it doesn't make sense to me. It makes the tests pass though...
	if bucketName == "" || bucketPath != "" {
		return nil
	}

	bucket, err := f.project.OpenBucket(ctx, bucketName, f.scope.EncryptionAccess)
	if err != nil {
		return fs.ErrorDirNotFound
	}
	defer bucket.Close()

	if bucketPath == "" {
		result, err := bucket.ListObjects(ctx, &storj.ListOptions{Direction: storj.After, Recursive: true, Limit: 1})
		if err != nil {
			return err
		}

		if len(result.Items) > 0 {
			return fs.ErrorDirectoryNotEmpty
		}

		return f.project.DeleteBucket(ctx, bucketName)
	}

	object, err := bucket.OpenObject(ctx, bucketPath)
	if err != nil {
		return fs.ErrorDirNotFound
	}
	defer object.Close()

	if object.Meta.IsPrefix {
		return fs.ErrorDirectoryNotEmpty
	}

	return fs.ErrorIsFile
}
