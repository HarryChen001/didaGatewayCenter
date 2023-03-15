package usecase

import (
	"crypto/md5"
	"didaGatewayCenter/domain"
	"encoding/hex"
	"strings"
)

func (m *Mqtt) addDidaMeter(useCase domain.ISystemUseCase) {

	sn := useCase.GetMachineInfoSn()
	tempMac := useCase.GetMachineInfoMac()
	server := useCase.GetMachineInfoServer()
	port := useCase.GetMachineInfoPort()
	intervalS := useCase.GetMachineInfoUploadIntervalS()
	enabled := useCase.GetMachineInfoEnabled()

	if intervalS == 0 {
		intervalS = 15
	}
	mac := strings.Replace(tempMac, ":", "", -1)
	username := sn + "&" + mac
	clientId := sn + "&" + mac + "&" + "WG40"
	md5New := md5.New()
	md5New.Write([]byte(clientId))
	password := hex.EncodeToString(md5New.Sum(nil)) + username

	tempMqtt := &domain.MQTTConfig{
		MQTTName:     "didaLink",
		Valid:        enabled,
		ServerType:   domain.ServerTypeDidaLinkType,
		Server:       server,
		Port:         port,
		User:         username,
		PW:           password,
		ClientID:     clientId,
		CleanSession: true,
		SSL:          false,
		KeepAlive:    60,
		PubTopics: []domain.PubTopicStruct{{
			Topic:       "/mq/" + clientId + "/property/post",
			QoS:         0,
			Type:        domain.PTopicTypeUpload,
			PayloadType: domain.PayloadTypeDidaLink,
			UpIntervalS: intervalS,
			Valid:       true,
		}},
		SubTopics: []domain.SubTopicStruct{{
			Topic:       "/mq/" + clientId + "/property/set",
			QoS:         0,
			Type:        domain.STopicTypeReceive,
			PayloadType: domain.PayloadTypeDidaLink,
			Valid:       true,
		}, {
			Topic:       "/mq/" + clientId + "/config/get",
			QoS:         0,
			Type:        domain.STopicTypeRemote,
			PayloadType: domain.PayloadTypeAlink,
			Valid:       true,
		}},
		CAFile:      "",
		CertFile:    "",
		KeyFile:     "",
		NetworkType: 5000,
	}
	m.nMqtt = append(m.nMqtt, &NewMqtt{
		Parent:     m,
		mqttConfig: tempMqtt,
		client:     nil,
	})
}
