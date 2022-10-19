package localstore

import (
	"archive/tar"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"sync"

	"github.com/redesblock/mop/core/cluster"
	"github.com/redesblock/mop/core/incentives/voucher"
	"github.com/redesblock/mop/core/storer/sharky"
	"github.com/redesblock/mop/core/storer/shed"
	"github.com/redesblock/mop/core/storer/storage"
)

const (
	// filename in tar archive that holds the information
	// about exported data format version
	exportVersionFilename = ".cluster-export-version"
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

		loc, err := sharky.LocationFromBinary(item.Location)
		if err != nil {
			return false, err
		}

		data := make([]byte, loc.Length)
		err = db.sharky.Read(context.TODO(), loc, data)
		if err != nil {
			return false, err
		}

		hdr := &tar.Header{
			Name: hex.EncodeToString(item.Address),
			Mode: 0644,
			Size: int64(voucher.StampSize + len(data)),
		}

		if err := tw.WriteHeader(hdr); err != nil {
			return false, err
		}
		write := func(buf []byte) {
			if err != nil {
				return
			}
			_, err = tw.Write(buf)
		}
		write(item.BatchID)
		write(item.Index)
		write(item.Timestamp)
		write(item.Sig)
		write(data)
		if err != nil {
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
				db.logger.Warning("export: ignoring non-chunk file", "name", hdr.Name)
				continue
			}

			keybytes, err := hex.DecodeString(hdr.Name)
			if err != nil {
				db.logger.Warning("export: ignoring invalid chunk file", "name", hdr.Name, "error", err)
				continue
			}

			rawdata, err := io.ReadAll(tr)
			if err != nil {
				select {
				case errC <- err:
				case <-ctx.Done():
				}
			}
			stamp := new(voucher.Stamp)
			err = stamp.UnmarshalBinary(rawdata[:voucher.StampSize])
			if err != nil {
				select {
				case errC <- err:
				case <-ctx.Done():
				}
			}
			data := rawdata[voucher.StampSize:]
			key := cluster.NewAddress(keybytes)

			var ch cluster.Chunk
			switch version {
			case currentExportVersion:
				ch = cluster.NewChunk(key, data).WithStamp(stamp)
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
