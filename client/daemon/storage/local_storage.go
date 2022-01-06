/*
 *     Copyright 2020 The Dragonfly Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package storage

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path"
	"sync"
	"time"

	"go.uber.org/atomic"

	"d7y.io/dragonfly/v2/client/clientutil"
	logger "d7y.io/dragonfly/v2/internal/dflog"
	"d7y.io/dragonfly/v2/pkg/rpc/base"
	"d7y.io/dragonfly/v2/pkg/util/digestutils"
)

type localTaskStore struct {
	*logger.SugaredLoggerOnWith
	persistentMetadata

	sync.RWMutex

	dataDir string

	metadataFile     *os.File
	metadataFilePath string

	expireTime    time.Duration
	lastAccess    atomic.Int64
	reclaimMarked atomic.Bool
	gcCallback    func(CommonTaskRequest)

	// when digest not match, invalid will be set
	invalid atomic.Bool
}

var _ TaskStorageDriver = (*localTaskStore)(nil)
var _ Reclaimer = (*localTaskStore)(nil)

func (t *localTaskStore) touch() {
	access := time.Now().UnixNano()
	t.lastAccess.Store(access)
}

func (t *localTaskStore) WritePiece(ctx context.Context, req *WritePieceRequest) (int64, error) {
	t.touch()

	// piece already exists
	t.RLock()
	if piece, ok := t.Pieces[req.Num]; ok {
		t.RUnlock()
		// discard data for back source
		n, err := io.Copy(io.Discard, io.LimitReader(req.Reader, req.Range.Length))
		if err != nil && err != io.EOF {
			return n, err
		}
		if n != piece.Range.Length {
			return n, ErrShortRead
		}
		return piece.Range.Length, nil
	}
	t.RUnlock()

	file, err := os.OpenFile(t.DataFilePath, os.O_RDWR, defaultFileMode)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	if _, err = file.Seek(req.Range.Start, io.SeekStart); err != nil {
		return 0, err
	}
	var (
		r  *io.LimitedReader
		ok bool
		bn int64 // copied bytes from BufferedReader.B
	)
	if r, ok = req.Reader.(*io.LimitedReader); ok {
		// by jim: drain buffer and use raw reader(normally tcp connection) for using optimised operator, like splice
		if br, bok := r.R.(*clientutil.BufferedReader); bok {
			bn, err := io.CopyN(file, br.B, int64(br.B.Buffered()))
			if err != nil && err != io.EOF {
				return 0, err
			}
			r = io.LimitReader(br.R, r.N-bn).(*io.LimitedReader)
		}
	} else {
		r = io.LimitReader(req.Reader, req.Range.Length).(*io.LimitedReader)
	}
	n, err := io.Copy(file, r)
	if err != nil {
		return 0, err
	}
	// when UnknownLength and size is align to piece num
	if req.UnknownLength && n == 0 {
		return 0, nil
	}
	// update copied bytes from BufferedReader.B
	n += bn
	if n != req.Range.Length {
		if req.UnknownLength {
			// when back source, and can not detect content length, we need update real length
			req.Range.Length = n
			// when n == 0, skip
			if n == 0 {
				return 0, nil
			}
		} else {
			return n, ErrShortRead
		}
	}
	// when Md5 is empty, try to get md5 from reader, it's useful for back source
	if req.PieceMetadata.Md5 == "" {
		t.Warnf("piece md5 not found in metadata, read from reader")
		if get, ok := req.Reader.(digestutils.DigestReader); ok {
			req.PieceMetadata.Md5 = get.Digest()
			t.Infof("read md5 from reader, value: %s", req.PieceMetadata.Md5)
		} else {
			t.Warnf("reader is not a DigestReader")
		}
	}
	t.Debugf("wrote %d bytes to file %s, piece %d, start %d, length: %d",
		n, t.DataFilePath, req.Num, req.Range.Start, req.Range.Length)
	t.Lock()
	defer t.Unlock()
	// double check
	if _, ok := t.Pieces[req.Num]; ok {
		return n, nil
	}
	t.Pieces[req.Num] = req.PieceMetadata
	return n, nil
}

func (t *localTaskStore) UpdateTask(ctx context.Context, req *UpdateTaskRequest) error {
	t.touch()
	t.Lock()
	defer t.Unlock()
	t.persistentMetadata.ContentLength = req.ContentLength
	if req.TotalPieces > 0 {
		t.TotalPieces = req.TotalPieces
		t.Debugf("update total pieces: %d", t.TotalPieces)
	}
	if len(t.PieceMd5Sign) == 0 {
		t.PieceMd5Sign = req.PieceMd5Sign
		t.Debugf("update piece md5 sign: %s", t.PieceMd5Sign)
	}
	if req.GenPieceDigest {
		var pieceDigests []string
		for i := int32(0); i < t.TotalPieces; i++ {
			pieceDigests = append(pieceDigests, t.Pieces[i].Md5)
		}

		digest := digestutils.Sha256(pieceDigests...)
		t.PieceMd5Sign = digest
		t.Infof("generated digest: %s", digest)
	}
	return nil
}

func (t *localTaskStore) ValidateDigest(*PeerTaskMetadata) error {
	t.Lock()
	defer t.Unlock()
	if t.persistentMetadata.PieceMd5Sign == "" {
		t.invalid.Store(true)
		return ErrDigestNotSet
	}
	if t.TotalPieces <= 0 {
		t.Errorf("total piece count not set when validate digest")
		t.invalid.Store(true)
		return ErrPieceCountNotSet
	}

	var pieceDigests []string
	for i := int32(0); i < t.TotalPieces; i++ {
		pieceDigests = append(pieceDigests, t.Pieces[i].Md5)
	}

	digest := digestutils.Sha256(pieceDigests...)
	if digest != t.PieceMd5Sign {
		t.Errorf("invalid digest, desired: %s, actual: %s", t.PieceMd5Sign, digest)
		t.invalid.Store(true)
		return ErrInvalidDigest
	}
	return nil
}

func (t *localTaskStore) IsInvalid(*PeerTaskMetadata) (bool, error) {
	return t.invalid.Load(), nil
}

// ReadPiece get a LimitReadCloser from task data with seeked, caller should read bytes and close it.
func (t *localTaskStore) ReadPiece(ctx context.Context, req *ReadPieceRequest) (io.Reader, io.Closer, error) {
	if t.invalid.Load() {
		t.Errorf("invalid digest, refuse to get pieces")
		return nil, nil, ErrInvalidDigest
	}

	t.touch()
	file, err := os.Open(t.DataFilePath)
	if err != nil {
		return nil, nil, err
	}

	// If req.Num is equal to -1, range has a fixed value.
	if req.Num != -1 {
		t.RLock()
		if piece, ok := t.persistentMetadata.Pieces[req.Num]; ok {
			t.RUnlock()
			req.Range = piece.Range
		} else {
			t.RUnlock()
			file.Close()
			t.Errorf("invalid piece num: %d", req.Num)
			return nil, nil, ErrPieceNotFound
		}
	}

	if _, err = file.Seek(req.Range.Start, io.SeekStart); err != nil {
		file.Close()
		t.Errorf("file seek failed: %v", err)
		return nil, nil, err
	}
	// who call ReadPiece, who close the io.ReadCloser
	return io.LimitReader(file, req.Range.Length), file, nil
}

func (t *localTaskStore) ReadAllPieces(ctx context.Context, req *PeerTaskMetadata) (io.ReadCloser, error) {
	if t.invalid.Load() {
		t.Errorf("invalid digest, refuse to read all pieces")
		return nil, ErrInvalidDigest
	}

	t.touch()
	file, err := os.Open(t.DataFilePath)
	if err != nil {
		return nil, err
	}
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		file.Close()
		t.Errorf("file seek failed: %v", err)
		return nil, err
	}
	// who call ReadPiece, who close the io.ReadCloser
	return file, nil
}

func (t *localTaskStore) Store(ctx context.Context, req *StoreRequest) error {
	// Store is be called in callback.Done, mark local task store done, for fast search
	t.Done = true
	t.touch()
	if req.TotalPieces > 0 {
		t.Lock()
		t.TotalPieces = req.TotalPieces
		t.Unlock()
	}
	if !req.StoreOnly {
		err := t.saveMetadata()
		if err != nil {
			t.Warnf("save task metadata error: %s", err)
			return err
		}
	}
	if req.MetadataOnly {
		return nil
	}

	src := t.DataFilePath
	dst := req.Destination
	// If requested to add to storage, req.Destination is source, which will be added to storage
	if req.AddToStorage {
		src = dst
		dst = t.DataFilePath
	}
	_, err := os.Stat(dst)
	if err == nil {
		// remove exist file
		t.Infof("destination file %q exists, purge it first", dst)
		os.Remove(dst)
	}
	// 1. try to link
	err = os.Link(src, dst)
	if err == nil {
		t.Infof("task data link to file %q success", dst)
		return nil
	}
	t.Warnf("task data link to file %q error: %s", dst, err)
	// 2. link failed, copy it
	file, err := os.Open(src)
	if err != nil {
		t.Debugf("open tasks data error: %s", err)
		return err
	}
	defer file.Close()

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		t.Debugf("task seek file error: %s", err)
		return err
	}
	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_RDWR|os.O_TRUNC, defaultFileMode)
	if err != nil {
		t.Errorf("open tasks destination file error: %s", err)
		return err
	}
	defer dstFile.Close()
	// copy_file_range is valid in linux
	// https://go-review.googlesource.com/c/go/+/229101/
	n, err := io.Copy(dstFile, file)
	t.Debugf("copied tasks data %d bytes to %s", n, req.Destination)
	return err
}

func (t *localTaskStore) GetPieces(ctx context.Context, req *base.PieceTaskRequest) (*base.PiecePacket, error) {
	if t.invalid.Load() {
		t.Errorf("invalid digest, refuse to get pieces")
		return nil, ErrInvalidDigest
	}

	t.RLock()
	defer t.RUnlock()
	t.touch()
	piecePacket := &base.PiecePacket{
		TaskId:        req.TaskId,
		DstPid:        t.PeerID,
		TotalPiece:    t.TotalPieces,
		ContentLength: t.ContentLength,
		PieceMd5Sign:  t.PieceMd5Sign,
	}
	if t.TotalPieces > -1 && int32(req.StartNum) >= t.TotalPieces {
		t.Warnf("invalid start num: %d", req.StartNum)
	}
	for i := int32(0); i < int32(req.Limit); i++ {
		if piece, ok := t.Pieces[int32(req.StartNum)+i]; ok {
			piecePacket.PieceInfos = append(piecePacket.PieceInfos, &base.PieceInfo{
				PieceNum:    piece.Num,
				RangeStart:  uint64(piece.Range.Start),
				RangeSize:   uint32(piece.Range.Length),
				PieceMd5:    piece.Md5,
				PieceOffset: piece.Offset,
				PieceStyle:  piece.Style,
			})
		}
	}
	return piecePacket, nil
}

func (t *localTaskStore) CanReclaim() bool {
	access := time.Unix(0, t.lastAccess.Load())
	reclaim := access.Add(t.expireTime).Before(time.Now())
	t.Debugf("reclaim check, last access: %v, reclaim: %v", access, reclaim)
	return reclaim
}

// MarkReclaim will try to invoke gcCallback (normal leave peer task)
func (t *localTaskStore) MarkReclaim() {
	if t.reclaimMarked.Load() {
		return
	}
	// leave task
	t.gcCallback(CommonTaskRequest{
		PeerID: t.PeerID,
		TaskID: t.TaskID,
	})
	t.reclaimMarked.Store(true)
	t.Infof("task %s/%s will be reclaimed, marked", t.TaskID, t.PeerID)
}

func (t *localTaskStore) Reclaim() error {
	t.Infof("start gc task data")
	err := t.reclaimData()
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// close and remove metadata
	err = t.reclaimMeta()
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// remove task work metaDir
	if err = os.Remove(t.dataDir); err != nil && !os.IsNotExist(err) {
		t.Warnf("remove task data directory %q error: %s", t.dataDir, err)
		return err
	}
	t.Infof("purged task work directory: %s", t.dataDir)

	taskDir := path.Dir(t.dataDir)
	if dirs, err := os.ReadDir(taskDir); err != nil {
		t.Warnf("stat task directory %q error: %s", taskDir, err)
	} else {
		if len(dirs) == 0 {
			if err := os.Remove(taskDir); err != nil {
				t.Warnf("remove unused task directory %q error: %s", taskDir, err)
			}
		} else {
			t.Warnf("task directory %q is not empty", taskDir)
		}
	}
	return nil
}

func (t *localTaskStore) reclaimData() error {
	// remove data
	data := path.Join(t.dataDir, taskData)
	stat, err := os.Lstat(data)
	if err != nil {
		t.Errorf("stat task data %q error: %s", data, err)
		return err
	}
	// remove symbol link cache file
	if stat.Mode()&os.ModeSymlink == os.ModeSymlink {
		dest, err0 := os.Readlink(data)
		if err0 == nil {
			if err = os.Remove(dest); err != nil && !os.IsNotExist(err) {
				t.Warnf("remove symlink target file %s error: %s", dest, err)
			} else {
				t.Infof("remove data file %s", dest)
			}
		}
	} else { // remove cache file
		if err = os.Remove(t.DataFilePath); err != nil && !os.IsNotExist(err) {
			t.Errorf("remove data file %s error: %s", data, err)
			return err
		}
	}
	if err = os.Remove(data); err != nil && !os.IsNotExist(err) {
		t.Errorf("remove data file %s error: %s", data, err)
		return err
	}
	t.Infof("purged task data: %s", data)
	return nil
}

func (t *localTaskStore) reclaimMeta() error {
	if err := t.metadataFile.Close(); err != nil {
		t.Warnf("close task meta data %q error: %s", t.metadataFilePath, err)
		return err
	}
	t.Infof("start gc task metadata")
	if err := os.Remove(t.metadataFilePath); err != nil && !os.IsNotExist(err) {
		t.Warnf("remove task meta data %q error: %s", t.metadataFilePath, err)
		return err
	}
	t.Infof("purged task mata data: %s", t.metadataFilePath)
	return nil
}

func (t *localTaskStore) saveMetadata() error {
	t.Lock()
	defer t.Unlock()
	data, err := json.Marshal(t.persistentMetadata)
	if err != nil {
		return err
	}
	_, err = t.metadataFile.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	_, err = t.metadataFile.Write(data)
	if err != nil {
		t.Errorf("save metadata error: %s", err)
	}
	return err
}
