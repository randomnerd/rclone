// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package streams

import (
	"context"
	"crypto/rand"
	"io"
	"io/ioutil"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/spacemonkeygo/monkit/v3"
	"github.com/zeebo/errs"
	"go.uber.org/zap"

	"storj.io/common/encryption"
	"storj.io/common/pb"
	"storj.io/common/ranger"
	"storj.io/common/storj"
	"storj.io/uplink/eestream"
	"storj.io/uplink/metainfo"
	"storj.io/uplink/storage/segments"
)

var mon = monkit.Package()

// Meta info about a stream
type Meta struct {
	Modified   time.Time
	Expiration time.Time
	Size       int64
	Data       []byte
}

// Store interface methods for streams to satisfy to be a store
type typedStore interface {
	Get(ctx context.Context, path Path, object storj.Object, pathCipher storj.CipherSuite) (ranger.Ranger, error)
	Put(ctx context.Context, path Path, pathCipher storj.CipherSuite, data io.Reader, metadata []byte, expiration time.Time) (Meta, error)
	Delete(ctx context.Context, path Path, pathCipher storj.CipherSuite) error
}

// streamStore is a store for streams. It implements typedStore as part of an ongoing migration
// to use typed paths. See the shim for the store that the rest of the world interacts with.
type streamStore struct {
	metainfo                *metainfo.Client
	segments                segments.Store
	segmentSize             int64
	encStore                *encryption.Store
	encBlockSize            int
	cipher                  storj.CipherSuite
	inlineThreshold         int
	maxEncryptedSegmentSize int64
}

// newTypedStreamStore constructs a typedStore backed by a streamStore.
func newTypedStreamStore(metainfo *metainfo.Client, segments segments.Store, segmentSize int64, encStore *encryption.Store, encBlockSize int, cipher storj.CipherSuite, inlineThreshold int, maxEncryptedSegmentSize int64) (typedStore, error) {
	if segmentSize <= 0 {
		return nil, errs.New("segment size must be larger than 0")
	}
	if encBlockSize <= 0 {
		return nil, errs.New("encryption block size must be larger than 0")
	}

	return &streamStore{
		metainfo:                metainfo,
		segments:                segments,
		segmentSize:             segmentSize,
		encStore:                encStore,
		encBlockSize:            encBlockSize,
		cipher:                  cipher,
		inlineThreshold:         inlineThreshold,
		maxEncryptedSegmentSize: maxEncryptedSegmentSize,
	}, nil
}

// Put breaks up data as it comes in into s.segmentSize length pieces, then
// store the first piece at s0/<path>, second piece at s1/<path>, and the
// *last* piece at l/<path>. Store the given metadata, along with the number
// of segments, in a new protobuf, in the metadata of l/<path>.
//
// If there is an error, it cleans up any uploaded segment before returning.
func (s *streamStore) Put(ctx context.Context, path Path, pathCipher storj.CipherSuite, data io.Reader, metadata []byte, expiration time.Time) (_ Meta, err error) {
	defer mon.Task()(&ctx)(&err)
	derivedKey, err := encryption.DeriveContentKey(path.Bucket(), path.UnencryptedPath(), s.encStore)
	if err != nil {
		return Meta{}, err
	}
	encPath, err := encryption.EncryptPath(path.Bucket(), path.UnencryptedPath(), pathCipher, s.encStore)
	if err != nil {
		return Meta{}, err
	}

	beginObjectReq := &metainfo.BeginObjectParams{
		Bucket:        []byte(path.Bucket()),
		EncryptedPath: []byte(encPath.Raw()),
		ExpiresAt:     expiration,
	}

	var (
		committedSegments int64
		streamID          storj.StreamID
	)
	defer func() {
		if err != nil {
			s.cancelHandler(context.Background(), streamID, committedSegments, path, pathCipher)
			return
		}

		select {
		case <-ctx.Done():
			s.cancelHandler(context.Background(), streamID, committedSegments, path, pathCipher)
		default:
		}
	}()

	var (
		currentSegment  int64
		contentKey      storj.Key
		streamSize      int64
		lastSegmentSize int64
		encryptedKey    []byte
		keyNonce        storj.Nonce
		rs              eestream.RedundancyStrategy

		requestsToBatch = make([]metainfo.BatchItem, 0, 2)
	)

	eofReader := NewEOFReader(data)
	for !eofReader.isEOF() && !eofReader.hasError() {
		// generate random key for encrypting the segment's content
		_, err := rand.Read(contentKey[:])
		if err != nil {
			return Meta{}, err
		}

		// Initialize the content nonce with the current total segment incremented
		// by 1 because at this moment the next segment has not been already
		// uploaded.
		// The increment by 1 is to avoid nonce reuse with the metadata encryption,
		// which is encrypted with the zero nonce.
		contentNonce := storj.Nonce{}
		_, err = encryption.Increment(&contentNonce, currentSegment+1)
		if err != nil {
			return Meta{}, err
		}

		// generate random nonce for encrypting the content key
		_, err = rand.Read(keyNonce[:])
		if err != nil {
			return Meta{}, err
		}

		encryptedKey, err = encryption.EncryptKey(&contentKey, s.cipher, derivedKey, &keyNonce)
		if err != nil {
			return Meta{}, err
		}

		sizeReader := NewSizeReader(eofReader)
		segmentReader := io.LimitReader(sizeReader, s.segmentSize)
		peekReader := segments.NewPeekThresholdReader(segmentReader)
		// If the data is larger than the inline threshold size, then it will be a remote segment
		isRemote, err := peekReader.IsLargerThan(s.inlineThreshold)
		if err != nil {
			return Meta{}, err
		}

		segmentEncryption := storj.SegmentEncryption{}
		if s.cipher != storj.EncNull {
			segmentEncryption = storj.SegmentEncryption{
				EncryptedKey:      encryptedKey,
				EncryptedKeyNonce: keyNonce,
			}
		}

		if isRemote {
			encrypter, err := encryption.NewEncrypter(s.cipher, &contentKey, &contentNonce, s.encBlockSize)
			if err != nil {
				return Meta{}, err
			}

			paddedReader := encryption.PadReader(ioutil.NopCloser(peekReader), encrypter.InBlockSize())
			transformedReader := encryption.TransformReader(paddedReader, encrypter, 0)

			beginSegment := &metainfo.BeginSegmentParams{
				MaxOrderLimit: s.maxEncryptedSegmentSize,
				Position: storj.SegmentPosition{
					Index: int32(currentSegment),
				},
			}

			var responses []metainfo.BatchResponse
			if currentSegment == 0 {
				responses, err = s.metainfo.Batch(ctx, beginObjectReq, beginSegment)
				if err != nil {
					return Meta{}, err
				}
				objResponse, err := responses[0].BeginObject()
				if err != nil {
					return Meta{}, err
				}
				streamID = objResponse.StreamID
				rs = objResponse.RedundancyStrategy
			} else {
				beginSegment.StreamID = streamID
				responses, err = s.metainfo.Batch(ctx, append(requestsToBatch, beginSegment)...)
				requestsToBatch = requestsToBatch[:0]
				if len(responses) > 0 {
					// We increment because the first request has succeeded
					committedSegments++
				}
				if err != nil {
					return Meta{}, err
				}
			}

			segResponse, err := responses[1].BeginSegment()
			if err != nil {
				return Meta{}, err
			}
			segmentID := segResponse.SegmentID
			limits := segResponse.Limits
			piecePrivateKey := segResponse.PiecePrivateKey

			uploadResults, size, err := s.segments.Put(ctx, transformedReader, expiration, limits, piecePrivateKey, rs)
			if err != nil {
				return Meta{}, err
			}

			requestsToBatch = append(requestsToBatch, &metainfo.CommitSegmentParams{
				SegmentID:         segmentID,
				SizeEncryptedData: size,
				Encryption:        segmentEncryption,
				UploadResult:      uploadResults,
			})
		} else {
			data, err := ioutil.ReadAll(peekReader)
			if err != nil {
				return Meta{}, err
			}
			cipherData, err := encryption.Encrypt(data, s.cipher, &contentKey, &contentNonce)
			if err != nil {
				return Meta{}, err
			}

			makeInlineSegment := &metainfo.MakeInlineSegmentParams{
				Position: storj.SegmentPosition{
					Index: int32(currentSegment),
				},
				Encryption:          segmentEncryption,
				EncryptedInlineData: cipherData,
			}
			if currentSegment == 0 {
				responses, err := s.metainfo.Batch(ctx, beginObjectReq, makeInlineSegment)
				if err != nil {
					return Meta{}, err
				}
				objResponse, err := responses[0].BeginObject()
				if err != nil {
					return Meta{}, err
				}
				streamID = objResponse.StreamID
			} else {
				makeInlineSegment.StreamID = streamID
				requestsToBatch = append(requestsToBatch, makeInlineSegment)
			}

			committedSegments++
		}

		lastSegmentSize = sizeReader.Size()
		streamSize += lastSegmentSize
		currentSegment++
	}

	totalSegments := currentSegment

	if eofReader.hasError() {
		return Meta{}, eofReader.err
	}

	streamInfo, err := proto.Marshal(&pb.StreamInfo{
		DeprecatedNumberOfSegments: totalSegments,
		SegmentsSize:               s.segmentSize,
		LastSegmentSize:            lastSegmentSize,
		Metadata:                   metadata,
	})
	if err != nil {
		return Meta{}, err
	}

	// encrypt metadata with the content encryption key and zero nonce
	encryptedStreamInfo, err := encryption.Encrypt(streamInfo, s.cipher, &contentKey, &storj.Nonce{})
	if err != nil {
		return Meta{}, err
	}

	streamMeta := pb.StreamMeta{
		NumberOfSegments:    totalSegments,
		EncryptedStreamInfo: encryptedStreamInfo,
		EncryptionType:      int32(s.cipher),
		EncryptionBlockSize: int32(s.encBlockSize),
	}

	if s.cipher != storj.EncNull {
		streamMeta.LastSegmentMeta = &pb.SegmentMeta{
			EncryptedKey: encryptedKey,
			KeyNonce:     keyNonce[:],
		}
	}

	objectMetadata, err := proto.Marshal(&streamMeta)
	if err != nil {
		return Meta{}, err
	}

	commitObject := metainfo.CommitObjectParams{
		StreamID:          streamID,
		EncryptedMetadata: objectMetadata,
	}
	if len(requestsToBatch) > 0 {
		var responses []metainfo.BatchResponse
		responses, err = s.metainfo.Batch(ctx, append(requestsToBatch, &commitObject)...)
		if len(responses) > 0 {
			// We increment because the first request has succeeded
			committedSegments++
		}
	} else {
		err = s.metainfo.CommitObject(ctx, commitObject)
	}
	if err != nil {
		return Meta{}, err
	}

	satStreamID := &pb.SatStreamID{}
	err = proto.Unmarshal(streamID, satStreamID)
	if err != nil {
		return Meta{}, err
	}

	resultMeta := Meta{
		Modified:   satStreamID.CreationDate,
		Expiration: expiration,
		Size:       streamSize,
		Data:       metadata,
	}

	return resultMeta, nil
}

// Get returns a ranger that knows what the overall size is (from l/<path>)
// and then returns the appropriate data from segments s0/<path>, s1/<path>,
// ..., l/<path>.
func (s *streamStore) Get(ctx context.Context, path Path, object storj.Object, pathCipher storj.CipherSuite) (rr ranger.Ranger, err error) {
	defer mon.Task()(&ctx)(&err)

	info, limits, err := s.metainfo.DownloadSegment(ctx, metainfo.DownloadSegmentParams{
		StreamID: object.ID,
		Position: storj.SegmentPosition{
			Index: -1, // Request the last segment
		},
	})
	if err != nil {
		return nil, err
	}

	lastSegmentRanger, err := s.segments.Ranger(ctx, info, limits, object.RedundancyScheme)
	if err != nil {
		return nil, err
	}

	derivedKey, err := encryption.DeriveContentKey(path.Bucket(), path.UnencryptedPath(), s.encStore)
	if err != nil {
		return nil, err
	}

	var rangers []ranger.Ranger
	for i := int64(0); i < object.SegmentCount-1; i++ {
		var contentNonce storj.Nonce
		_, err = encryption.Increment(&contentNonce, i+1)
		if err != nil {
			return nil, err
		}

		rangers = append(rangers, &lazySegmentRanger{
			metainfo:      s.metainfo,
			segments:      s.segments,
			streamID:      object.ID,
			segmentIndex:  int32(i),
			rs:            object.RedundancyScheme,
			size:          object.FixedSegmentSize,
			derivedKey:    derivedKey,
			startingNonce: &contentNonce,
			encBlockSize:  int(object.EncryptionParameters.BlockSize),
			cipher:        object.CipherSuite,
		})
	}

	var contentNonce storj.Nonce
	_, err = encryption.Increment(&contentNonce, object.SegmentCount)
	if err != nil {
		return nil, err
	}

	decryptedLastSegmentRanger, err := decryptRanger(
		ctx,
		lastSegmentRanger,
		object.LastSegment.Size,
		object.CipherSuite,
		derivedKey,
		info.SegmentEncryption.EncryptedKey,
		&info.SegmentEncryption.EncryptedKeyNonce,
		&contentNonce,
		int(object.EncryptionParameters.BlockSize),
	)
	if err != nil {
		return nil, err
	}

	rangers = append(rangers, decryptedLastSegmentRanger)
	return ranger.Concat(rangers...), nil
}

// Delete all the segments, with the last one last
func (s *streamStore) Delete(ctx context.Context, path Path, pathCipher storj.CipherSuite) (err error) {
	defer mon.Task()(&ctx)(&err)

	encPath, err := encryption.EncryptPath(path.Bucket(), path.UnencryptedPath(), pathCipher, s.encStore)
	if err != nil {
		return err
	}

	_, err = s.metainfo.BeginDeleteObject(ctx, metainfo.BeginDeleteObjectParams{
		Bucket:        []byte(path.Bucket()),
		EncryptedPath: []byte(encPath.Raw()),
	})
	return err
}

// ListItem is a single item in a listing
type ListItem struct {
	Path     string
	Meta     Meta
	IsPrefix bool
}

type lazySegmentRanger struct {
	ranger        ranger.Ranger
	metainfo      *metainfo.Client
	segments      segments.Store
	streamID      storj.StreamID
	segmentIndex  int32
	rs            storj.RedundancyScheme
	size          int64
	derivedKey    *storj.Key
	startingNonce *storj.Nonce
	encBlockSize  int
	cipher        storj.CipherSuite
}

// Size implements Ranger.Size.
func (lr *lazySegmentRanger) Size() int64 {
	return lr.size
}

// Range implements Ranger.Range to be lazily connected.
func (lr *lazySegmentRanger) Range(ctx context.Context, offset, length int64) (_ io.ReadCloser, err error) {
	defer mon.Task()(&ctx)(&err)
	if lr.ranger == nil {
		info, limits, err := lr.metainfo.DownloadSegment(ctx, metainfo.DownloadSegmentParams{
			StreamID: lr.streamID,
			Position: storj.SegmentPosition{
				Index: lr.segmentIndex,
			},
		})
		if err != nil {
			return nil, err
		}

		rr, err := lr.segments.Ranger(ctx, info, limits, lr.rs)
		if err != nil {
			return nil, err
		}

		encryptedKey, keyNonce := info.SegmentEncryption.EncryptedKey, info.SegmentEncryption.EncryptedKeyNonce
		lr.ranger, err = decryptRanger(ctx, rr, lr.size, lr.cipher, lr.derivedKey, encryptedKey, &keyNonce, lr.startingNonce, lr.encBlockSize)
		if err != nil {
			return nil, err
		}
	}
	return lr.ranger.Range(ctx, offset, length)
}

// decryptRanger returns a decrypted ranger of the given rr ranger.
func decryptRanger(ctx context.Context, rr ranger.Ranger, decryptedSize int64, cipher storj.CipherSuite, derivedKey *storj.Key, encryptedKey storj.EncryptedPrivateKey, encryptedKeyNonce, startingNonce *storj.Nonce, encBlockSize int) (decrypted ranger.Ranger, err error) {
	defer mon.Task()(&ctx)(&err)
	contentKey, err := encryption.DecryptKey(encryptedKey, cipher, derivedKey, encryptedKeyNonce)
	if err != nil {
		return nil, err
	}

	decrypter, err := encryption.NewDecrypter(cipher, contentKey, startingNonce, encBlockSize)
	if err != nil {
		return nil, err
	}

	var rd ranger.Ranger
	if rr.Size()%int64(decrypter.InBlockSize()) != 0 {
		reader, err := rr.Range(ctx, 0, rr.Size())
		if err != nil {
			return nil, err
		}
		defer func() { err = errs.Combine(err, reader.Close()) }()
		cipherData, err := ioutil.ReadAll(reader)
		if err != nil {
			return nil, err
		}
		data, err := encryption.Decrypt(cipherData, cipher, contentKey, startingNonce)
		if err != nil {
			return nil, err
		}
		return ranger.ByteRanger(data), nil
	}

	rd, err = encryption.Transform(rr, decrypter)
	if err != nil {
		return nil, err
	}
	return encryption.Unpad(rd, int(rd.Size()-decryptedSize))
}

// CancelHandler handles clean up of segments on receiving CTRL+C
func (s *streamStore) cancelHandler(ctx context.Context, streamID storj.StreamID, totalSegments int64, path Path, pathCipher storj.CipherSuite) {
	defer mon.Task()(&ctx)(nil)

	for i := int64(0); i < totalSegments; i++ {
		err := s.segments.Delete(ctx, streamID, int32(i))
		if err != nil {
			zap.L().Warn("Failed deleting segment", zap.Stringer("path", path), zap.Int64("Segment Index", i), zap.Error(err))
			continue
		}
	}
}

func getEncryptedKeyAndNonce(m *pb.SegmentMeta) (storj.EncryptedPrivateKey, *storj.Nonce) {
	if m == nil {
		return nil, nil
	}

	var nonce storj.Nonce
	copy(nonce[:], m.KeyNonce)

	return m.EncryptedKey, &nonce
}

// TypedDecryptStreamInfo decrypts stream info
func TypedDecryptStreamInfo(ctx context.Context, streamMetaBytes []byte, path Path, encStore *encryption.Store) (
	_ *pb.StreamInfo, streamMeta pb.StreamMeta, err error) {
	defer mon.Task()(&ctx)(&err)

	err = proto.Unmarshal(streamMetaBytes, &streamMeta)
	if err != nil {
		return nil, pb.StreamMeta{}, err
	}

	if encStore.EncryptionBypass {
		return nil, streamMeta, nil
	}

	derivedKey, err := encryption.DeriveContentKey(path.Bucket(), path.UnencryptedPath(), encStore)
	if err != nil {
		return nil, pb.StreamMeta{}, err
	}

	cipher := storj.CipherSuite(streamMeta.EncryptionType)
	encryptedKey, keyNonce := getEncryptedKeyAndNonce(streamMeta.LastSegmentMeta)
	contentKey, err := encryption.DecryptKey(encryptedKey, cipher, derivedKey, keyNonce)
	if err != nil {
		return nil, pb.StreamMeta{}, err
	}

	// decrypt metadata with the content encryption key and zero nonce
	streamInfo, err := encryption.Decrypt(streamMeta.EncryptedStreamInfo, cipher, contentKey, &storj.Nonce{})
	if err != nil {
		return nil, pb.StreamMeta{}, err
	}

	var stream pb.StreamInfo
	if err := proto.Unmarshal(streamInfo, &stream); err != nil {
		return nil, pb.StreamMeta{}, err
	}

	return &stream, streamMeta, nil
}
