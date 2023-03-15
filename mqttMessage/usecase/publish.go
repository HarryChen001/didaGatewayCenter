package usecase

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func (m *mqttMessageUsecase) GetPublishMsg(isRealTime bool) ([]byte, error) {

	m.message.Lock.Lock()
	defer m.message.Lock.Unlock()

	_ = json.Unmarshal(m.message.MsgTemplate, &m.message.Msg)

	m.generateMsgFormatObject(m.message.Msg, isRealTime)

	t, _ := json.Marshal(m.message.Msg)
	return t, nil
}
func (m *mqttMessageUsecase) generateMsgFormatObject(v map[string]interface{}, isRealTime bool) {
	for key, value := range v {
		switch value.(type) {
		case map[string]interface{}:
			m.generateMsgFormatObject(value.(map[string]interface{}), isRealTime)
		case []interface{}:
			m.generateMsgFormatArray(value.([]interface{}), isRealTime)
		case string:
			if strings.HasPrefix(value.(string), "${") {
				if strings.Contains(value.(string), "${timestampMs") {
					reg := regexp1[regexpPatternTimestampMs]
					aaa := reg.FindStringSubmatch(value.(string))
					valueType := aaa[1]
					switch valueType {
					case "int64":
						v[key] = time.Now().UnixMicro() / 1e3
					case "string":
						v[key] = fmt.Sprintf("%d", time.Now().UnixMicro()/1e3)
					}
				} else if strings.Contains(value.(string), "${timestampS") {
					reg := regexp1[regexpPatternTimestampS]
					aaa := reg.FindStringSubmatch(value.(string))
					valueType := aaa[1]
					switch valueType {
					case "int64":
						v[key] = time.Now().Unix()
					case "string":
						v[key] = fmt.Sprintf("%d", time.Now().Unix())
					}
				} else if strings.Contains(value.(string), "${variable}.") {
					reg := regexp1[regexpPatternVariable]
					aaa := reg.FindStringSubmatch(value.(string))
					variableName := aaa[1]
					id1, _ := strconv.ParseInt(variableName, 10, 64)
					variableValueType := aaa[2]
					variableValue, _ := m.iDPU.ReadById(id1, isRealTime)

					if variableValue == nil {
						v[key] = nil
					}
					switch variableValueType {
					case "float64":
						v[key] = variableValue
					case "string":
						v[key] = fmt.Sprintf("%f", variableValue)
					}
				}
			}
		}
	}
}
func (m *mqttMessageUsecase) generateMsgFormatArray(value []interface{}, isRealTime bool) {
	for _, v := range value {
		switch v.(type) {
		case []interface{}:
			m.generateMsgFormatArray(v.([]interface{}), isRealTime)
		case map[string]interface{}:
			m.generateMsgFormatObject(v.(map[string]interface{}), isRealTime)
		}
	}
}
