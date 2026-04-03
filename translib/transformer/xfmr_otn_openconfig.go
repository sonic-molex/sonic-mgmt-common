package transformer

import (
	"errors"
	"strconv"
	"strings"

	log "github.com/golang/glog"
)

func init() {
	XlateFuncBind("YangToDb_oc_name_key_xfmr", YangToDb_oc_name_key_xfmr)
	XlateFuncBind("DbToYang_oc_name_key_xfmr", DbToYang_oc_name_key_xfmr)
	XlateFuncBind("YangToDb_oc_name_field_xfmr", YangToDb_oc_name_field_xfmr)
	XlateFuncBind("DbToYang_oc_name_field_xfmr", DbToYang_oc_name_field_xfmr)
	XlateFuncBind("YangToDb_ocm_channel_key_xfmr", YangToDb_ocm_channel_key_xfmr)
	XlateFuncBind("DbToYang_ocm_channel_key_xfmr", DbToYang_ocm_channel_key_xfmr)
	XlateFuncBind("YangToDb_ocm_channel_lower_frequency_xfmr", YangToDb_ocm_channel_lower_frequency_xfmr)
	XlateFuncBind("DbToYang_ocm_channel_lower_frequency_xfmr", DbToYang_ocm_channel_lower_frequency_xfmr)
	XlateFuncBind("YangToDb_ocm_channel_upper_frequency_xfmr", YangToDb_ocm_channel_upper_frequency_xfmr)
	XlateFuncBind("DbToYang_ocm_channel_upper_frequency_xfmr", DbToYang_ocm_channel_upper_frequency_xfmr)
	XlateFuncBind("YangToDb_osc_key_xfmr", YangToDb_osc_key_xfmr)
	XlateFuncBind("DbToYang_osc_key_xfmr", DbToYang_osc_key_xfmr)
	XlateFuncBind("YangToDb_osc_interface_xfmr", YangToDb_osc_interface_xfmr)
	XlateFuncBind("DbToYang_osc_interface_xfmr", DbToYang_osc_interface_xfmr)
	XlateFuncBind("otn_table_xfmr", otn_table_xfmr)
	// OTN WSS / wavelength-router transformers
	XlateFuncBind("YangToDb_wss_index_key_xfmr", YangToDb_wss_index_key_xfmr)
	XlateFuncBind("DbToYang_wss_index_key_xfmr", DbToYang_wss_index_key_xfmr)
	XlateFuncBind("YangToDb_wss_index_field_xfmr", YangToDb_wss_index_field_xfmr)
	XlateFuncBind("DbToYang_wss_index_field_xfmr", DbToYang_wss_index_field_xfmr)
	XlateFuncBind("YangToDb_wss_power_profile_key_xfmr", YangToDb_wss_power_profile_key_xfmr)
	XlateFuncBind("DbToYang_wss_power_profile_key_xfmr", DbToYang_wss_power_profile_key_xfmr)
	XlateFuncBind("YangToDb_wss_power_profile_upper_frequency_xfmr", YangToDb_wss_power_profile_upper_frequency_xfmr)
	XlateFuncBind("DbToYang_wss_power_profile_upper_frequency_xfmr", DbToYang_wss_power_profile_upper_frequency_xfmr)
	XlateFuncBind("YangToDb_wss_power_profile_lower_frequency_xfmr", YangToDb_wss_power_profile_lower_frequency_xfmr)
	XlateFuncBind("DbToYang_wss_power_profile_lower_frequency_xfmr", DbToYang_wss_power_profile_lower_frequency_xfmr)
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
	if name == "*" {
		return name, nil
	}

	lower := pathInfo.Var("lower-frequency")
	upper := pathInfo.Var("upper-frequency")
	if lower == "*" || lower == "" || upper == "*" || upper == "" {
		// Return empty so the framework enumerates all keys in the table.
		// Partial wildcards like "name|*" are treated as literal keys by
		// the framework and fail; only "" or "*" trigger key enumeration.
		return "", nil
	}

	idx := name + "|" + lower + "|" + upper

	log.Info("YangToDb_ocm_channel_key_xfmr - return idx ", idx)
	return idx, nil
}

var DbToYang_ocm_channel_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	rmap := make(map[string]interface{})
	tableKeys := strings.Split(inParams.key, "|")

	if len(tableKeys) >= 2 {
		rmap["lower-frequency"] = tableKeys[1]
	}
	if len(tableKeys) >= 3 {
		rmap["upper-frequency"] = tableKeys[2]
	}

	log.Info("DbToYang_ocm_channel_key_xfmr rmap ", rmap)
	return rmap, nil
}

var YangToDb_ocm_channel_lower_frequency_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
	var err error
	rmap := make(map[string]string)
	rmap["NULL"] = "NULL"
	return rmap, err
}

var DbToYang_ocm_channel_lower_frequency_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	rmap := make(map[string]interface{})
	parts := strings.Split(inParams.key, "|")
	if len(parts) >= 2 {
		rmap["lower-frequency"] = parts[1]
	}
	return rmap, nil
}

var YangToDb_ocm_channel_upper_frequency_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
	var err error
	rmap := make(map[string]string)
	rmap["NULL"] = "NULL"
	return rmap, err
}

var DbToYang_ocm_channel_upper_frequency_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	rmap := make(map[string]interface{})
	parts := strings.Split(inParams.key, "|")
	if len(parts) >= 3 {
		rmap["upper-frequency"] = parts[2]
	}
	return rmap, nil
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

	case strings.HasPrefix(targetUriPath,
		"/openconfig-wavelength-router:wavelength-router/media-channels/channel/spectrum-power-profile/distribution"):
		tblList = append(tblList, "OTN_WSS_SPEC_POWER")

	case strings.HasPrefix(targetUriPath,
		"/openconfig-wavelength-router:wavelength-router/media-channels/channel/spectrum-power-profile"):
		tblList = append(tblList, "OTN_WSS_SPEC_POWER")

	case strings.HasPrefix(targetUriPath,
		"/openconfig-wavelength-router:wavelength-router/media-channels/channel"):
		tblList = append(tblList, "OTN_WSS")

	case strings.HasPrefix(targetUriPath,
		"/openconfig-wavelength-router:wavelength-router/media-channels"):
		return tblList, nil

	case strings.HasPrefix(targetUriPath,
		"/openconfig-wavelength-router:wavelength-router"):
		return tblList, nil

	}

	if len(tblList) == 0 {
		log.Errorf("otn_table_xfmr: NO MATCHING TABLE for yangPath=%s", targetUriPath)
		return nil, errors.New("otn_table_xfmr: no matching table")
	}

	return tblList, nil
}

// OTN WSS / wavelength-router: channel keyed by index
var YangToDb_wss_index_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
	if log.V(3) {
		log.Info("YangToDb_oc_index_key_xfmr: root: ", inParams.ygRoot,
			", uri: ", inParams.uri)
	}
	pathInfo := NewPathInfo(inParams.uri)
	ockey := pathInfo.Var("index")
	return ockey, nil
}

var DbToYang_wss_index_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	res := make(map[string]interface{}, 1)
	if log.V(3) {
		log.Info("DbToYang_oc_index_key_xfmr: ", inParams.key)
	}
	// Accept either "index" or "index|something"
	k := inParams.key
	if strings.Contains(k, "|") {
		k = strings.SplitN(k, "|", 2)[0]
	}

	v, err := strconv.ParseUint(k, 10, 32)
	if err != nil {
		return res, err
	}
	res["index"] = uint32(v)
	return res, nil
}

// Index is the table key; do not store as separate field in DB
var YangToDb_wss_index_field_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
	resMap := make(map[string]string)
	resMap["NULL"] = "NULL"
	return resMap, nil
}

var DbToYang_wss_index_field_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	rmap := make(map[string]interface{})

	v, err := strconv.ParseUint(inParams.key, 10, 32)
	if err != nil {
		return rmap, err
	}
	rmap["index"] = uint32(v)

	return rmap, nil
}

// OTN WSS spectrum-power-profile distribution: CONFIG_DB key is index|lower-frequency|upper-frequency.
// YANG list distribution is keyed by (lower-frequency, upper-frequency).
var YangToDb_wss_power_profile_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
	if log.V(3) {
		log.Info("YangToDb_oc_power_profile_key_xfmr: root: ", inParams.ygRoot,
			", uri: ", inParams.uri)
	}
	pathInfo := NewPathInfo(inParams.uri)
	channelIndex := pathInfo.Var("index")
	lower := pathInfo.Var("lower-frequency")
	upper := pathInfo.Var("upper-frequency")
	if lower == "*" || lower == "" || upper == "*" || upper == "" {
		// For list GET on distribution (no both keys in URI), keep the query
		// constrained to the parent channel key instead of scanning all channels.
		if channelIndex != "" && channelIndex != "*" {
			return channelIndex + "|*", nil
		}
		return "*", nil
	}
	key := channelIndex + "|" + lower + "|" + upper
	return key, nil
}

var DbToYang_wss_power_profile_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	rmap := make(map[string]interface{})

	// DB key is "index|lower-frequency|upper-frequency"
	parts := strings.Split(inParams.key, "|")
	if len(parts) >= 2 {
		rmap["lower-frequency"] = parts[1]
	}
	if len(parts) >= 3 {
		rmap["upper-frequency"] = parts[2]
	}

	return rmap, nil
}

var YangToDb_wss_power_profile_upper_frequency_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
	rmap := make(map[string]string)
	rmap["NULL"] = "NULL"
	return rmap, nil
}

var DbToYang_wss_power_profile_upper_frequency_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	rmap := make(map[string]interface{})
	parts := strings.Split(inParams.key, "|")
	if len(parts) >= 3 {
		rmap["upper-frequency"] = parts[2]
	}
	return rmap, nil
}

var YangToDb_wss_power_profile_lower_frequency_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
	rmap := make(map[string]string)
	rmap["NULL"] = "NULL"
	return rmap, nil
}

var DbToYang_wss_power_profile_lower_frequency_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
	rmap := make(map[string]interface{})
	parts := strings.Split(inParams.key, "|")
	if len(parts) >= 2 {
		rmap["lower-frequency"] = parts[1]
	}
	return rmap, nil
}
