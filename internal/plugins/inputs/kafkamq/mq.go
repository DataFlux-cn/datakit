// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package kafkamq  mq
package kafkamq

import (
	"crypto/sha256"
	"crypto/sha512"
	"crypto/tls"
	"crypto/x509"
	"os"
	"strings"

	"github.com/xdg-go/scram"

	"github.com/IBM/sarama"
)

var (
	// kafka 分区分配策略.
	assignors = map[string]sarama.BalanceStrategy{
		"range":      sarama.NewBalanceStrategyRange(),
		"roundrobin": sarama.NewBalanceStrategyRoundRobin(),
		"sticky":     sarama.NewBalanceStrategySticky(),
	}
	defaultAssignors = sarama.NewBalanceStrategyRange() // 轮训模式最适合 datakit 的工作模式.
)

func getKafkaVersion(ver string) sarama.KafkaVersion {
	version, err := sarama.ParseKafkaVersion(ver)
	if err != nil {
		log.Infof("can not get version from conf:[%s], use default version:%s", ver, sarama.DefaultVersion.String())
		return sarama.DefaultVersion
	}
	log.Infof("use version:%s", version.String())
	return version
}

type option func(con *sarama.Config)

func withVersion(version string) option {
	v, err := sarama.ParseKafkaVersion(version)
	if err != nil {
		log.Infof("can not get version from conf:[%s], use default version:%s", version, sarama.DefaultVersion.String())
		v = sarama.DefaultVersion
	}
	return func(con *sarama.Config) {
		con.Version = v
	}
}

func withAssignors(balance string) option {
	var bt sarama.BalanceStrategy
	if assignor, ok := assignors[balance]; ok {
		bt = assignor
	} else {
		log.Infof("can not find assignor, use default `roundrobin`")
		bt = defaultAssignors
	}

	return func(con *sarama.Config) {
		con.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{bt}
	}
}

func withOffset(offset int64) option {
	return func(con *sarama.Config) {
		con.Consumer.Offsets.Initial = sarama.OffsetNewest
		if offset == sarama.OffsetOldest {
			con.Consumer.Offsets.Initial = sarama.OffsetOldest
		}
	}
}

func withSASL(enable bool, protocol, mechanism, username, pw, cert string) option {
	return func(config *sarama.Config) {
		if enable {
			config.Net.SASL.Enable = true
			config.Net.SASL.User = username
			config.Net.SASL.Password = pw
			config.Net.SASL.Mechanism = sarama.SASLMechanism(mechanism)
			config.Net.SASL.Version = sarama.SASLHandshakeV1
			switch strings.ToUpper(mechanism) {
			case sarama.SASLTypeSCRAMSHA512:
				config.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient { return &XDGSCRAMClient{HashGeneratorFcn: SHA512} }
			case sarama.SASLTypeSCRAMSHA256:
				config.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient { return &XDGSCRAMClient{HashGeneratorFcn: SHA256} }
			default:
			}

			if protocol == "SASL_SSL" {
				bts, err := os.ReadFile(cert) //nolint
				if err != nil {
					log.Errorf("can not read file:%s and err=%v", cert, err)
					return
				}
				pool := x509.NewCertPool()
				if ok := pool.AppendCertsFromPEM(bts); !ok {
					log.Errorf("failed to parse root certificate file")
				}

				config.Net.TLS.Config = &tls.Config{
					RootCAs:            pool,
					InsecureSkipVerify: true, //nolint
				}
				config.Net.TLS.Enable = true
			}
		}
	}
}

var (
	SHA256 scram.HashGeneratorFcn = sha256.New
	SHA512 scram.HashGeneratorFcn = sha512.New
)

type XDGSCRAMClient struct {
	*scram.Client
	*scram.ClientConversation
	scram.HashGeneratorFcn
}

func (x *XDGSCRAMClient) Begin(userName, password, authzID string) (err error) {
	x.Client, err = x.HashGeneratorFcn.NewClient(userName, password, authzID)
	if err != nil {
		return err
	}
	x.ClientConversation = x.Client.NewConversation()
	return nil
}

func (x *XDGSCRAMClient) Step(challenge string) (response string, err error) {
	response, err = x.ClientConversation.Step(challenge)
	return
}

func (x *XDGSCRAMClient) Done() bool {
	return x.ClientConversation.Done()
}

func newSaramaConfig(opts ...option) *sarama.Config {
	conf := sarama.NewConfig()
	conf.Consumer.Return.Errors = false

	conf.Consumer.Offsets.Retry.Max = 10
	name, _ := os.Hostname()
	conf.ClientID = name

	for _, opt := range opts {
		opt(conf)
	}
	return conf
}

func getAddrs(addr string, addrs []string) []string {
	kafkaAddress := make([]string, 0)
	if addr != "" {
		kafkaAddress = append(kafkaAddress, addr)
	}
	if addrs != nil {
		kafkaAddress = append(kafkaAddress, addrs...)
	}
	return kafkaAddress
}
