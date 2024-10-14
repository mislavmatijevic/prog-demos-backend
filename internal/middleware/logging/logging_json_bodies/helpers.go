package logging_json_bodies

import (
	"bytes"
	"encoding/json"
	"io"
	"maps"
	"net/http"
	"strings"
)

func ReadRequestBodyWithoutClosing(r *http.Request) []byte {
	var bodyBuffer bytes.Buffer
	io.Copy(&bodyBuffer, r.Body)
	r.Body.Close()
	r.Body = io.NopCloser(io.LimitReader(bytes.NewReader(bodyBuffer.Bytes()), 1024))
	return bodyBuffer.Bytes()
}

func HideFieldsFromJsonBody(jsonBody []byte, shouldLeaveSomeFirstBytes bool, fieldNames ...string) string {
	var jsonReqBody string
	var data map[string]interface{}

	if jsonBody != nil {
		json.Unmarshal(jsonBody, &data)
		for _, field := range fieldNames {
			data = findAndChangeFields(data, shouldLeaveSomeFirstBytes, field)
		}
		jsonBody, _ = json.Marshal(&data)
		jsonReqBody = string(jsonBody)
	}

	return jsonReqBody
}

func findAndChangeFields(data map[string]interface{}, shouldLeaveSomeFirstBytes bool, fieldName string) map[string]interface{} {
	if strings.Contains(fieldName, ".") {
		fieldParts := strings.SplitN(fieldName, ".", 2)
		firstLevelFieldName := fieldParts[0]
		secondLevelFieldName := fieldParts[1]
		secondLevelData := data[firstLevelFieldName]
		if secondLevelData != nil {
			changedData := findAndChangeFields(secondLevelData.(map[string]interface{}), shouldLeaveSomeFirstBytes, secondLevelFieldName)
			data[firstLevelFieldName] = changedData
		}
	} else {
		data = changeFieldValue(fieldName, data, shouldLeaveSomeFirstBytes)
	}
	return data
}

func changeFieldValue(field string, data map[string]interface{}, shouldLeaveSomeFirstBytes bool) (changedData map[string]interface{}) {
	changedData = maps.Clone(data)

	if shouldLeaveSomeFirstBytes && data[field] != nil {
		var firstBytesToLeave = len(data[field].(string))
		if firstBytesToLeave > 10 {
			firstBytesToLeave = 10
		}
		changedData[field] = data[field].(string)[0:firstBytesToLeave] + "... [HIDDEN]"
	} else {
		delete(changedData, field)
	}

	return changedData
}
