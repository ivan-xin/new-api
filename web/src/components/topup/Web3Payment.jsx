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

import React, { useState, useEffect } from 'react';
import { Modal, Button, Card, Space, Typography, Spin, Toast, Select } from '@douyinfe/semi-ui';
import { useConnectWallet } from '@web3-onboard/react';
import { ethers } from 'ethers';
import { API } from '../../helpers';
import { IconCreditCard } from '@douyinfe/semi-icons';
import { useTranslation } from 'react-i18next';
import {
  USDC_ABI,
  USDC_CONTRACTS,
  CHAIN_NAMES,
  CHAIN_LABELS,
  DEFAULT_CHAIN_NAME,
  isChainSupported,
  getChainConfig,
  isTestnetChain,
  getAvailableChains,
  isTestnetEnabled
} from '../../constants/web3.constants';

const { Text, Title } = Typography;

export default function Web3Payment({ amount, visible, onCancel, onSuccess }) {
  const { t } = useTranslation();
  const [{ wallet, connecting }, connect, disconnect] = useConnectWallet();
  const [loading, setLoading] = useState(false);
  const [payInfo, setPayInfo] = useState(null);
  const [balance, setBalance] = useState('0');
  const [modalVisible, setModalVisible] = useState(false);

  const chainId = wallet?.chains[0]?.id;
  const account = wallet?.accounts[0]?.address;
  const chainName = CHAIN_NAMES[chainId] || DEFAULT_CHAIN_NAME;
  
  // 检查当前连接的链是否为测试网
  const isTestnet = chainId ? isTestnetChain(chainId) : false;
  const currentChainConfig = chainId ? getChainConfig(chainId) : null;
  
  // 获取可用的主网链列表
  const availableChains = getAvailableChains();
  
  // 检查是否允许测试网（根据环境变量）
  const testnetAllowed = isTestnetEnabled();
  
  // 只有在测试网未启用且当前是测试网时才显示警告
  const shouldWarnTestnet = isTestnet && !testnetAllowed;

  // 同步外部 visible 到内部 modalVisible
  useEffect(() => {
    if (visible) {
      setModalVisible(true);
    } else {
      setModalVisible(false);
    }
  }, [visible]);

  // 监听钱包连接状态，连接完成后恢复显示 Modal
  useEffect(() => {
    if (wallet && !connecting && visible) {
      // 钱包连接成功，且父组件 visible 为 true 时，恢复显示 Modal
      setModalVisible(true);
    }
  }, [wallet, connecting, visible]);

  // 获取 USDC 余额
  useEffect(() => {
    if (wallet && chainId && account) {
      fetchBalance();
    }
  }, [wallet, chainId, account]);

  const fetchBalance = async () => {
    try {
      const provider = new ethers.BrowserProvider(wallet.provider);
      const usdcAddress = USDC_CONTRACTS[chainId];

      if (!usdcAddress) {
        return;
      }

      const usdcContract = new ethers.Contract(usdcAddress, USDC_ABI, provider);
      const bal = await usdcContract.balanceOf(account);
      const decimals = await usdcContract.decimals();

      setBalance(ethers.formatUnits(bal, decimals));
    } catch (error) {
      console.error(t('获取余额失败:'), error);
    }
  };

  // 获取支付信息
  useEffect(() => {
    if (amount && chainName && visible) {
      fetchPayInfo();
    }
  }, [amount, chainName, visible]);

  const fetchPayInfo = async () => {
    try {
      const res = await API.get(`/api/user/web3/info?amount=${amount}&chain=${chainName}`);
      if (res.data.message === 'success') {
        setPayInfo(res.data.data);
      } else {
        Toast.error(res.data.data || t('获取支付信息失败'));
      }
    } catch (error) {
      console.error(t('获取支付信息失败'), error);
      Toast.error(t('获取支付信息失败'));
    }
  };

  // 切换网络
  const switchToChain = async (targetChainId) => {
    if (!wallet) return;

    try {
      await wallet.provider.request({
        method: 'wallet_switchEthereumChain',
        params: [{ chainId: targetChainId }]
      });
      Toast.success(t('网络切换成功'));
      // 刷新余额和支付信息
      setTimeout(() => {
        fetchBalance();
        fetchPayInfo();
      }, 1000);
    } catch (error) {
      console.error(t('切换网络失败，请在钱包中手动切换'), error);
      if (error.code === 4902) {
        Toast.error(t('该网络未添加到钱包，请手动添加'));
      } else {
        Toast.error(t('切换网络失败，请在钱包中手动切换'));
      }
    }
  };

  // 发起支付
  const handlePay = async () => {
    if (!wallet) {
      Toast.error(t('请先连接钱包'));
      return;
    }

    // 只在测试网未启用时检查并拒绝测试网
    if (shouldWarnTestnet) {
      Toast.error(t('生产环境禁止使用测试网充值，请切换到主网'));
      return;
    }

    if (!USDC_CONTRACTS[chainId]) {
      Toast.error(t('当前网络不支持 USDC 支付，请切换网络'));
      return;
    }

    setLoading(true);

    try {
      // 1. 创建订单
      Toast.info(t('正在创建订单...'));
      const orderRes = await API.post('/api/user/web3/pay', {
        amount: parseInt(amount),
        payment_method: 'web3_usdc',
        chain: chainName
      });

      if (orderRes.data.message !== 'success') {
        throw new Error(orderRes.data.data);
      }

      const orderData = orderRes.data.data;

      // 2. 连接 provider
      const provider = new ethers.BrowserProvider(wallet.provider);
      const signer = await provider.getSigner();

      // 3. 创建 USDC 合约实例
      const usdcContract = new ethers.Contract(USDC_CONTRACTS[chainId], USDC_ABI, signer);

      // 4. 获取 decimals
      const decimals = await usdcContract.decimals();

      // 5. 转换金额
      const amountToSend = ethers.parseUnits(orderData.amount_usdc, decimals);

      // 6. 检查余额
      const userBalance = await usdcContract.balanceOf(account);
      if (userBalance < amountToSend) {
        throw new Error(`${t('USDC 余额不足，需要')} ${orderData.amount_usdc} USDC`);
      }

      // 7. 发起转账
      Toast.info(t('请在钱包中确认交易...'));

      const tx = await usdcContract.transfer(orderData.receiver_address, amountToSend);

      Toast.info(`${t('交易已提交！哈希:')} ${tx.hash.substring(0, 10)}...`);
      Toast.info(t('等待区块确认中，请稍候...'));

      // 8. 等待确认
      const receipt = await tx.wait();

      if (receipt.status !== 1) {
        throw new Error(t('交易失败'));
      }

      Toast.success(t('交易已确认！正在验证...'));

      // 9. 验证交易
      const verifyRes = await API.post('/api/user/web3/verify', {
        trade_no: orderData.trade_no,
        tx_hash: receipt.hash,
        chain: chainName
      });

      if (verifyRes.data.message === 'success') {
        Toast.success(t('充值成功！'));
        if (onSuccess) {
          onSuccess();
        }
        if (onCancel) {
          onCancel();
        }
      } else if (verifyRes.data.message === 'pending') {
        Toast.info(t('交易正在确认中，请稍后刷新页面查看'));
        if (onSuccess) {
          onSuccess();
        }
      } else {
        throw new Error(verifyRes.data.data || t('验证失败'));
      }
    } catch (error) {
      console.error(t('支付失败，请重试'), error);

      if (error.code === 'ACTION_REJECTED' || error.code === 4001) {
        Toast.warning(t('用户取消了交易'));
      } else if (error.message) {
        Toast.error(error.message);
      } else {
        Toast.error(t('支付失败，请重试'));
      }
    } finally {
      setLoading(false);
    }
  };

  const renderContent = () => {
    if (!wallet) {
      return (
        <div style={{ textAlign: 'center', padding: '40px 20px' }}>
          <IconCreditCard size="extra-large" style={{ fontSize: 64, marginBottom: 20 }} />
          <Title heading={4}>{t('选择钱包连接')}</Title>
          <Text type="tertiary">{t('支持 MetaMask、WalletConnect、Coinbase Wallet 等')}</Text>
          <Button
            theme="solid"
            type="primary"
            size="large"
            onClick={() => {
              // 关闭 Modal，避免遮挡 Web3-Onboard 钱包选择界面
              setModalVisible(false);
              // 触发钱包连接
              connect();
            }}
            loading={connecting}
            block
            style={{ marginTop: 24 }}
          >
            {connecting ? t('连接中...') : t('连接钱包')}
          </Button>
        </div>
      );
    }

    return (
      <Space vertical spacing="loose" style={{ width: '100%' }}>
        {/* 测试网警告 - 只在禁用测试网时显示 */}
        {shouldWarnTestnet && (
          <Card 
            bodyStyle={{ 
              padding: 16,
              background: 'var(--semi-color-warning-light-default)',
              border: '2px solid var(--semi-color-warning)'
            }}
          >
            <Space vertical spacing="tight" style={{ width: '100%' }}>
              <Text strong style={{ color: 'var(--semi-color-warning-dark)', fontSize: 16 }}>
                ⚠️ {t('警告：您当前连接的是测试网')}
              </Text>
              <Text type="tertiary">
                {t('当前网络：')}<Text strong>{currentChainConfig?.label || chainName}</Text>{t('（测试网）')}
              </Text>
              <Text type="tertiary">
                {t('生产环境禁止使用测试网充值。请切换到主网（如 Base、Arbitrum One 等）后再进行充值操作。')}
              </Text>
              {availableChains.length > 0 && (
                <Select
                  placeholder={t('选择要切换的主网')}
                  style={{ width: '100%', marginTop: 8 }}
                  onChange={(value) => switchToChain(value)}
                >
                  {availableChains.map(chain => (
                    <Select.Option key={chain.id} value={chain.id}>
                      {chain.label} {chain.isRecommended && t('(推荐)')}
                    </Select.Option>
                  ))}
                </Select>
              )}
            </Space>
          </Card>
        )}
        
        {/* 钱包信息 */}
        <Card bodyStyle={{ padding: 16 }}>
          <Space vertical spacing="tight" style={{ width: '100%' }}>
            <div>
              <Text strong>{t('钱包地址')}</Text>
              <br />
              <Text code copyable style={{ fontSize: 12 }}>
                {account}
              </Text>
            </div>
            <div>
              <Text strong>{t('当前网络')}</Text>
              <br />
              <Text type="success">{CHAIN_LABELS[chainId] || chainName.toUpperCase()}</Text>
            </div>
            <div>
              <Text strong>{t('USDC 余额')}</Text>
              <br />
              <Text type="success" style={{ fontSize: 20, fontWeight: 600 }}>
                {parseFloat(balance).toFixed(2)} USDC
              </Text>
            </div>
          </Space>
        </Card>

        {/* 支付信息 */}
        {payInfo && (
          <Card
            bodyStyle={{
              padding: 16,
              background: 'var(--semi-color-fill-0)'
            }}
          >
            <Text strong>{t('需支付金额')}</Text>
            <Title
              heading={2}
              style={{ margin: '10px 0', color: 'var(--semi-color-success)' }}
            >
              {payInfo.amount_usdc} USDC
            </Title>

            <Text type="tertiary" size="small">
              {t('收款地址:')}{' '}
              {payInfo.receiver_address.substring(0, 10)}...
              {payInfo.receiver_address.substring(payInfo.receiver_address.length - 8)}
            </Text>
          </Card>
        )}

        {/* 网络切换提示 */}
        {!USDC_CONTRACTS[chainId] && (
          <Card
            bodyStyle={{
              padding: 16,
              background: 'var(--semi-color-warning-light-default)'
            }}
          >
            <Text type="warning">{t('⚠️ 当前网络不支持 USDC 支付')}</Text>
            <br />
            <Text type="tertiary" size="small">
              {t('请切换到 Base、BSC、Arbitrum 或 Ethereum 网络')}
            </Text>
            <br />
            <Space style={{ marginTop: 12 }}>
              <Button size="small" onClick={() => switchToChain('0x2105')}>
                {t('切换到 Base')}
              </Button>
              <Button size="small" onClick={() => switchToChain('0x38')}>
                {t('切换到 BSC')}
              </Button>
              <Button size="small" onClick={() => switchToChain('0xa4b1')}>
                {t('切换到 Arbitrum')}
              </Button>
            </Space>
          </Card>
        )}

        {/* 余额不足提示 */}
        {payInfo && parseFloat(balance) < parseFloat(payInfo.amount_usdc) && (
          <Card
            bodyStyle={{
              padding: 16,
              background: 'var(--semi-color-danger-light-default)'
            }}
          >
            <Text type="danger">{t('⚠️ USDC 余额不足')}</Text>
            <br />
            <Text type="tertiary" size="small">
              {t('需要:')}: {payInfo.amount_usdc} USDC
              <br />
              {t('当前:')}: {parseFloat(balance).toFixed(2)} USDC
            </Text>
          </Card>
        )}

        {/* 按钮组 */}
        <Space style={{ width: '100%', marginTop: 12 }}>
          <Button size="large" onClick={() => disconnect(wallet)} style={{ flex: 1 }}>
            {t('断开钱包')}
          </Button>
          <Button
            theme="solid"
            type="primary"
            size="large"
            onClick={handlePay}
            loading={loading}
            disabled={
              shouldWarnTestnet ||
              !USDC_CONTRACTS[chainId] ||
              !payInfo ||
              parseFloat(balance) < parseFloat(payInfo.amount_usdc || 0)
            }
            style={{ flex: 2 }}
          >
            {loading ? t('处理中...') : shouldWarnTestnet ? t('测试网禁止充值') : t('立即支付')}
          </Button>
        </Space>

        {/* 提示信息 */}
        <Text
          type="tertiary"
          size="small"
          style={{ textAlign: 'center', display: 'block', marginTop: 8 }}
        >
          {t('💡 交易确认后自动充值，通常需要 1-3 分钟')}
        </Text>
      </Space>
    );
  };

  return (
    <Modal
      title={t('Web3 USDC 支付')}
      visible={modalVisible}
      onCancel={onCancel}
      footer={null}
      width={500}
      style={{ maxWidth: '90vw' }}
      zIndex={2000}
    >
      <Spin spinning={loading && wallet}>{renderContent()}</Spin>
    </Modal>
  );
}
