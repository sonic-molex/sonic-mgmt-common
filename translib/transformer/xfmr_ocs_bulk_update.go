package transformer

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Azure/sonic-mgmt-common/translib/db"
	"github.com/Azure/sonic-mgmt-common/translib/tlerr"
	"github.com/golang/glog"
)

func init() {
	XlateFuncBind("rpc_ocs_bulk_update_cb", rpc_ocs_bulk_update_cb)
}

// Input JSON shape:
//
//	{
//	  "sonic-ocs:input": {
//	    "delete": ["1A-68B", "2A-67B"],
//	    "add": [
//	      {"cross_connect_id":"3A-66B", "a_side":"3A", "b_side":"66B"}
//	    ]
//	  }
//	}
type ocsBulkUpdateInput struct {
	Input struct {
		Delete []string          `json:"delete"`
		Add    []ocsCrossConnect `json:"add"`
	} `json:"sonic-ocs:input"`
}

type ocsCrossConnect struct {
	CrossConnectID string `json:"cross_connect_id"`
	ASide          string `json:"a_side"`
	BSide          string `json:"b_side"`
}

type ocsBulkUpdateOutput struct {
	Output struct {
		DeletedCount uint32   `json:"deleted_count"`
		AddedCount   uint32   `json:"added_count"`
		FailedKeys   []string `json:"failed_keys,omitempty"`
	} `json:"sonic-ocs:output"`
}

const ocsXConnectTable = "OCS_CROSS_CONNECT"

var rpc_ocs_bulk_update_cb RpcCallpoint = func(body []byte, dbs [db.MaxDB]*db.DB) ([]byte, error) {
	tRecv := time.Now().UnixMilli()
	glog.Infof("bulk-update: received, body length=%d, t=%d", len(body), tRecv)

	var input ocsBulkUpdateInput
	if err := json.Unmarshal(body, &input); err != nil {
		glog.Errorf("rpc_ocs_bulk_update_cb: failed to parse input: %v", err)
		return nil, tlerr.InvalidArgs("Invalid rpc input: %v", err)
	}

	deleteKeys := input.Input.Delete
	addEntries := input.Input.Add

	glog.Infof("rpc_ocs_bulk_update_cb: %d delete, %d add", len(deleteKeys), len(addEntries))

	if len(deleteKeys) == 0 && len(addEntries) == 0 {
		glog.Info("rpc_ocs_bulk_update_cb: empty request, returning no-op")
		return marshalOutput(0, 0, nil)
	}

	// Use the ConfigDB connection provided by Action()/getAllDbs().
	// Do NOT open a new connection — Action() already holds the ConfigDBLock
	// via dbs[db.ConfigDB], so a second NewDB(ConfigDB) would deadlock.
	cfgDb := dbs[db.ConfigDB]
	if cfgDb == nil {
		glog.Errorf("rpc_ocs_bulk_update_cb: dbs[ConfigDB] is nil")
		return nil, tlerr.InternalError{Format: "CONFIG_DB handle not available", Path: "bulk-update"}
	}

	ts := &db.TableSpec{Name: ocsXConnectTable}

	// Start a Redis MULTI/EXEC transaction for atomicity.
	if err := cfgDb.StartTx(nil, nil); err != nil {
		glog.Errorf("rpc_ocs_bulk_update_cb: failed to start transaction: %v", err)
		return nil, tlerr.InternalError{Format: "Failed to start CONFIG_DB transaction", Path: "bulk-update"}
	}

	var failedKeys []string
	var deletedCount uint32
	var addedCount uint32

	// Phase 1: deletes (before adds to free ports)
	for _, keyStr := range deleteKeys {
		key := db.Key{Comp: []string{keyStr}}
		if glog.V(3) {
			glog.Infof("rpc_ocs_bulk_update_cb: deleting key=%s", keyStr)
		}
		if err := cfgDb.DeleteEntry(ts, key); err != nil {
			glog.Warningf("rpc_ocs_bulk_update_cb: delete failed for key=%s: %v", keyStr, err)
			failedKeys = append(failedKeys, keyStr)
		} else {
			deletedCount++
		}
	}

	// Phase 2: adds
	for _, entry := range addEntries {
		if entry.CrossConnectID == "" {
			glog.Warning("rpc_ocs_bulk_update_cb: skipping add entry with empty cross_connect_id")
			failedKeys = append(failedKeys, "")
			continue
		}
		key := db.Key{Comp: []string{entry.CrossConnectID}}
		value := db.Value{Field: map[string]string{
			"a_side": entry.ASide,
			"b_side": entry.BSide,
		}}
		if glog.V(3) {
			glog.Infof("rpc_ocs_bulk_update_cb: adding key=%s a_side=%s b_side=%s",
				entry.CrossConnectID, entry.ASide, entry.BSide)
		}
		if err := cfgDb.SetEntry(ts, key, value); err != nil {
			glog.Warningf("rpc_ocs_bulk_update_cb: add failed for key=%s: %v", entry.CrossConnectID, err)
			failedKeys = append(failedKeys, entry.CrossConnectID)
		} else {
			addedCount++
		}
	}

	// Commit the transaction
	if err := cfgDb.CommitTx(); err != nil {
		glog.Errorf("rpc_ocs_bulk_update_cb: transaction commit failed: %v", err)
		cfgDb.AbortTx()
		return nil, tlerr.InternalError{
			Format: fmt.Sprintf("CONFIG_DB transaction commit failed: %v", err),
			Path:   "bulk-update",
		}
	}

	tCommit := time.Now().UnixMilli()
	glog.Infof("bulk-update: CONFIG_DB committed, %d delete, %d add, %d failed, t=%d, elapsed=%dms",
		deletedCount, addedCount, len(failedKeys), tCommit, tCommit-tRecv)

	return marshalOutput(deletedCount, addedCount, failedKeys)
}

func marshalOutput(deleted, added uint32, failed []string) ([]byte, error) {
	var out ocsBulkUpdateOutput
	out.Output.DeletedCount = deleted
	out.Output.AddedCount = added
	if len(failed) > 0 {
		out.Output.FailedKeys = failed
	}
	result, err := json.Marshal(&out)
	if err != nil {
		glog.Errorf("rpc_ocs_bulk_update_cb: failed to marshal output: %v", err)
		return nil, tlerr.InternalError{Format: "Failed to marshal output", Path: "bulk-update"}
	}
	return result, nil
}
