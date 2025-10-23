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

import React, { useEffect, useRef, useState } from 'react';
import { Button, Card, Col, Form, Row, Space, Typography } from '@douyinfe/semi-ui';
import { API, showError, showSuccess, showWarning, compareObjects } from '../../../helpers';
import { useTranslation } from 'react-i18next';
import { CircleDollarSign, AlertCircle } from 'lucide-react';
import { 
  SUPPORTED_CHAINS, 
  RPC_CONFIG_FIELDS,
  getSupportedChainNames,
  formatChainLabel 
} from '../../../constants/web3.constants';

const { Text } = Typography;

export default function SettingsPaymentGatewayWeb3(props) {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  
  // 动态生成初始状态（基于配置文件）
  const getInitialInputs = () => {
    const inputs = { Web3ReceiverAddress: '' };
    SUPPORTED_CHAINS.forEach(chain => {
      const fieldName = RPC_CONFIG_FIELDS[chain.name];
      if (fieldName) {
        inputs[fieldName] = '';
      }
    });
    return inputs;
  };
  
  const [inputs, setInputs] = useState(getInitialInputs());
  const [inputsRow, setInputsRow] = useState(inputs);
  const refForm = useRef();

  useEffect(() => {
    const currentInputs = {};
    
    // 收款地址
    if (props.options['Web3ReceiverAddress'] !== undefined) {
      currentInputs['Web3ReceiverAddress'] = props.options['Web3ReceiverAddress'];
    }
    
    // 动态加载各链的RPC配置
    SUPPORTED_CHAINS.forEach(chain => {
      const fieldName = RPC_CONFIG_FIELDS[chain.name];
      if (fieldName && props.options[fieldName] !== undefined) {
        currentInputs[fieldName] = props.options[fieldName];
      }
    });

    setInputs(currentInputs);
    setInputsRow(structuredClone(currentInputs));
    if (refForm.current) {
      refForm.current.setValues(currentInputs);
    }
  }, [props.options]);

  function onSubmit() {
    const updateArray = compareObjects(inputs, inputsRow);
    if (!updateArray.length) return showWarning(t('你似乎并没有修改什么'));

    const requestQueue = updateArray.map((item) => {
      let value = '';
      if (typeof inputs[item.key] === 'boolean') {
        value = String(inputs[item.key]);
      } else {
        value = inputs[item.key];
      }
      return API.put('/api/option/', {
        key: item.key,
        value,
      });
    });

    setLoading(true);
    Promise.all(requestQueue)
      .then((res) => {
        if (requestQueue.length === 1) {
          if (res.includes(undefined)) return;
        } else if (requestQueue.length > 1) {
          if (res.includes(undefined))
            return showError(t('部分保存失败，请重试'));
        }
        showSuccess(t('保存成功'));
        props.refresh();
      })
      .catch(() => {
        showError(t('保存失败，请重试'));
      })
      .finally(() => {
        setLoading(false);
      });
  }

  const handleInputChange = (name, value) => {
    setInputs((inputs) => ({ ...inputs, [name]: value }));
  };

  return (
    <>
      <Form
        values={inputs}
        getFormApi={(formAPI) => (refForm.current = formAPI)}
        style={{ marginBottom: 15 }}
        labelPosition='left'
        labelAlign='left'
        labelCol={{ span: 6 }}
        wrapperCol={{ span: 18 }}
      >
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: '10px',
            marginBottom: '20px',
          }}
        >
          <CircleDollarSign size={20} color='#2775CA' />
          <Text strong style={{ fontSize: '16px' }}>
            {t('Web3 USDC 支付配置')}
          </Text>
        </div>

        <div
          style={{
            background: 'var(--semi-color-fill-0)',
            padding: '12px 16px',
            borderRadius: '8px',
            marginBottom: '20px',
            border: '1px solid var(--semi-color-border)',
          }}
        >
          <div style={{ display: 'flex', gap: '8px', alignItems: 'start' }}>
            <AlertCircle size={16} color='var(--semi-color-primary)' />
            <div>
              <Text strong style={{ color: 'var(--semi-color-primary)' }}>
                {t('配置说明')}
              </Text>
              <div style={{ marginTop: '8px' }}>
                <Text type='secondary' style={{ fontSize: '13px' }}>
                  {t(
                    '• 收款地址: 必填，用于接收用户支付的 USDC（确保你拥有该地址的私钥）',
                  )}
                </Text>
                <br />
                <Text type='secondary' style={{ fontSize: '13px' }}>
                  {t(
                    '• RPC 节点: 可选，推荐配置以提高交易验证速度和稳定性',
                  )}
                </Text>
                <br />
                <Text type='secondary' style={{ fontSize: '13px' }}>
                  {t('• 推荐 RPC 服务: Infura, Alchemy, QuickNode, Ankr')}
                </Text>
                <br />
                <Text type='secondary' style={{ fontSize: '13px' }}>
                  {t('• 支持的链: ' + 
                    SUPPORTED_CHAINS.map(chain => 
                      chain.label + (chain.isRecommended ? ' (推荐)' : '')
                    ).join(', ')
                  )}
                </Text>
              </div>
            </div>
          </div>
        </div>

        <Form.Section text={t('收款配置')}>
          <Row gutter={16}>
            <Col span={24}>
              <Form.Input
                field='Web3ReceiverAddress'
                label={t('收款地址')}
                placeholder={t('0x...')}
                value={inputs.Web3ReceiverAddress}
                onChange={(value) =>
                  handleInputChange('Web3ReceiverAddress', value)
                }
                rules={[
                  {
                    pattern: /^0x[a-fA-F0-9]{40}$/,
                    message: t('请输入有效的以太坊地址 (0x + 40位十六进制)'),
                  },
                ]}
                extraText={
                  <Text type='tertiary' size='small'>
                    {t('必填 - 用于接收 USDC 的钱包地址')}
                  </Text>
                }
                showClear
              />
            </Col>
          </Row>
        </Form.Section>

        <Form.Section text={t('RPC 节点配置（可选）')}>
          <Row gutter={16}>
            {SUPPORTED_CHAINS.map((chain, index) => {
              const fieldName = RPC_CONFIG_FIELDS[chain.name];
              if (!fieldName) return null;
              
              return (
                <Col key={chain.id} xs={24} sm={24} md={12}>
                  <Form.Input
                    field={fieldName}
                    label={t(`${chain.label} RPC`)}
                    placeholder={chain.defaultRpcPlaceholder}
                    value={inputs[fieldName]}
                    onChange={(value) => handleInputChange(fieldName, value)}
                    extraText={
                      <Text type='tertiary' size='small'>
                        {chain.isRecommended ? t('推荐配置') : t('可选')} - {t(chain.description)}
                      </Text>
                    }
                    showClear
                  />
                </Col>
              );
            })}
          </Row>
        </Form.Section>

        <div
          style={{
            background: 'var(--semi-color-warning-light-default)',
            padding: '12px 16px',
            borderRadius: '8px',
            marginTop: '20px',
            marginBottom: '20px',
            border: '1px solid var(--semi-color-warning-light-active)',
          }}
        >
          <Text strong style={{ color: 'var(--semi-color-warning-dark)' }}>
            ⚠️ {t('安全提示')}
          </Text>
          <div style={{ marginTop: '8px' }}>
            <Text style={{ fontSize: '13px' }}>
              {t('• 请确保收款地址正确，并妥善保管私钥')}
            </Text>
            <br />
            <Text style={{ fontSize: '13px' }}>
              {t('• 建议定期将收到的 USDC 转移至冷钱包')}
            </Text>
            <br />
            <Text style={{ fontSize: '13px' }}>
              {t('• RPC 节点配置为可选，不填写将使用默认公共节点')}
            </Text>
            <br />
            <Text style={{ fontSize: '13px' }}>
              {t('• 保存后立即生效，无需重启服务')}
            </Text>
          </div>
        </div>

        <Space>
          <Button
            theme='solid'
            type='primary'
            htmlType='submit'
            onClick={onSubmit}
            loading={loading}
          >
            {t('保存 Web3 配置')}
          </Button>
        </Space>
      </Form>
    </>
  );
}
