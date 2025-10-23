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

import init from '@web3-onboard/core';
import injectedModule from '@web3-onboard/injected-wallets';
import walletConnectModule from '@web3-onboard/walletconnect';
import coinbaseWalletModule from '@web3-onboard/coinbase';
import { 
  SUPPORTED_CHAINS, 
  WALLETCONNECT_CHAINS, 
  WALLETCONNECT_PROJECT_ID 
} from '../constants/web3.constants';

// 支持的链配置（从配置文件导入）
const chains = SUPPORTED_CHAINS.map(chain => ({
  id: chain.id,
  token: chain.token,
  label: chain.label,
  rpcUrl: chain.rpcUrl
}));

// 钱包模块配置
const injected = injectedModule();

const walletConnect = walletConnectModule({
  projectId: WALLETCONNECT_PROJECT_ID, // 从配置文件导入
  requiredChains: WALLETCONNECT_CHAINS,
  dappUrl: window.location.origin
});

const coinbaseWallet = coinbaseWalletModule({ 
  darkMode: true 
});

// 初始化 Web3-Onboard
const web3Onboard = init({
  wallets: [injected, walletConnect, coinbaseWallet],
  chains: chains,
  appMetadata: {
    name: 'New API',
    icon: '/logo.png',
    description: 'AI API Gateway - Web3 Payment',
    recommendedInjectedWallets: [
      { name: 'MetaMask', url: 'https://metamask.io' },
      { name: 'Coinbase', url: 'https://wallet.coinbase.com/' }
    ]
  },
  accountCenter: {
    desktop: {
      enabled: true,
      position: 'topRight'
    },
    mobile: {
      enabled: true,
      position: 'topRight'
    }
  },
  connect: {
    autoConnectLastWallet: true
  }
});

export default web3Onboard;
