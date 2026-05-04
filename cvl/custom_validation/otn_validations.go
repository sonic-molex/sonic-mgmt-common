package custom_validation

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/Azure/sonic-mgmt-common/cvl/internal/util"
)

// ValidateOtnGain validates target-gain against per-amplifier bounds stored
// in CONFIG_DB OTN_OA|<name> (fields: min-gain, max-gain).
//
// If min-gain or max-gain is absent the check is skipped (permissive).
func (t *CustomValidation) ValidateOtnGain(vc *CustValidationCtxt) CVLErrorInfo {
	if vc.CurCfg.VOp == OP_DELETE {
		return CVLErrorInfo{ErrCode: CVL_SUCCESS}
	}

	// Key format: "OTN_OA|<name>"
	parts := strings.SplitN(vc.CurCfg.Key, "|", 2)
	if len(parts) != 2 {
		return CVLErrorInfo{ErrCode: CVL_SUCCESS}
	}
	ampName := parts[1]

	cfgDb := util.NewDbClient("CONFIG_DB")
	if cfgDb == nil {
		return CVLErrorInfo{
			ErrCode:       CVL_INTERNAL_UNKNOWN,
			CVLErrDetails: "CONFIG_DB connection failed",
		}
	}

	entryKey := "OTN_OA|" + ampName
	entry, err := cfgDb.HGetAll(entryKey).Result()
	if err != nil || len(entry) == 0 {
		return CVLErrorInfo{ErrCode: CVL_SUCCESS}
	}

	minGainStr, hasMin := entry["min-gain"]
	maxGainStr, hasMax := entry["max-gain"]
	if !hasMin || !hasMax {
		return CVLErrorInfo{ErrCode: CVL_SUCCESS}
	}

	minGain, errMin := strconv.ParseFloat(minGainStr, 64)
	maxGain, errMax := strconv.ParseFloat(maxGainStr, 64)
	if errMin != nil || errMax != nil {
		return CVLErrorInfo{ErrCode: CVL_SUCCESS}
	}

	val, errVal := strconv.ParseFloat(vc.YNodeVal, 64)
	if errVal != nil {
		return CVLErrorInfo{
			ErrCode:          CVL_SYNTAX_ERROR,
			TableName:        parts[0],
			Keys:             parts,
			Field:            vc.YNodeName,
			Value:            vc.YNodeVal,
			ConstraintErrMsg: "target-gain must be a valid decimal number",
		}
	}

	val = math.Round(val*100) / 100

	if val < minGain || val > maxGain {
		return CVLErrorInfo{
			ErrCode:          CVL_SEMANTIC_ERROR,
			TableName:        parts[0],
			Keys:             parts,
			Field:            vc.YNodeName,
			Value:            vc.YNodeVal,
			ConstraintErrMsg: fmt.Sprintf("target-gain %.2f out of configured range [%.2f, %.2f]", val, minGain, maxGain),
		}
	}

	return CVLErrorInfo{ErrCode: CVL_SUCCESS}
}
