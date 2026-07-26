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
	AiConfig       AiConfig      `yaml:"AiConfig"`
	MilvusConfig   MilvusConfig  `yaml:"MilvusConfig"`
	GeneralConfig  GeneralConfig `yaml:"GeneralConfig"`
	NetworkConfig  NetworkConfig `yaml:"NetworkConfig"`
	MySQLConfig    MySQLConfig   `yaml:"MySQLConfig"`
	DefaultInclude []string      `yaml:"DefaultInclude"`
	DefaultExclude []string      `yaml:"DefaultExclude"`
}
type AiConfig struct {
	ChatApiKey  string `yaml:"chat_api_key"`
	ChatUrl     string `yaml:"chat_url"`
	ChatModel   string `yaml:"chat_model"`
	ChatContext int    `yaml:"chat_context"`
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
	NetworkHost string `yaml:"network_host"`
	NetworkPort string `yaml:"network_port"`
}
type MySQLConfig struct {
	UserName     string `yaml:"user_name"`
	Password     string `yaml:"password"`
	Host         string `yaml:"host"`
	Port         string `yaml:"port"`
	DataBaseName string `yaml:"database_name"`
	ExtraConfig  string `yaml:"extra_config"`
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
func GetChatContext() int {
	return configData.AiConfig.ChatContext
}
func GetMachineId() int64 {
	return configData.GeneralConfig.MachineId
}
func GetNetworkHost() string {
	return configData.NetworkConfig.NetworkHost
}
func GetNetworkPort() string {
	return configData.NetworkConfig.NetworkPort
}
func GetDatabaseName() string {
	return configData.MySQLConfig.DataBaseName
}
func GetDatabaseHost() string {
	return configData.MySQLConfig.Host
}
func GetDatabasePort() string {
	return configData.MySQLConfig.Port
}
func GetDataBaseUserName() string {
	return configData.MySQLConfig.UserName
}
func GetDataBasePassword() string {
	return configData.MySQLConfig.Password
}
func GetDataBaseExtraConfig() string {
	return configData.MySQLConfig.ExtraConfig
}
func GetDefaultInclude() []string {
	return configData.DefaultInclude
}
func GetDefaultExclude() []string {
	return configData.DefaultExclude
}
