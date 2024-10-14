package logging_json_bodies

import (
	"bytes"
	"encoding/json"
	"io"
	"maps"
	"net/http"
	"strings"
)

func LimitRequestBodySize(r *http.Request) []byte {
	return ReadRequestBodyWithoutClosingWithCustomLimit(r, 1024)
}

func ReadRequestBodyWithoutClosingWithCustomLimit(r *http.Request, maxBytes int64) []byte {
	var bodyBuffer bytes.Buffer
	io.Copy(&bodyBuffer, r.Body)
	r.Body.Close()
	r.Body = io.NopCloser(io.LimitReader(bytes.NewReader(bodyBuffer.Bytes()), maxBytes))
	return bodyBuffer.Bytes()
}

func GetJsonBodyWithFieldsRemoved(jsonBody []byte, maxFirstBytesToLeave int, fieldNames ...string) string {
	var jsonReqBody string
	var data map[string]interface{}

	if jsonBody != nil {
		json.Unmarshal(jsonBody, &data)
		for _, field := range fieldNames {
			data = findAndChangeFields(data, maxFirstBytesToLeave, field)
		}
		jsonBody, _ = json.Marshal(&data)
		jsonReqBody = string(jsonBody)
	}

	return jsonReqBody
}

func findAndChangeFields(data map[string]interface{}, maxFirstBytesToLeave int, fieldName string) map[string]interface{} {
	if strings.Contains(fieldName, ".") {
		fieldParts := strings.SplitN(fieldName, ".", 2)
		firstLevelFieldName := fieldParts[0]
		secondLevelFieldName := fieldParts[1]
		secondLevelData := data[firstLevelFieldName]
		if secondLevelData != nil {
			changedData := findAndChangeFields(secondLevelData.(map[string]interface{}), maxFirstBytesToLeave, secondLevelFieldName)
			data[firstLevelFieldName] = changedData
		}
	} else {
		data = changeFieldValue(fieldName, data, maxFirstBytesToLeave)
	}
	return data
}

func changeFieldValue(field string, data map[string]interface{}, maxFirstBytesToLeave int) (changedData map[string]interface{}) {
	changedData = maps.Clone(data)

	if maxFirstBytesToLeave > 0 && data[field] != nil {
		var firstBytesToLeave = len(data[field].(string))
		if firstBytesToLeave > maxFirstBytesToLeave {
			firstBytesToLeave = maxFirstBytesToLeave
		}
		changedData[field] = data[field].(string)[0:firstBytesToLeave] + "... "
	} else {
		changedData[field] = ""
	}
	changedData[field] = changedData[field].(string) + "[HIDDEN]"

	return changedData
}
