package localstore

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/redesblock/mop/core/postage"
	"github.com/redesblock/mop/core/shed"
	"github.com/syndtr/goleveldb/leveldb"
)

// DBSchemaBatchIndex is the mop schema identifier for dead-push.
const DBSchemaDeadPush = "dead-push"

// migrateDeadPush cleans up dangling push index entries that make the pusher stop pushing entries
func migrateDeadPush(db *DB) error {
	start := time.Now()
	db.logger.Debug("removing dangling entries from push index")
	batch := new(leveldb.Batch)
	count := 0
	headerSize := 16 + postage.VouchSize
	retrievalDataIndex, err := db.shed.NewIndex("Address->StoreTimestamp|BinID|BatchID|BatchIndex|Sig|Data", shed.IndexFuncs{
		EncodeKey: func(fields shed.Item) (key []byte, err error) {
			return fields.Address, nil
		},
		DecodeKey: func(key []byte) (e shed.Item, err error) {
			e.Address = key
			return e, nil
		},
		EncodeValue: func(fields shed.Item) (value []byte, err error) {
			b := make([]byte, headerSize)
			binary.BigEndian.PutUint64(b[:8], fields.BinID)
			binary.BigEndian.PutUint64(b[8:16], uint64(fields.StoreTimestamp))
			vouch, err := postage.NewVouch(fields.BatchID, fields.Index, fields.Timestamp, fields.Sig).MarshalBinary()
			if err != nil {
				return nil, err
			}
			copy(b[16:], vouch)
			value = append(b, fields.Data...)
			return value, nil
		},
		DecodeValue: func(keyItem shed.Item, value []byte) (e shed.Item, err error) {
			e.StoreTimestamp = int64(binary.BigEndian.Uint64(value[8:16]))
			e.BinID = binary.BigEndian.Uint64(value[:8])
			vouch := new(postage.Vouch)
			if err = vouch.UnmarshalBinary(value[16:headerSize]); err != nil {
				return e, err
			}
			e.BatchID = vouch.BatchID()
			e.Index = vouch.Index()
			e.Timestamp = vouch.Timestamp()
			e.Sig = vouch.Sig()
			e.Data = value[headerSize:]
			return e, nil
		},
	})
	if err != nil {
		return err
	}
	pushIndex, err := db.shed.NewIndex("StoreTimestamp|Hash->Tags", shed.IndexFuncs{
		EncodeKey: func(fields shed.Item) (key []byte, err error) {
			key = make([]byte, 40)
			binary.BigEndian.PutUint64(key[:8], uint64(fields.StoreTimestamp))
			copy(key[8:], fields.Address)
			return key, nil
		},
		DecodeKey: func(key []byte) (e shed.Item, err error) {
			e.Address = key[8:]
			e.StoreTimestamp = int64(binary.BigEndian.Uint64(key[:8]))
			return e, nil
		},
		EncodeValue: func(fields shed.Item) (value []byte, err error) {
			tag := make([]byte, 4)
			binary.BigEndian.PutUint32(tag, fields.Tag)
			return tag, nil
		},
		DecodeValue: func(keyItem shed.Item, value []byte) (e shed.Item, err error) {
			if len(value) == 4 { // only values with tag should be decoded
				e.Tag = binary.BigEndian.Uint32(value)
			}
			return e, nil
		},
	})
	if err != nil {
		return err
	}
	err = pushIndex.Iterate(func(item shed.Item) (stop bool, err error) {
		has, err := retrievalDataIndex.Has(item)
		if err != nil {
			return true, err
		}
		if !has {
			if err = pushIndex.DeleteInBatch(batch, item); err != nil {
				return true, err
			}
			count++
		}
		return false, nil
	}, nil)
	if err != nil {
		return fmt.Errorf("iterate index: %w", err)
	}
	db.logger.Debugf("found %d entries to remove. trying to flush...", count)
	err = db.shed.WriteBatch(batch)
	if err != nil {
		return fmt.Errorf("write batch: %w", err)
	}
	db.logger.Debugf("done cleaning index. took %s", time.Since(start))
	return nil
}
