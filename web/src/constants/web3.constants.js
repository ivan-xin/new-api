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

/**
 * Web3 区块链配置常量
 * 
 * 包含所有支持的区块链网络配置信息
 * 包括链ID、名称、RPC URL、USDC合约地址等
 */

// ==================== 链配置 ====================

/**
 * 支持的区块链配置
 * @type {Array<{id: string, chainId: number, name: string, label: string, token: string, rpcUrl: string, isRecommended: boolean, description: string}>}
 */
export const SUPPORTED_CHAINS = [
  // ========== 主网 Mainnet ==========
  {
    id: '0x2105',              // Hex Chain ID
    chainId: 8453,             // Decimal Chain ID
    name: 'base',              // 内部使用的名称（小写）
    label: 'Base',             // 显示名称
    token: 'ETH',              // 原生代币
    rpcUrl: 'https://mainnet.base.org',
    defaultRpcPlaceholder: 'https://base-mainnet.infura.io/v3/YOUR_KEY',
    isRecommended: true,
    isTestnet: false,
    description: 'Coinbase L2 网络',
    explorer: 'https://basescan.org',
    category: 'L2'
  },
  {
    id: '0xa4b1',
    chainId: 42161,
    name: 'arbitrum',
    label: 'Arbitrum One',
    token: 'ETH',
    rpcUrl: 'https://arb1.arbitrum.io/rpc',
    defaultRpcPlaceholder: 'https://arb-mainnet.g.alchemy.com/v2/YOUR_KEY',
    isRecommended: false,
    isTestnet: false,
    description: 'Layer 2 扩容网络',
    explorer: 'https://arbiscan.io',
    category: 'L2'
  },
  {
    id: '0x38',
    chainId: 56,
    name: 'bsc',
    label: 'BSC',
    token: 'BNB',
    rpcUrl: 'https://bsc-dataseed.binance.org',
    defaultRpcPlaceholder: 'https://bsc-dataseed1.binance.org',
    isRecommended: false,
    isTestnet: false,
    description: '币安智能链',
    explorer: 'https://bscscan.com',
    category: 'L1'
  },
  {
    id: '0x1',
    chainId: 1,
    name: 'ethereum',
    label: 'Ethereum',
    token: 'ETH',
    rpcUrl: 'https://eth.llamarpc.com',
    defaultRpcPlaceholder: 'https://mainnet.infura.io/v3/YOUR_KEY',
    isRecommended: false,
    isTestnet: false,
    description: '以太坊主网（Gas 费较高）',
    explorer: 'https://etherscan.io',
    category: 'L1'
  },
  
  // ========== 测试网 Testnet ==========
  {
    id: '0x14a34',            // Hex Chain ID (84532)
    chainId: 84532,           // Decimal Chain ID
    name: 'base-sepolia',     // 内部使用的名称（小写）
    label: 'Base Sepolia',    // 显示名称
    token: 'ETH',             // 原生代币
    rpcUrl: 'https://sepolia.base.org',
    defaultRpcPlaceholder: 'https://base-sepolia.infura.io/v3/YOUR_KEY',
    isRecommended: false,
    isTestnet: true,
    description: 'Base 测试网络（Sepolia）',
    explorer: 'https://sepolia.basescan.org',
    category: 'L2'
  }
];

/**
 * USDC 合约地址映射
 * @type {Object<string, string>}
 */
export const USDC_CONTRACTS = {
  // 主网
  '0x2105': '0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913',  // Base (USDC Native)
  '0xa4b1': '0xaf88d065e77c8cC2239327C5EDb3A432268e5831',  // Arbitrum (USDC Native)
  '0x38': '0x8AC76a51cc950d9822D68b83fE1Ad97B32Cd580d',    // BSC (USDC)
  '0x1': '0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48',     // Ethereum (USDC)
  
  // 测试网 (自定义部署的测试代币)
  '0x14a34': '0xb27fE6ACf871B72C58ae450c3399D4CdA0349c03'  // Base Sepolia (USDC - 自定义部署)
};

/**
 * 链ID到链名称的映射
 * @type {Object<string, string>}
 */
export const CHAIN_NAMES = {
  // 主网
  '0x2105': 'base',
  '0xa4b1': 'arbitrum',
  '0x38': 'bsc',
  '0x1': 'ethereum',
  
  // 测试网
  '0x14a34': 'base-sepolia'
};

/**
 * 链ID到显示标签的映射
 * @type {Object<string, string>}
 */
export const CHAIN_LABELS = {
  // 主网
  '0x2105': 'Base',
  '0xa4b1': 'Arbitrum',
  '0x38': 'BSC',
  '0x1': 'Ethereum',
  
  // 测试网
  '0x14a34': 'Base Sepolia'
};

/**
 * 链ID到十进制链ID的映射
 * @type {Object<string, number>}
 */
export const CHAIN_IDS = {
  // 主网
  '0x2105': 8453,
  '0xa4b1': 42161,
  '0x38': 56,
  '0x1': 1,
  
  // 测试网
  '0x14a34': 84532
};

// ==================== 默认配置 ====================

/**
 * 默认推荐的链
 */
export const DEFAULT_CHAIN = SUPPORTED_CHAINS.find(chain => chain.isRecommended) || SUPPORTED_CHAINS[0];

/**
 * 默认链名称
 */
export const DEFAULT_CHAIN_NAME = DEFAULT_CHAIN.name;

/**
 * 默认链ID（十六进制）
 */
export const DEFAULT_CHAIN_ID = DEFAULT_CHAIN.id;

// ==================== RPC 配置字段 ====================

/**
 * RPC 配置字段名映射
 * 用于后端配置项的字段名
 */
export const RPC_CONFIG_FIELDS = {
  // 主网
  'base': 'Web3RPC_BASE',
  'arbitrum': 'Web3RPC_ARBITRUM',
  'bsc': 'Web3RPC_BSC',
  'ethereum': 'Web3RPC_ETHEREUM',
  
  // 测试网
  'base-sepolia': 'Web3RPC_BASE_SEPOLIA'
};

// ==================== USDC ABI ====================

/**
 * USDC 合约 ABI
 * 包含常用的转账、余额查询等方法
 */
export const USDC_ABI = [
  'function transfer(address to, uint256 amount) returns (bool)',
  'function balanceOf(address owner) view returns (uint256)',
  'function decimals() view returns (uint8)',
  'function symbol() view returns (string)',
  'function name() view returns (string)',
  'function totalSupply() view returns (uint256)',
  'function allowance(address owner, address spender) view returns (uint256)',
  'function approve(address spender, uint256 amount) returns (bool)',
  'function transferFrom(address from, address to, uint256 amount) returns (bool)'
];

// ==================== WalletConnect 配置 ====================

/**
 * WalletConnect 需要的链ID列表（十进制）
 * 注意：这是静态导出，在实际使用时应该调用 getAvailableChainIds()
 */
export const WALLETCONNECT_CHAINS = SUPPORTED_CHAINS.map(chain => chain.chainId);

/**
 * 获取可用的链ID列表（十进制，用于 WalletConnect）
 * @returns {Array<number>} 可用的链ID列表
 */
export function getAvailableChainIds() {
  return getAvailableChains().map(chain => chain.chainId);
}

/**
 * WalletConnect Project ID
 * 注意：需要替换为你自己的 Project ID
 * 在 https://cloud.walletconnect.com/ 创建
 */
export const WALLETCONNECT_PROJECT_ID = '2c8e0b58c2d567f5c7d6e1f4b3a8c9d0';

// ==================== 工具函数 ====================

/**
 * 根据链ID获取链配置
 * @param {string} chainId - 十六进制链ID（如 '0x2105'）
 * @returns {Object|null} 链配置对象
 */
export function getChainConfig(chainId) {
  return SUPPORTED_CHAINS.find(chain => chain.id === chainId) || null;
}

/**
 * 根据链名称获取链配置
 * @param {string} chainName - 链名称（如 'base'）
 * @returns {Object|null} 链配置对象
 */
export function getChainConfigByName(chainName) {
  return SUPPORTED_CHAINS.find(chain => chain.name === chainName) || null;
}

/**
 * 获取 USDC 合约地址
 * @param {string} chainId - 十六进制链ID
 * @returns {string|null} USDC 合约地址
 */
export function getUSDCContract(chainId) {
  return USDC_CONTRACTS[chainId] || null;
}

/**
 * 检查链是否支持
 * @param {string} chainId - 十六进制链ID
 * @returns {boolean} 是否支持
 */
export function isChainSupported(chainId) {
  return SUPPORTED_CHAINS.some(chain => chain.id === chainId);
}

/**
 * 获取推荐的链列表
 * @returns {Array} 推荐的链配置列表
 */
export function getRecommendedChains() {
  return SUPPORTED_CHAINS.filter(chain => chain.isRecommended);
}

/**
 * 获取主网链列表
 * @returns {Array} 主网链配置列表
 */
export function getMainnetChains() {
  return SUPPORTED_CHAINS.filter(chain => !chain.isTestnet);
}

/**
 * 获取测试网链列表
 * @returns {Array} 测试网链配置列表
 */
export function getTestnetChains() {
  return SUPPORTED_CHAINS.filter(chain => chain.isTestnet);
}

/**
 * 检查是否启用测试网
 * 通过环境变量 VITE_ENABLE_TESTNET 控制
 * 默认为 false（生产环境不启用测试网）
 * @returns {boolean} 是否启用测试网
 */
export function isTestnetEnabled() {
  const envValue = import.meta.env.VITE_ENABLE_TESTNET;
  if (envValue === undefined || envValue === null || envValue === '') {
    return false; // 默认禁用测试网
  }
  return envValue === 'true' || envValue === '1' || envValue === true;
}

/**
 * 获取可用的链配置列表（根据环境变量过滤测试网）
 * 如果 VITE_ENABLE_TESTNET 为 false（默认），则只返回主网链
 * @returns {Array} 可用的链配置列表
 */
export function getAvailableChains() {
  if (isTestnetEnabled()) {
    return SUPPORTED_CHAINS;
  }
  return getMainnetChains();
}

/**
 * 检查指定链ID是否为测试网
 * @param {string} chainId - 十六进制链ID
 * @returns {boolean} 是否为测试网
 */
export function isTestnetChain(chainId) {
  const chain = getChainConfig(chainId);
  return chain ? chain.isTestnet : false;
}

/**
 * 获取所有支持的链ID（十六进制）
 * @returns {Array<string>} 链ID列表
 */
export function getSupportedChainIds() {
  return SUPPORTED_CHAINS.map(chain => chain.id);
}

/**
 * 获取所有支持的链名称
 * @returns {Array<string>} 链名称列表
 */
export function getSupportedChainNames() {
  return SUPPORTED_CHAINS.map(chain => chain.name);
}

/**
 * 格式化链名称用于显示
 * @param {string} chainNameOrId - 链名称或链ID
 * @returns {string} 格式化后的显示名称
 */
export function formatChainLabel(chainNameOrId) {
  // 如果是链ID（以0x开头）
  if (chainNameOrId.startsWith('0x')) {
    const config = getChainConfig(chainNameOrId);
    return config ? config.label : chainNameOrId;
  }
  // 如果是链名称
  const config = getChainConfigByName(chainNameOrId);
  return config ? config.label : chainNameOrId;
}

/**
 * 获取 RPC 配置字段名
 * @param {string} chainName - 链名称
 * @returns {string|null} RPC 配置字段名
 */
export function getRPCConfigField(chainName) {
  return RPC_CONFIG_FIELDS[chainName] || null;
}
