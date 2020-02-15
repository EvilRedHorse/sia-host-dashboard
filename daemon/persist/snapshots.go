package persist

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/siacentral/host-dashboard/daemon/types"
	"gitlab.com/NebulousLabs/bolt"
)

//SaveHostSnapshot SaveHostSnapshot
func SaveHostSnapshot(snapshot types.HostSnapshot) error {
	snapshot.Timestamp = snapshot.Timestamp.Truncate(time.Hour).UTC()

	return db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketHostSnapshots)

		buf, err := json.Marshal(snapshot)

		if err != nil {
			return fmt.Errorf("json encode: %s", err)
		}

		bucket.Put(timeID(snapshot.Timestamp), buf)

		return nil
	})
}

//GetHostSnapshots returns all snapshots between two timestamps (inclusive)
func GetHostSnapshots(start, end time.Time) (snapshots []types.HostSnapshot, err error) {
	if start.After(end) {
		err = errors.New("start must be before end")
		return
	}

	startID := timeID(start)
	endID := timeID(end)

	err = db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(bucketHostSnapshots).Cursor()

		for key, buf := c.Seek(startID); key != nil; key, buf = c.Next() {
			if bytes.Compare(key, endID) > 0 {
				break
			}

			var snapshot types.HostSnapshot

			if err := json.Unmarshal(buf, &snapshot); err != nil {
				return err
			}

			snapshots = append(snapshots, snapshot)
		}

		return nil
	})

	return
}

// GetDailySnapshots returns snapshot totals for every day between two timestamps (inclusive)
func GetDailySnapshots(start, end time.Time) (snapshots []types.HostSnapshot, err error) {
	if start.After(end) {
		err = errors.New("start must be before end")
		return
	}

	snapshots = append(snapshots, types.HostSnapshot{
		Timestamp: start,
	})

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketHostSnapshots)
		next := start.AddDate(0, 0, 1)

		for current := start; current.Before(end); current = current.Add(time.Hour) {
			var snapshot types.HostSnapshot

			id := timeID(current)
			i := len(snapshots) - 1
			buf := b.Get(id)

			if current.After(next) {
				snapshots = append(snapshots, types.HostSnapshot{
					Timestamp: next,
				})
				next = next.AddDate(0, 0, 1)
				i++
			}

			if buf == nil {
				continue
			}

			if err = json.Unmarshal(buf, &snapshot); err != nil {
				return err
			}

			snapshots[i].ActiveContracts = snapshot.ActiveContracts
			snapshots[i].NewContracts += snapshot.NewContracts
			snapshots[i].ExpiredContracts += snapshot.ExpiredContracts
			snapshots[i].SuccessfulContracts += snapshot.SuccessfulContracts
			snapshots[i].FailedContracts += snapshot.FailedContracts

			snapshots[i].Payout = snapshots[i].Payout.Add(snapshot.Payout)
			snapshots[i].EarnedRevenue = snapshots[i].EarnedRevenue.Add(snapshot.EarnedRevenue)
			snapshots[i].PotentialRevenue = snapshots[i].PotentialRevenue.Add(snapshot.PotentialRevenue)
			snapshots[i].BurntCollateral = snapshots[i].BurntCollateral.Add(snapshot.BurntCollateral)
		}

		return nil
	})

	return
}
