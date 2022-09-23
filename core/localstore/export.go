package localstore

import (
	"archive/tar"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"sync"

	"github.com/redesblock/mop/core/postage"
	"github.com/redesblock/mop/core/shed"
	"github.com/redesblock/mop/core/storage"
	"github.com/redesblock/mop/core/swarm"
)

const (
	// filename in tar archive that holds the information
	// about exported data format version
	exportVersionFilename = ".swarm-export-version"
	// current export format version
	currentExportVersion = "3"
)

// Export writes a tar structured data to the writer of
// all chunks in the retrieval data index. It returns the
// number of chunks exported.
func (db *DB) Export(w io.Writer) (count int64, err error) {
	tw := tar.NewWriter(w)
	defer tw.Close()

	if err := tw.WriteHeader(&tar.Header{
		Name: exportVersionFilename,
		Mode: 0644,
		Size: int64(len(currentExportVersion)),
	}); err != nil {
		return 0, err
	}
	if _, err := tw.Write([]byte(currentExportVersion)); err != nil {
		return 0, err
	}

	err = db.retrievalDataIndex.Iterate(func(item shed.Item) (stop bool, err error) {

		hdr := &tar.Header{
			Name: hex.EncodeToString(item.Address),
			Mode: 0644,
			Size: int64(postage.StampSize + len(item.Data)),
		}

		if err := tw.WriteHeader(hdr); err != nil {
			return false, err
		}
		if _, err := tw.Write(item.BatchID); err != nil {
			return false, err
		}
		if _, err := tw.Write(item.Index); err != nil {
			return false, err
		}
		if _, err := tw.Write(item.Timestamp); err != nil {
			return false, err
		}
		if _, err := tw.Write(item.Sig); err != nil {
			return false, err
		}
		if _, err := tw.Write(item.Data); err != nil {
			return false, err
		}
		count++
		return false, nil
	}, nil)

	return count, err
}

// Import reads a tar structured data from the reader and
// stores chunks in the database. It returns the number of
// chunks imported.
func (db *DB) Import(ctx context.Context, r io.Reader) (count int64, err error) {
	tr := tar.NewReader(r)

	errC := make(chan error)
	doneC := make(chan struct{})
	tokenPool := make(chan struct{}, 100)
	var wg sync.WaitGroup
	go func() {
		var (
			firstFile = true

			// if exportVersionFilename file is not present
			// assume current version
			version = currentExportVersion
		)
		for {
			hdr, err := tr.Next()
			if err != nil {
				if err == io.EOF {
					break
				}
				select {
				case errC <- err:
				case <-ctx.Done():
				}
			}
			if firstFile {
				firstFile = false
				if hdr.Name == exportVersionFilename {
					data, err := io.ReadAll(tr)
					if err != nil {
						select {
						case errC <- err:
						case <-ctx.Done():
						}
					}
					version = string(data)
					continue
				}
			}

			if len(hdr.Name) != 64 {
				db.logger.Warningf("localstore export: ignoring non-chunk file: %s", hdr.Name)
				continue
			}

			keybytes, err := hex.DecodeString(hdr.Name)
			if err != nil {
				db.logger.Warningf("localstore export: ignoring invalid chunk file %s: %v", hdr.Name, err)
				continue
			}

			rawdata, err := io.ReadAll(tr)
			if err != nil {
				select {
				case errC <- err:
				case <-ctx.Done():
				}
			}
			stamp := new(postage.Stamp)
			err = stamp.UnmarshalBinary(rawdata[:postage.StampSize])
			if err != nil {
				select {
				case errC <- err:
				case <-ctx.Done():
				}
			}
			data := rawdata[postage.StampSize:]
			key := swarm.NewAddress(keybytes)

			var ch swarm.Chunk
			switch version {
			case currentExportVersion:
				ch = swarm.NewChunk(key, data).WithStamp(stamp)
			default:
				select {
				case errC <- fmt.Errorf("unsupported export data version %q", version):
				case <-ctx.Done():
				}
			}
			tokenPool <- struct{}{}
			wg.Add(1)

			go func() {
				_, err := db.Put(ctx, storage.ModePutUpload, ch)
				select {
				case errC <- err:
				case <-ctx.Done():
					wg.Done()
					<-tokenPool
				default:
					_, err := db.Put(ctx, storage.ModePutUpload, ch)
					if err != nil {
						errC <- err
					}
					wg.Done()
					<-tokenPool
				}
			}()

			count++
		}
		wg.Wait()
		close(doneC)
	}()

	// wait for all chunks to be stored
	for {
		select {
		case err := <-errC:
			if err != nil {
				return count, err
			}
		case <-ctx.Done():
			return count, ctx.Err()
		default:
			select {
			case <-doneC:
				return count, nil
			default:
			}
		}
	}
}
