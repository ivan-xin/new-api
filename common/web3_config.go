/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

package common

// ChainConfig 链配置结构
type ChainConfig struct {
	ID                    string // 十六进制链ID，如 "0x2105"
	ChainID               int64  // 十进制链ID，如 8453
	Name                  string // 链名称（小写），如 "base"
	Label                 string // 显示名称，如 "Base"
	Token                 string // 原生代币，如 "ETH"
	RpcURL                string // 默认RPC URL
	DefaultRpcPlaceholder string // 默认RPC占位符
	IsRecommended         bool   // 是否推荐
	IsTestnet             bool   // 是否为测试网
	Description           string // 描述
	Explorer              string // 区块浏览器
	Category              string // 类别：L1/L2
	USDCContract          string // USDC合约地址
}

// IsTestnetEnabled 检查是否启用测试网
// 通过环境变量 WEB3_ENABLE_TESTNET 控制
// 默认为 false（生产环境不启用测试网）
func IsTestnetEnabled() bool {
	return GetEnvOrDefaultBool("WEB3_ENABLE_TESTNET", false)
}

// SupportedChains 支持的链配置列表
var SupportedChains = []ChainConfig{
	// ========== 主网 Mainnet ==========
	{
		ID:                    "0x2105",
		ChainID:               8453,
		Name:                  "base",
		Label:                 "Base",
		Token:                 "ETH",
		RpcURL:                "https://mainnet.base.org",
		DefaultRpcPlaceholder: "https://base-mainnet.infura.io/v3/YOUR_KEY",
		IsRecommended:         true,
		IsTestnet:             false,
		Description:           "Coinbase L2 网络",
		Explorer:              "https://basescan.org",
		Category:              "L2",
		USDCContract:          "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913", // USDC Native
	},
	{
		ID:                    "0xa4b1",
		ChainID:               42161,
		Name:                  "arbitrum",
		Label:                 "Arbitrum One",
		Token:                 "ETH",
		RpcURL:                "https://arb1.arbitrum.io/rpc",
		DefaultRpcPlaceholder: "https://arb-mainnet.g.alchemy.com/v2/YOUR_KEY",
		IsRecommended:         false,
		IsTestnet:             false,
		Description:           "Layer 2 扩容网络",
		Explorer:              "https://arbiscan.io",
		Category:              "L2",
		USDCContract:          "0xaf88d065e77c8cC2239327C5EDb3A432268e5831", // USDC Native
	},
	{
		ID:                    "0x38",
		ChainID:               56,
		Name:                  "bsc",
		Label:                 "BSC",
		Token:                 "BNB",
		RpcURL:                "https://bsc-dataseed.binance.org",
		DefaultRpcPlaceholder: "https://bsc-dataseed1.binance.org",
		IsRecommended:         false,
		IsTestnet:             false,
		Description:           "币安智能链",
		Explorer:              "https://bscscan.com",
		Category:              "L1",
		USDCContract:          "0x8AC76a51cc950d9822D68b83fE1Ad97B32Cd580d", // USDC
	},
	{
		ID:                    "0x1",
		ChainID:               1,
		Name:                  "ethereum",
		Label:                 "Ethereum",
		Token:                 "ETH",
		RpcURL:                "https://eth.llamarpc.com",
		DefaultRpcPlaceholder: "https://mainnet.infura.io/v3/YOUR_KEY",
		IsRecommended:         false,
		IsTestnet:             false,
		Description:           "以太坊主网（Gas 费较高）",
		Explorer:              "https://etherscan.io",
		Category:              "L1",
		USDCContract:          "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48", // USDC
	},

	// ========== 测试网 Testnet ==========
	{
		ID:                    "0x14a34",
		ChainID:               84532,
		Name:                  "base-sepolia",
		Label:                 "Base Sepolia",
		Token:                 "ETH",
		RpcURL:                "https://sepolia.base.org",
		DefaultRpcPlaceholder: "https://base-sepolia.infura.io/v3/YOUR_KEY",
		IsRecommended:         false,
		IsTestnet:             true,
		Description:           "Base 测试网络（Sepolia）",
		Explorer:              "https://sepolia.basescan.org",
		Category:              "L2",
		USDCContract:          "0xb27fE6ACf871B72C58ae450c3399D4CdA0349c03", // USDC 自定义部署
	},
}

// USDCContracts USDC合约地址映射 (链名称 -> 合约地址)
var USDCContracts = map[string]string{
	// 主网
	"base":     "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
	"arbitrum": "0xaf88d065e77c8cC2239327C5EDb3A432268e5831",
	"bsc":      "0x8AC76a51cc950d9822D68b83fE1Ad97B32Cd580d",
	"ethereum": "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",

	// 测试网
	"base-sepolia": "0xb27fE6ACf871B72C58ae450c3399D4CdA0349c03",
}

// RPCConfigKeys RPC配置字段映射 (链名称 -> 配置键)
var RPCConfigKeys = map[string]string{
	// 主网
	"base":     "Web3RPC_BASE",
	"arbitrum": "Web3RPC_ARBITRUM",
	"bsc":      "Web3RPC_BSC",
	"ethereum": "Web3RPC_ETHEREUM",

	// 测试网
	"base-sepolia": "Web3RPC_BASE_SEPOLIA",
}

// DefaultChainName 默认推荐的链名称
const DefaultChainName = "base"

// GetChainConfig 根据链名称获取链配置
func GetChainConfig(chainName string) *ChainConfig {
	for _, chain := range SupportedChains {
		if chain.Name == chainName {
			return &chain
		}
	}
	return nil
}

// GetChainConfigByID 根据链ID（十六进制）获取链配置
func GetChainConfigByID(chainID string) *ChainConfig {
	for _, chain := range SupportedChains {
		if chain.ID == chainID {
			return &chain
		}
	}
	return nil
}

// GetUSDCContract 根据链名称获取USDC合约地址
func GetUSDCContract(chainName string) string {
	if contract, ok := USDCContracts[chainName]; ok {
		return contract
	}
	return ""
}

// GetRPCConfigKey 根据链名称获取RPC配置键
func GetRPCConfigKey(chainName string) string {
	if key, ok := RPCConfigKeys[chainName]; ok {
		return key
	}
	return ""
}

// IsChainSupported 检查链是否支持
func IsChainSupported(chainName string) bool {
	_, ok := USDCContracts[chainName]
	return ok
}

// IsTestnetChain 检查指定链是否为测试网
func IsTestnetChain(chainName string) bool {
	for _, chain := range SupportedChains {
		if chain.Name == chainName {
			return chain.IsTestnet
		}
	}
	return false
}

// GetSupportedChainNames 获取所有支持的链名称
func GetSupportedChainNames() []string {
	names := make([]string, 0, len(SupportedChains))
	for _, chain := range SupportedChains {
		names = append(names, chain.Name)
	}
	return names
}

// GetRecommendedChains 获取推荐的链配置列表
func GetRecommendedChains() []ChainConfig {
	recommended := make([]ChainConfig, 0)
	for _, chain := range SupportedChains {
		if chain.IsRecommended {
			recommended = append(recommended, chain)
		}
	}
	return recommended
}

// GetMainnetChains 获取主网链配置列表
func GetMainnetChains() []ChainConfig {
	mainnet := make([]ChainConfig, 0)
	for _, chain := range SupportedChains {
		if !chain.IsTestnet {
			mainnet = append(mainnet, chain)
		}
	}
	return mainnet
}

// GetTestnetChains 获取测试网链配置列表
func GetTestnetChains() []ChainConfig {
	testnet := make([]ChainConfig, 0)
	for _, chain := range SupportedChains {
		if chain.IsTestnet {
			testnet = append(testnet, chain)
		}
	}
	return testnet
}

// GetDefaultChainConfig 获取默认链配置
func GetDefaultChainConfig() *ChainConfig {
	return GetChainConfig(DefaultChainName)
}

// GetAvailableChains 获取可用的链配置列表（根据环境变量过滤测试网）
// 如果 WEB3_ENABLE_TESTNET 为 false（默认），则只返回主网链
func GetAvailableChains() []ChainConfig {
	if IsTestnetEnabled() {
		return SupportedChains
	}
	return GetMainnetChains()
}

// GetAvailableChainNames 获取可用的链名称列表（根据环境变量过滤测试网）
func GetAvailableChainNames() []string {
	chains := GetAvailableChains()
	names := make([]string, 0, len(chains))
	for _, chain := range chains {
		names = append(names, chain.Name)
	}
	return names
}
