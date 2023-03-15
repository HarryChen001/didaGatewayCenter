package usecase

import (
	"crypto/tls"
	"crypto/x509"
	"didaGatewayCenter/domain"
	usecase2 "didaGatewayCenter/mqttMessage/usecase"
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"
	"log"
	"os"
	"time"
)

type Mqtt struct {
	iSU   domain.ISystemUseCase
	iDPU  domain.IDataPointUseCase
	iACU  domain.IAppConfigUseCase
	iLogU domain.ILogUsecase
	nMqtt []*NewMqtt
}

type NewMqtt struct {
	Parent     *Mqtt
	mqttConfig *domain.MQTTConfig
	client     mqtt.Client
	iPMMU      []domain.IMqttMessageUsecase
	iMMUS      []domain.IMqttMessageUsecase
}

func (m *Mqtt) PublishDataPoints() {
	for _, singleMqtt := range m.nMqtt {
		for _, singleIMMU := range singleMqtt.iPMMU {
			topicName := singleIMMU.GetTopicName()
			msg, _ := singleIMMU.GetPublishMsg(true)
			_ = singleMqtt.client.Publish(topicName, 0, false, msg)
		}
	}
}
func NewMqttUseCase(iACU domain.IAppConfigUseCase, iLog domain.ILogUsecase, useCase domain.ISystemUseCase, iDPU domain.IDataPointUseCase) domain.IMqttUseCase {
	m := &Mqtt{
		iSU:   useCase,
		iDPU:  iDPU,
		iACU:  iACU,
		iLogU: iLog,
	}
	mqttConfig := domain.MqttConfigStruct{}
	logger := iLog.GetLogger()
	mqttConfigFile := iACU.GetConfig().MqttConfig.File
	fileInfo, err := os.ReadFile(mqttConfigFile)
	if err != nil {
		iLog.GetLogger().Error("Failed to read mqtt config file", zap.String("filename", mqttConfigFile), zap.Error(err))
		return nil
	}
	if err := json.Unmarshal(fileInfo, &mqttConfig); err != nil {
		logger.Error("cannot unmarshal mqtt config", zap.String("filename", mqttConfigFile), zap.Error(err))
		return nil
	}
	for _, singleMqtt := range mqttConfig.MqttConfigs {
		m.nMqtt = append(m.nMqtt, &NewMqtt{
			Parent:     m,
			mqttConfig: singleMqtt,
		})
	}
	if iACU.IsDebug() {
		mqtt.DEBUG = log.New(os.Stdout, "", log.LstdFlags)
		mqtt.WARN = log.New(os.Stdout, "", log.LstdFlags)
		mqtt.CRITICAL = log.New(os.Stdout, "", log.LstdFlags)
		mqtt.ERROR = log.New(os.Stdout, "", log.LstdFlags)
	}
	m.addDidaMeter(useCase)
	for _, singleMqtt := range m.nMqtt {
		go singleMqtt.Connect()
	}
	return m
}
func (n *NewMqtt) Connect() {
	singleMqtt := n.mqttConfig
	if !singleMqtt.Valid {
		return
	}
	mqttName := singleMqtt.MQTTName
	broker := fmt.Sprintf("%s:%d", singleMqtt.Server, singleMqtt.Port)
	opt := mqtt.NewClientOptions()

	opt.SetUsername(singleMqtt.User)
	opt.SetPassword(singleMqtt.PW)

	opt.SetClientID(singleMqtt.ClientID)
	opt.SetCleanSession(singleMqtt.CleanSession)
	opt.SetKeepAlive(time.Duration(singleMqtt.KeepAlive) * time.Second)

	opt.SetConnectTimeout(time.Second * 5)

	//reconnect when connection lost
	opt.SetAutoReconnect(true)
	//	opt.SetMaxReconnectInterval(time.Second * 10)
	//retry when first connection failed
	opt.SetConnectRetry(true)
	//	opt.SetConnectRetryInterval(time.Second * 10)

	if n.mqttConfig.SSL {
		certPool := x509.NewCertPool()
		certPool.AppendCertsFromPEM([]byte(n.mqttConfig.CAFile))
		pair, _ := tls.X509KeyPair([]byte(n.mqttConfig.CertFile), []byte(n.mqttConfig.KeyFile))
		opt.SetTLSConfig(&tls.Config{
			RootCAs:            certPool,
			Certificates:       []tls.Certificate{pair},
			InsecureSkipVerify: false,
		})
		broker = fmt.Sprintf("tls://%s", broker)
	}

	opt.AddBroker(broker)

	opt.OnConnect = n.onConnect
	opt.OnConnectionLost = n.onConnectLost

	n.client = mqtt.NewClient(opt)

	if token := n.client.Connect(); token.Wait() && token.Error() != nil {
		n.Parent.iLogU.GetLogger().Error("mqtt connect failed", zap.String("name", mqttName),
			zap.String("broker", broker), zap.String("clientId", singleMqtt.ClientID),
			zap.String("username", singleMqtt.User), zap.Error(token.Error()))
	}
	if n.mqttConfig.ServerType == domain.ServerTypeAlinkType {
		n.client.AddRoute("/ext/network/probe/+", func(client mqtt.Client, message mqtt.Message) {})
		if err := n.alinkDeviceInfoUpload(); err != nil {
			n.Parent.iLogU.GetLogger().Warn("upload device info to alink failed", zap.String("mqttName", mqttName), zap.Error(err))
		}
		if err := n.alinkOtaInfoUpload(); err != nil {
			n.Parent.iLogU.GetLogger().Warn("upload ota info to alink failed", zap.String("mqttName", mqttName), zap.Error(err))
		}
	}
	for _, singlePublishTopic := range n.mqttConfig.PubTopics {
		if !singlePublishTopic.Valid {
			continue
		}
		switch singlePublishTopic.Type {
		case domain.PTopicTypeUpload, domain.PTopicTypeAlinkPropertyPost:
			payloadName := fmt.Sprintf("P%d.json", singlePublishTopic.PayloadType)
			p, err := usecase2.NewMqttMessageUsecase(mqttName, singlePublishTopic.Topic, payloadName, n.Parent.iACU, n.Parent.iDPU)
			if err != nil {
				n.Parent.iLogU.GetLogger().Warn("failed to set publish message format", zap.String("mqttName", mqttName),
					zap.String("topic", singlePublishTopic.Topic), zap.Error(err))
				continue
			} else {
				n.Parent.iLogU.GetLogger().Info("set publish message format success", zap.String("mqttName", mqttName), zap.String("topic", singlePublishTopic.Topic))
			}
			n.iPMMU = append(n.iPMMU, p)
			go n.publishMsg(singlePublishTopic.Topic, byte(singlePublishTopic.QoS), p, time.Duration(singlePublishTopic.UpIntervalS)*time.Second)

		}
	}
}

func (n *NewMqtt) publishMsg(topic string, qos byte, publishUsecase domain.IMqttMessageUsecase, interval time.Duration) {
	client := n.client
	iLogU := n.Parent.iLogU
	m2, _ := publishUsecase.GetPublishMsg(false)
	if token := client.Publish(topic, qos, false, m2); token.Wait() {
		if token.Error() != nil {
			iLogU.GetLogger().Warn("publish dataPoints message error", zap.String("mqttName", publishUsecase.GetMqttName()),
				zap.String("payloadName", publishUsecase.GetPayloadName()), zap.Error(token.Error()))
		} else {
			iLogU.GetLogger().Info("publish dataPoints message success",
				zap.String("mqttName", publishUsecase.GetMqttName()), zap.String("payloadName", publishUsecase.GetPayloadName()))
		}
	}

	timer := time.NewTimer(interval)

	for {
		select {
		case <-timer.C:
			timer.Reset(interval)
		default:
			time.Sleep(time.Millisecond * 100)
			continue
		}

		if !client.IsConnectionOpen() {
			continue
		}
		p, _ := publishUsecase.GetPublishMsg(false)
		if token := client.Publish(topic, qos, false, p); token.Wait() {
			if token.Error() != nil {
				iLogU.GetLogger().Warn("publish message failed", zap.String("topic", topic),
					zap.Error(token.Error()))
			} else {
				iLogU.GetLogger().Debug("publish message succeeded", zap.String("topic", topic),
					zap.String("payload", string(p)))
			}
		}
	}
}

func (n *NewMqtt) onConnect(client mqtt.Client) {

	opt := client.OptionsReader()
	clientId := opt.ClientID()
	userName := opt.Username()
	broker := opt.Servers()

	mqttName := n.mqttConfig.MQTTName
	n.Parent.iLogU.GetLogger()
	n.Parent.iLogU.GetLogger().Info("mqtt connected", zap.String("name", mqttName),
		zap.String("broker", broker[0].String()), zap.String("clientId", clientId),
		zap.String("username", userName))

	n.subscribe()
}

func (n *NewMqtt) subscribe() {
	mqttName := n.mqttConfig.MQTTName
	for _, singSubTopic := range n.mqttConfig.SubTopics {
		topicName := singSubTopic.Topic
		payloadType := singSubTopic.PayloadType
		qos := singSubTopic.QoS
		switch singSubTopic.Type {
		case domain.STopicTypeReceive:
			payloadName := fmt.Sprintf("S%d.json", payloadType)
			s, err := usecase2.NewMqttMessageUsecase(mqttName, topicName, payloadName, n.Parent.iACU, n.Parent.iDPU)
			if err != nil {
				n.Parent.iLogU.GetLogger().Error("set subscribe message format failed", zap.String("mqttName", mqttName),
					zap.String("topic", topicName), zap.String("payloadName", payloadName), zap.Error(err))
				continue
			} else {
				n.Parent.iLogU.GetLogger().Info("set subscribe message format success", zap.String("mqttName", mqttName),
					zap.String("topic", topicName), zap.String("payloadName", payloadName))
			}
			token := n.client.Subscribe(topicName, byte(qos), func(client mqtt.Client, message mqtt.Message) {
				s.GetCallBack().DataPointSet(client, message)
				go n.Parent.PublishDataPoints()
			})
			if token.Wait() {
				if err := token.Error(); err != nil {
					n.Parent.iLogU.GetLogger().Error("subscribe topic failed", zap.String("mqttName", mqttName), zap.String("topic", topicName), zap.Error(err))
				}
				n.Parent.iLogU.GetLogger().Info("subscribe topic succeeded", zap.String("mqttName", mqttName), zap.String("topic", topicName))
			}
		}
	}
}
func (n *NewMqtt) onConnectLost(client mqtt.Client, err error) {
	opt := client.OptionsReader()
	clientId := opt.ClientID()
	userName := opt.Username()
	broker := opt.Servers()

	n.Parent.iLogU.GetLogger().Warn("mqtt connection lost", zap.String("name", n.mqttConfig.MQTTName),
		zap.String("broker", broker[0].String()), zap.String("clientId", clientId),
		zap.String("username", userName), zap.Error(err))

}
