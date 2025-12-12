package transformer

import (
	"errors"
	"github.com/Azure/sonic-mgmt-common/translib/db"
	log "github.com/golang/glog"
	"strings"
)

func init() {
	XlateFuncBind("YangToDb_oc_name_key_xfmr", YangToDb_oc_name_key_xfmr)
	XlateFuncBind("DbToYang_oc_name_key_xfmr", DbToYang_oc_name_key_xfmr)
	XlateFuncBind("YangToDb_oc_name_field_xfmr", YangToDb_oc_name_field_xfmr)
	XlateFuncBind("DbToYang_oc_name_field_xfmr", DbToYang_oc_name_field_xfmr)
	XlateFuncBind("YangToDb_ocm_channel_key_xfmr", YangToDb_ocm_channel_key_xfmr)
	XlateFuncBind("DbToYang_ocm_channel_key_xfmr", DbToYang_ocm_channel_key_xfmr)
	XlateFuncBind("YangToDb_ocm_lower_frequency_xfmr", YangToDb_ocm_lower_frequency_xfmr)
	XlateFuncBind("YangToDb_osc_key_xfmr", YangToDb_osc_key_xfmr)
	XlateFuncBind("DbToYang_osc_key_xfmr", DbToYang_osc_key_xfmr)
	XlateFuncBind("YangToDb_osc_interface_xfmr", YangToDb_osc_interface_xfmr)
	XlateFuncBind("DbToYang_osc_interface_xfmr", DbToYang_osc_interface_xfmr)
	XlateFuncBind("otn_table_xfmr", otn_table_xfmr)
}

// Generic KeyXfmr for openconfig "name"
var YangToDb_oc_name_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
	if log.V(3) {
		log.Info("YangToDb_oc_key_xfmr: root: ", inParams.ygRoot,
			", uri: ", inParams.uri)
	}
	pathInfo := NewPathInfo(inParams.uri)
	ockey := pathInfo.Var("name")

	return ockey, nil
}

var DbToYang_oc_name_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	res_map := make(map[string]interface{}, 1)
	var err error

	if log.V(3) {
		log.Info("DbToYang_oc_key_xfmr: ", inParams.key)
	}

	res_map["name"] = inParams.key

	return res_map, err
}

var YangToDb_oc_name_field_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
	res_map := make(map[string]string)
	var err error
	res_map["NULL"] = "NULL"
	return res_map, err
}

var DbToYang_oc_name_field_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	var err error
	rmap := make(map[string]interface{})
	rmap["name"] = inParams.key

	return rmap, err
}

// OCM Channel KeyXfmrs
var YangToDb_ocm_channel_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
	if log.V(3) {
		log.Info("YangToDb_ocm_key_xfmr: root: ", inParams.ygRoot,
			", uri: ", inParams.uri)
	}
	pathInfo := NewPathInfo(inParams.uri)
	name := pathInfo.Var("name")
	lower := pathInfo.Var("lower-frequency")
	key := name + "|" + lower

	return key, nil
}

var DbToYang_ocm_channel_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	rmap := make(map[string]interface{})
	tableKeys := strings.Split(inParams.key, "|")

	if len(tableKeys) >= 2 {
		rmap["lower-frequency"] = tableKeys[1]
	}

	tableName := ""
	if strings.Contains(inParams.uri, "channels/channel") {
		tableName = "OTN_OCM_CHANNEL_TABLE"
	}

	upperFreq := ""

	// 1. Try Cache
	if inParams.dbDataMap != nil && tableName != "" {
		if tblData, ok := (*inParams.dbDataMap)[inParams.curDb][tableName]; ok {
			if row, ok2 := tblData[inParams.key]; ok2 {
				upperFreq = row.Field["upper-frequency"]
			}
		}
	}

	// 2. Direct Redis Fetch (fixes the undefined: db error and the type mismatch)
	if upperFreq == "" && tableName != "" {
		ts := &db.TableSpec{Name: tableName}
		// Using Comp for composite keys allows the driver to handle the '|' separator correctly
		rowKey := db.Key{Comp: strings.Split(inParams.key, "|")}

		entry, err := inParams.dbs[inParams.curDb].GetEntry(ts, rowKey)
		if err == nil {
			upperFreq = entry.Field["upper-frequency"]
		}
	}

	if upperFreq != "" {
		rmap["upper-frequency"] = upperFreq
	} else {
		log.Warningf("DbToYang_ocm_channel_key_xfmr: Could not resolve upper-frequency for %s", inParams.key)
	}

	return rmap, nil
}

var YangToDb_ocm_lower_frequency_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
	var err error
	rmap := make(map[string]string)

	rmap["NULL"] = "NULL"

	return rmap, err
}

// OSC Interface KeyXfmrs
var YangToDb_osc_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
	if log.V(3) {
		log.Info("YangToDb_osc_key_xfmr: root: ", inParams.ygRoot,
			", uri: ", inParams.uri)
	}
	pathInfo := NewPathInfo(inParams.uri)
	osckey := pathInfo.Var("interface")

	return osckey, nil
}

var DbToYang_osc_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	res_map := make(map[string]interface{}, 1)
	var err error

	if log.V(3) {
		log.Info("DbToYang_osc_key_xfmr: ", inParams.key)
	}

	res_map["interface"] = inParams.key

	return res_map, err
}

var YangToDb_osc_interface_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
	res_map := make(map[string]string)
	var err error
	return res_map, err
}

var DbToYang_osc_interface_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	var err error
	rmap := make(map[string]interface{})
	rmap["interface"] = inParams.key

	return rmap, err
}

var otn_table_xfmr TableXfmrFunc = func(inParams XfmrParams) ([]string, error) {

	// Check for nil
	if inParams.uri == "" {
		log.Error("otn_table_xfmr: uri is empty!")
		return nil, errors.New("otn_table_xfmr: uri is empty")
	}

	pathInfo := NewPathInfo(inParams.uri)
	targetUriPath := pathInfo.YangPath

	tblList := []string{}

	switch {
	case strings.HasPrefix(targetUriPath, "/openconfig-optical-attenuator:optical-attenuator/attenuators/attenuator"):
		tblList = append(tblList, "OTN_ATTENUATOR")

		// 1. Check the specific LIST first (Longer path)
	case strings.HasPrefix(targetUriPath, "/openconfig-channel-monitor:channel-monitors/channel-monitor/channels/channel"):
		tblList = append(tblList, "OTN_OCM_CHANNEL_TABLE")

	// 2. Check the CONTAINER next (Shorter path)
	case strings.HasPrefix(targetUriPath, "/openconfig-channel-monitor:channel-monitors/channel-monitor/channels"):
		// Returning empty here forces the framework to recurse
		// down to the 'channel' list where it will find the table above.
		return tblList, nil

	// 3. Check the PARENT MONITOR
	case strings.HasPrefix(targetUriPath, "/openconfig-channel-monitor:channel-monitors/channel-monitor"):
		tblList = append(tblList, "OTN_OCM")

	case strings.HasPrefix(targetUriPath, "/openconfig-optical-amplifier:optical-amplifier/amplifiers/amplifier"):
		tblList = append(tblList, "OTN_OA")

	case strings.HasPrefix(targetUriPath, "/openconfig-optical-amplifier:optical-amplifier/supervisory-channels/supervisory-channel"):
		tblList = append(tblList, "OTN_OSC")
	}

	if len(tblList) == 0 {
		log.Errorf("otn_table_xfmr: NO MATCHING TABLE for yangPath=%s", targetUriPath)
		return nil, errors.New("otn_table_xfmr: no matching table")
	}

	return tblList, nil
}
