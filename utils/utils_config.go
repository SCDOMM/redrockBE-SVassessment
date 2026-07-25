package utils

import (
	"fmt"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

var (
	configData ConfigData
	once       sync.Once
)

var _ = yaml.Unmarshal

type ConfigData struct {
	AiConfig      AiConfig      `yaml:"AiConfig"`
	MilvusConfig  MilvusConfig  `yaml:"MilvusConfig"`
	GeneralConfig GeneralConfig `yaml:"GeneralConfig"`
	NetworkConfig NetworkConfig `yaml:"NetworkConfig"`
}
type AiConfig struct {
	ChatApiKey  string `yaml:"chat_api_key"`
	ChatUrl     string `yaml:"chat_url"`
	ChatModel   string `yaml:"chat_model"`
	EmbedApiKey string `yaml:"embed_api_key"`
	EmbedUrl    string `yaml:"embed_url"`
	EmbedModel  string `yaml:"embed_model"`
	EmbedDim    int64  `yaml:"embed_dim"`
}
type MilvusConfig struct {
	CollectionsName string `yaml:"collections_name"`
}
type GeneralConfig struct {
	MachineId int64 `yaml:"machine_id"`
}
type NetworkConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

func init() {
	once.Do(func() {
		dataBytes, err := os.ReadFile("./config.yaml")
		if err != nil {
			fmt.Println("读取配置失败！" + err.Error())
		}
		err = yaml.Unmarshal(dataBytes, &configData)
		if err != nil {
			fmt.Println("解析配置失败！" + err.Error())
		}
	})
}
func GetChatApiKeyConfig() string {
	return configData.AiConfig.ChatApiKey
}
func GetEmbedApiKeyConfig() string {
	return configData.AiConfig.EmbedApiKey
}
func GetChatUrl() string {
	return configData.AiConfig.ChatUrl
}
func GetEmbedUrl() string {
	return configData.AiConfig.EmbedUrl
}
func GetEmbedModel() string {
	return configData.AiConfig.EmbedModel
}
func GetEmbedDim() int64 {
	return configData.AiConfig.EmbedDim
}
func GetCollectionsName() string {
	return configData.MilvusConfig.CollectionsName
}
func GetChatModel() string {
	return configData.AiConfig.ChatModel
}
func GetMachineId() int64 {
	return configData.GeneralConfig.MachineId
}
func GetHost() string {
	return configData.NetworkConfig.Host
}
func GetPort() string {
	return configData.NetworkConfig.Port
}
