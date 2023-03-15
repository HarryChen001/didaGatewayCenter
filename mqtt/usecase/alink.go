package usecase

import (
	"didaGatewayCenter"
	"didaGatewayCenter/mqtt/alinkSDK"
	"didaGatewayCenter/mqtt/alinkSDK/deviceInfo"
	"didaGatewayCenter/mqtt/alinkSDK/ota"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// upload device tag info (e.g.:sn,mac etc.) to aliyun server
func (n *NewMqtt) alinkDeviceInfoUpload() error {
	userName := n.mqttConfig.User
	productKey, deviceName, err := getProductKeyAndDeviceName(userName)
	if err != nil {
		return err
	}
	topic := fmt.Sprintf(alinkSDK.UploadDeviceInfoTopic, productKey, deviceName)
	d := deviceInfo.DeviceInfo{
		ID:      fmt.Sprintf("%d", time.Now().Unix()),
		Version: alinkSDK.Version,
		Params: []deviceInfo.Params{{
			AttrKey:   "SN",
			AttrValue: n.Parent.iSU.GetMachineInfoSn(),
		}, {
			AttrKey:   "MAC",
			AttrValue: n.Parent.iSU.GetMachineInfoSn(),
		}},
		Method: deviceInfo.Method,
	}
	p, _ := json.Marshal(d)
	token := n.client.Publish(topic, 0, false, p)
	return token.Error()
}

// upload ota info to aliyun server
func (n *NewMqtt) alinkOtaInfoUpload() error {
	userName := n.mqttConfig.User
	productKey, deviceName, err := getProductKeyAndDeviceName(userName)
	if err != nil {
		return err
	}
	topic := fmt.Sprintf(alinkSDK.UploadOtaInfoTopic, productKey, deviceName)
	otaInfo := ota.UploadInfo{
		Id: fmt.Sprintf("%d", time.Now().Unix()),
		Params: ota.Params{
			Version: didaGatewayCenter.Version,
			Module:  "main",
		},
	}
	payload, _ := json.Marshal(otaInfo)
	return n.client.Publish(topic, 0, false, payload).Error()
}

func getProductKeyAndDeviceName(username string) (productKey string, deviceName string, err error) {

	if !strings.Contains(username, "&") {
		return "", "", fmt.Errorf("username must contain '&',make sure this is aliyun broker")
	}
	alinkInfo := strings.Split(username, "&")
	deviceName = alinkInfo[0]
	productKey = alinkInfo[1]
	return productKey, deviceName, nil
}
