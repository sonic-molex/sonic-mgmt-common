package transformer

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Azure/sonic-mgmt-common/translib/db"
	"github.com/Azure/sonic-mgmt-common/translib/tlerr"
	log "github.com/golang/glog"
)

const (
	ocmNotificationChannel = "OTN_OCM_NOTIFICATION"
	ocmReplyChannel        = "OTN_OCM_REPLY"
	ocmRpcTimeout          = 10 // seconds
)

func init() {
	XlateFuncBind("act_get_ocm_raw_cb", act_get_ocm_raw_cb)
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

// act_get_ocm_raw_cb handles the get-ocm-raw action.
//
// Flow:
//  1. Get "port" from URI path variables (channel-monitor list key "name").
//  2. Subscribe to OTN_OCM_REPLY before publishing, to avoid missing the reply.
//  3. Publish ["get-ocm-raw", "<port>", []] to OTN_OCM_NOTIFICATION.
//  4. Wait for orchagent reply on OTN_OCM_REPLY.
//  5. Decode reply and return as openconfig-channel-monitor output JSON.
var act_get_ocm_raw_cb ActionCallpoint = func(vars map[string]string, body []byte, dbs [db.MaxDB]*db.DB) ([]byte, error) {

	// --- 1. Get port from URI path variables (channel-monitor list key) ---
	port := vars["name"]
	if port == "" {
		return nil, tlerr.New("missing channel-monitor name in URI")
	}
	log.V(3).Infof("act_get_ocm_raw_cb: port=%s", port)

	// --- 2. Subscribe to reply channel BEFORE publishing request ---
	replyDB, err := db.PubSubRpcDB(db.Options{DBNo: db.ApplDB}, ocmReplyChannel)
	if err != nil {
		log.Errorf("act_get_ocm_raw_cb: PubSubRpcDB failed: %v", err)
		return nil, tlerr.New("failed to subscribe to OCM reply channel")
	}
	defer replyDB.ClosePubSubRpcDB()

	// --- 3. Build and publish the swsscommon notification message ---
	msg, err := buildSwssNotification("get-ocm-raw", port, nil)
	if err != nil {
		return nil, tlerr.New("failed to build notification message")
	}

	listeners, err := replyDB.SendRpcRequest(ocmNotificationChannel, msg)
	if err != nil {
		log.Errorf("act_get_ocm_raw_cb: SendRpcRequest failed: %v", err)
		return nil, tlerr.New("failed to send OCM RPC request")
	}
	if listeners == 0 {
		log.Warningf("act_get_ocm_raw_cb: no listeners on %s", ocmNotificationChannel)
		return nil, tlerr.New("no orchagent listener on OCM notification channel")
	}

	// --- 4. Wait for reply ---
	replies, err := replyDB.GetRpcResponse(1, ocmRpcTimeout)
	if err != nil {
		log.Errorf("act_get_ocm_raw_cb: GetRpcResponse failed: %v", err)
		return nil, tlerr.New("OCM RPC timed out or failed")
	}
	if len(replies) == 0 {
		return nil, tlerr.New("OCM RPC: empty reply")
	}

	// --- 5. Decode swsscommon reply and build output ---
	op, _, fvs, err := parseSwssNotification(replies[0])
	if err != nil {
		log.Errorf("act_get_ocm_raw_cb: failed to parse reply: %v", err)
		return nil, tlerr.New("failed to parse OCM reply")
	}

	// Output struct mirrors proto GetOcmRawResponse.Output JSON shape
	var resp struct {
		Output struct {
			Length        uint32 `json:"length"`
			Data          string `json:"data"`
			Status        string `json:"status"`
			StatusMessage string `json:"status-message"`
		} `json:"sonic-oc-action-ext:output"`
	}

	resp.Output.Status = "Successful"
	if op != "SUCCESS" {
		resp.Output.Status = "Failed"
		resp.Output.StatusMessage = fmt.Sprintf("orchagent returned: %s", op)
	}

	for _, fv := range fvs {
		switch fv[0] {
		case "count":
			if n, err := strconv.ParseUint(fv[1], 10, 32); err == nil {
				resp.Output.Length = uint32(n)
			}
		case "data":
			resp.Output.Data = fv[1]
		}
	}

	result, err := json.Marshal(&resp)
	if err != nil {
		return nil, tlerr.New("failed to marshal OCM output")
	}
	return result, nil
}

// buildSwssNotification encodes a swsscommon NotificationProducer message.
// swsscommon flat format: ["<op>", "<data>", "field1", "val1", ...]
func buildSwssNotification(op, data string, fvs [][2]string) (string, error) {
	msg := []interface{}{op, data}
	for _, fv := range fvs {
		msg = append(msg, fv[0], fv[1])
	}
	b, err := json.Marshal(msg)
	return string(b), err
}

// parseSwssNotification decodes a swsscommon NotificationProducer reply.
// swsscommon flat format: ["<op>", "<data>", "field1", "val1", ...]
func parseSwssNotification(raw string) (op, data string, fvs [][2]string, err error) {
	var arr []json.RawMessage
	if err = json.Unmarshal([]byte(raw), &arr); err != nil {
		return
	}
	if len(arr) < 2 {
		err = fmt.Errorf("notification has %d elements, want at least 2", len(arr))
		return
	}
	if err = json.Unmarshal(arr[0], &op); err != nil {
		return
	}
	if err = json.Unmarshal(arr[1], &data); err != nil {
		return
	}
	for i := 2; i+1 < len(arr); i += 2 {
		var key, val string
		if err = json.Unmarshal(arr[i], &key); err != nil {
			return
		}
		if err = json.Unmarshal(arr[i+1], &val); err != nil {
			return
		}
		fvs = append(fvs, [2]string{key, val})
	}
	return
}
