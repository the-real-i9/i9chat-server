package helpers

import (
	"encoding/json"
	"errors"
	"i9chat/appGlobals"
	"i9chat/appTypes"
)

func MapToStruct(val map[string]any, structData any) {
	bt, _ := json.Marshal(val)

	json.Unmarshal(bt, structData)
}

func ToStruct(val any, structData any) {
	bt, _ := json.Marshal(val)

	json.Unmarshal(bt, structData)
}

func ErrResp(code int, err error) appTypes.WSResp {
	if errors.Is(err, appGlobals.ErrInternalServerError) {
		return appTypes.WSResp{StatusCode: 500, Error: appGlobals.ErrInternalServerError.Error()}
	}

	return appTypes.WSResp{StatusCode: code, Error: err.Error()}
}
