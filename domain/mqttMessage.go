package domain

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"sync"
)

type RegexpPatternType string

type Message struct {
	Lock        sync.Mutex
	MqttName    string
	TopicName   string
	PayloadName string
	MsgTemplate []byte
	Msg         map[string]interface{}
}
type CallBack interface {
	DataPointSet(client mqtt.Client, message mqtt.Message)
}
type IMqttMessageUsecase interface {
	GetTopicName() string
	GetMqttName() string
	GetPayloadName() string
	GetPublishMsg(isRealTime bool) ([]byte, error)
	GetCallBack() CallBack
}
