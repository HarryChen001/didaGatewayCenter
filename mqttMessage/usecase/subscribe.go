package usecase

import (
	"regexp"
	"strconv"
	"strings"
)

func (m *mqttMessageUsecase) parseReceive(template map[string]interface{}, payload map[string]interface{}) {

	for key, value := range template {
		if _, ok := payload[key]; !ok {
			continue
		}

		switch value.(type) {
		case map[string]interface{}:
			switch payload[key].(type) {
			case map[string]interface{}:
			default:
				return
			}
			m.parseReceive(value.(map[string]interface{}), payload[key].(map[string]interface{}))
		case []interface{}:
		case string:
			tempValue := value.(string)
			if strings.HasPrefix(tempValue, "${") {
				if strings.Contains(tempValue, "${variable}.") {
					variableValue := payload[key].(float64)
					idString := regexp.MustCompile(`^\$\{variable}\.\$\{(.*)}\.(\w+)$`).FindStringSubmatch(tempValue)[1]
					id, _ := strconv.ParseInt(idString, 10, 64)
					_, err := m.iDPU.WriteById(id, variableValue)
					// TODO: error handling
					_ = err
				}
			}
		}
	}
	return
}
