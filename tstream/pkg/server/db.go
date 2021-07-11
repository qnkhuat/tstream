package server

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/qnkhuat/tstream/pkg/message"
)

const (
	// Bucket names
	BROOMS string = "ROOMS"
)

type DB struct {
	*bolt.DB
}

func SetupDB(path string) (*DB, error) {

	bdb, err := bolt.Open(fmt.Sprintf("%s.boltdb", path), 0600, nil)

	if err != nil {
		return nil, fmt.Errorf("could not open db, %v", err)
	}

	err = bdb.Update(func(tx *bolt.Tx) error {

		// Store archived rooms
		_, err := tx.CreateBucketIfNotExists([]byte(BROOMS))
		if err != nil {
			return fmt.Errorf("could not create root bucket: %v", err)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("could not set up buckets, %v", err)
	}

	db := &DB{bdb} // wrap with DB wrapper
	return db, nil
}

// itob returns an 8-byte big endian representation of v.
func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}

func (db *DB) UpdateRooms(rooms map[uint64]message.RoomInfo) error {
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BROOMS))

    for id, msg := range rooms {
      buf, err := json.Marshal(msg)
      if err != nil {
        return err
      }

      err = b.Put(itob(id), []byte(buf))
      if err != nil {
        return err
      }  
    }
		return nil
	})

	return err
}

/*
DB
- ROOMS
  - ROOMID: ROOMINFO
  - ROOMID: ROOMINFO
ROOMID is auto increment
*/
func (db *DB) AddRoom(obj message.RoomInfo) (uint64, error) {
	var id uint64
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BROOMS))

		// newest record will be at the end of table
		id, _ = b.NextSequence()
    obj.Id = id

		buf, err := json.Marshal(obj)
		if err != nil {
			return err
		}

		b.Put(itob(id), []byte(buf))
		if err != nil {
			return fmt.Errorf("Failed to put: %v", err)
		}
		return nil
	})

	return id, err
}

// skip: number of records to skip
// n : number of records toget. Set to 0 to get all
// return a list of rooms with the first item is the lastest room
func (db *DB) GetRooms(statuses []message.RoomStatus, skip int, n int) ([]message.RoomInfo, error) {

	var rooms []message.RoomInfo

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BROOMS))
		c := b.Cursor()

		count := 0
		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			if n != 0 && count < skip {
				count += 1
				continue
			}

			// stop when get enough
			if n != 0 && count == (n+skip) {
				break
			}

			// decode room record
			room := message.RoomInfo{}
			err := json.Unmarshal(v, &room)
			if err != nil {
				return err
			}

			// filter by status
			// empty string to get all
      for _, status := range statuses {
        if room.Status == status  {
          rooms = append(rooms, room)
          count += 1
          break
        }
      }
    }

    return nil
  })
  return rooms, err
}
