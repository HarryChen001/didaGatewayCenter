package usecase

import (
	"didaGatewayCenter/domain"
	"encoding/json"
	"errors"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"os"
	"path"
	"regexp"
	"sync"
)

const (
	regexpPatternTimestampMs domain.RegexpPatternType = `^\${timestampMs}\.(\w+)$`
	regexpPatternTimestampS  domain.RegexpPatternType = `^\${timestampS}\.(\w+)$`
	regexpPatternVariable    domain.RegexpPatternType = `^\$\{variable}\.\$\{(.*)}\.(\w+)$`
)

var regexp1 map[domain.RegexpPatternType]*regexp.Regexp

type mqttMessageUsecase struct {
	iDPU    domain.IDataPointUseCase
	message domain.Message
}

func (m *mqttMessageUsecase) GetTopicName() string {
	return m.message.TopicName
}

func NewMqttMessageUsecase(mqttName string, topicName string, payloadName string, iACU domain.IAppConfigUseCase, iDPU domain.IDataPointUseCase) (domain.IMqttMessageUsecase, error) {

	if regexp1 == nil {
		regexp1 = make(map[domain.RegexpPatternType]*regexp.Regexp)

		r := regexp.MustCompile(string(regexpPatternTimestampMs))
		regexp1[regexpPatternTimestampMs] = r
		r1 := regexp.MustCompile(string(regexpPatternTimestampS))
		regexp1[regexpPatternTimestampS] = r1
		r2 := regexp.MustCompile(string(regexpPatternVariable))
		regexp1[regexpPatternVariable] = r2

	}

	p := mqttMessageUsecase{
		message: domain.Message{
			Lock:        sync.Mutex{},
			MqttName:    mqttName,
			TopicName:   topicName,
			PayloadName: payloadName,
		},
		iDPU: iDPU,
	}

	dir := iACU.GetAppMqttConfig().MessageConfig.Dir
	fileLocation := path.Join(dir, mqttName, payloadName)
	fileInfo, err := os.ReadFile(fileLocation)
	if err != nil {
		return nil, err
	}
	if ok := json.Valid(fileInfo); !ok {
		return nil, errors.New("the message is not json format")
	}
	p.message.MsgTemplate = fileInfo
	return &p, nil
}

func (m *mqttMessageUsecase) GetMqttName() string {
	return m.message.MqttName
}

func (m *mqttMessageUsecase) GetPayloadName() string {
	return m.message.PayloadName
}

func (m *mqttMessageUsecase) GetCallBack() domain.CallBack {
	return m
}

func (m *mqttMessageUsecase) DataPointSet(client mqtt.Client, message mqtt.Message) {
	m.message.Lock.Lock()
	defer func() {
		m.message.Lock.Unlock()
	}()
	msgPayload := message.Payload()
	msgTemplate := m.message.MsgTemplate
	template := make(map[string]interface{})
	payloadMap := make(map[string]interface{})
	json.Unmarshal(msgTemplate, &template)
	json.Unmarshal(msgPayload, &payloadMap)
	m.parseReceive(template, payloadMap)
}
