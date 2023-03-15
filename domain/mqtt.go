package domain

import mqtt "github.com/eclipse/paho.mqtt.golang"

type ServerType int
type PayloadType int
type PTopicType int
type STopicType int

const (
	ServerTypeNormalType      ServerType = 1001
	ServerTypeAlinkType       ServerType = 1002
	ServerTypeOneNetType      ServerType = 1003
	ServerTypeThingsBoardType ServerType = 1004
	ServerTypeDidaLinkType    ServerType = 1005
)

const (
	PayloadTypeCustom       PayloadType = 2001
	PayloadTypeAlink        PayloadType = 2002
	PayloadTypeSerialUpload PayloadType = 2003
	PayloadTypeAlinkRRPC    PayloadType = 2004
	PayloadTypeFujian       PayloadType = 2005
	PayloadTypeOneNet       PayloadType = 2006
	PayloadTypeThingsBoard  PayloadType = 2007
	PayloadTypeDidaLink     PayloadType = 2008
)
const (
	PTopicTypeUpload               PTopicType = 4001
	PTopicTypeSerialUpload         PTopicType = 4005
	PTopicTypeAlinkPropertyPost    PTopicType = 4006
	PTopicTypeAlinkEventPost       PTopicType = 4008
	PTopicTypeHistoryPost          PTopicType = 4011
	PTopicTypeRemoteUpdateResponse PTopicType = 4014
)
const (
	STopicTypeReceive           STopicType = 4002
	STopicTypeAlinkRRPC         STopicType = 4003
	STopicTypeAlinkOTA          STopicType = 4004
	STopicTypeAlinkPropertySet  STopicType = 4007
	STopicTypeAlinkCallService  STopicType = 4009
	STopicTypeRemote            STopicType = 4013
	STopicTypeAlinkRemoteUpdate STopicType = 4015
)

type Mqtt struct {
	MqttClient mqtt.Client
	MqttConfig *MQTTConfig
}
type MQTTConfig struct {
	MQTTName     string           `json:"MQTTName"`
	Valid        bool             `json:"Vaild"`
	ServerType   ServerType       `json:"ServerType"`
	Server       string           `json:"Server"`
	Port         int              `json:"Port"`
	User         string           `json:"User"`
	PW           string           `json:"PW"`
	ClientID     string           `json:"ClientID"`
	CleanSession bool             `json:"CleanSession"`
	SSL          bool             `json:"SSL"`
	KeepAlive    int              `json:"KeepAlive"`
	PubTopics    []PubTopicStruct `json:"PubTopics"`
	SubTopics    []SubTopicStruct `json:"SubTopics"`
	CAFile       string           `json:"CAFile"`
	CertFile     string           `json:"CertFile"`
	KeyFile      string           `json:"KeyFile"`
	NetworkType  int              `json:"NetworkType"`
}
type PubTopicStruct struct {
	Topic       string      `json:"Topic"`
	QoS         int         `json:"QoS"`
	Type        PTopicType  `json:"Type"`
	PayloadType PayloadType `json:"PayloadType"`
	UpIntervalS int         `json:"UpIntervalS"`
	Valid       bool        `json:"Vaild"`
}
type SubTopicStruct struct {
	Topic       string      `json:"Topic"`
	QoS         int         `json:"QoS"`
	Type        STopicType  `json:"Type"`
	PayloadType PayloadType `json:"PayloadType"`
	Valid       bool        `json:"Vaild"`
}

type MqttConfigStruct struct {
	MqttConfigs []*MQTTConfig `json:"MQTTConfigs"`
}

type IMqttUseCase interface {
	PublishDataPoints()
	//	GetMqtt() ([]*Mqtt, error)
	//	GetMqttByName(name string) (*Mqtt, error)
	//	GetMqttByClient(client mqtt.Client) (*Mqtt, error)
	//	GetMqttConfigBySubTopicAndClient(subTopic string, client mqtt.Client) (*Mqtt, *SubTopicStruct, error)
}
