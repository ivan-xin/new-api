package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/operation_setting"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

const (
	PaymentMethodWeb3USDC = "web3_usdc"
	USDCDecimals          = 6 // USDC has 6 decimals
)

// USDC ABI (简化版，仅包含 Transfer 事件)
const USDCABI = `[
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "from", "type": "address"},
			{"indexed": true, "name": "to", "type": "address"},
			{"indexed": false, "name": "value", "type": "uint256"}
		],
		"name": "Transfer",
		"type": "event"
	}
]`

type Web3PayRequest struct {
	Amount        int64  `json:"amount"`
	PaymentMethod string `json:"payment_method"`
	Chain         string `json:"chain"` // ethereum, base, bsc, arbitrum
}

type Web3PayInfo struct {
	Chain            string `json:"chain"`
	USDCContract     string `json:"usdc_contract"`
	ReceiverAddress  string `json:"receiver_address"`
	RequiredAmount   string `json:"required_amount"`
	AmountUSDC       string `json:"amount_usdc"`
	MinConfirmations int    `json:"min_confirmations"`
}

// GetWeb3PayInfo 获取 Web3 支付信息
func GetWeb3PayInfo(c *gin.Context) {
	chain := c.DefaultQuery("chain", "base") // 默认使用 Base
	amount := c.Query("amount")

	if amount == "" {
		c.JSON(200, gin.H{"message": "error", "data": "请提供充值金额"})
		return
	}

	amountInt, err := strconv.ParseInt(amount, 10, 64)
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "金额格式错误"})
		return
	}

	// 获取用户分组
	id := c.GetInt("id")
	group, err := model.GetUserGroup(id, true)
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "获取用户分组失败"})
		return
	}

	// 计算 USDC 金额
	payMoney := getWeb3PayMoney(float64(amountInt), group)

	// USDC 金额 (6 decimals)
	usdcAmount := fmt.Sprintf("%.6f", payMoney)

	// 从环境变量获取收款地址
	receiverAddress := getWeb3ReceiverAddress()
	if receiverAddress == "" {
		c.JSON(200, gin.H{"message": "error", "data": "管理员未配置Web3收款地址"})
		return
	}

	// 获取 USDC 合约地址
	usdcContract := common.GetUSDCContract(chain)

	info := Web3PayInfo{
		Chain:            chain,
		USDCContract:     usdcContract,
		ReceiverAddress:  receiverAddress,
		RequiredAmount:   amount,
		AmountUSDC:       usdcAmount,
		MinConfirmations: 3,
	}

	c.JSON(200, gin.H{"message": "success", "data": info})
}

// RequestWeb3Pay 创建 Web3 支付订单
func RequestWeb3Pay(c *gin.Context) {
	var req Web3PayRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "参数错误"})
		return
	}

	if req.Amount < getWeb3MinTopup() {
		c.JSON(200, gin.H{"message": "error", "data": fmt.Sprintf("充值数量不能小于 %d", getWeb3MinTopup())})
		return
	}

	// 验证链是否支持
	if !common.IsChainSupported(req.Chain) {
		c.JSON(200, gin.H{"message": "error", "data": "不支持的区块链网络"})
		return
	}

	// 检查是否为测试网（只在测试网未启用时拒绝）
	if common.IsTestnetChain(req.Chain) && !common.IsTestnetEnabled() {
		c.JSON(200, gin.H{"message": "error", "data": "生产环境禁止使用测试网充值，请切换到主网"})
		return
	}

	id := c.GetInt("id")
	group, err := model.GetUserGroup(id, true)
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "获取用户分组失败"})
		return
	}

	payMoney := getWeb3PayMoney(float64(req.Amount), group)

	// 获取收款地址
	receiverAddress := getWeb3ReceiverAddress()
	if receiverAddress == "" {
		c.JSON(200, gin.H{"message": "error", "data": "管理员未配置Web3收款地址"})
		return
	}

	// 创建订单
	tradeNo := fmt.Sprintf("WEB3%dNO%s%d", id, common.GetRandomString(6), time.Now().Unix())

	metadata := map[string]string{
		"chain":   req.Chain,
		"tx_hash": "",
	}
	metadataJSON, _ := json.Marshal(metadata)

	topUp := &model.TopUp{
		UserId:        id,
		Amount:        req.Amount,
		Money:         payMoney,
		TradeNo:       tradeNo,
		PaymentMethod: PaymentMethodWeb3USDC,
		CreateTime:    time.Now().Unix(),
		Status:        "pending",
	}

	err = topUp.Insert()
	if err != nil {
		log.Printf("创建Web3订单失败: %v", err)
		c.JSON(200, gin.H{"message": "error", "data": "创建订单失败"})
		return
	}

	log.Printf("Web3订单创建成功: %s, 用户: %d, 金额: %.6f USDC", tradeNo, id, payMoney)

	c.JSON(200, gin.H{
		"message": "success",
		"data": gin.H{
			"trade_no":         tradeNo,
			"chain":            req.Chain,
			"usdc_contract":    common.GetUSDCContract(req.Chain),
			"receiver_address": receiverAddress,
			"amount_usdc":      fmt.Sprintf("%.6f", payMoney),
			"metadata":         string(metadataJSON),
		},
	})
}

// VerifyWeb3Transaction 验证 Web3 交易
func VerifyWeb3Transaction(c *gin.Context) {
	var req struct {
		TradeNo string `json:"trade_no"`
		TxHash  string `json:"tx_hash"`
		Chain   string `json:"chain"`
	}

	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "参数错误"})
		return
	}

	// 获取订单
	LockOrder(req.TradeNo)
	defer UnlockOrder(req.TradeNo)

	topUp := model.GetTopUpByTradeNo(req.TradeNo)
	if topUp == nil {
		c.JSON(200, gin.H{"message": "error", "data": "订单不存在"})
		return
	}

	if topUp.Status != "pending" {
		c.JSON(200, gin.H{"message": "error", "data": "订单状态错误"})
		return
	}

	// 验证交易
	rpcURL := getWeb3RPCURL(req.Chain)
	if rpcURL == "" {
		c.JSON(200, gin.H{"message": "error", "data": "未配置区块链节点"})
		return
	}

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		log.Printf("连接区块链节点失败: %v", err)
		c.JSON(200, gin.H{"message": "error", "data": "连接区块链失败"})
		return
	}
	defer client.Close()

	// 获取交易详情
	txHash := ethcommon.HexToHash(req.TxHash)
	_, isPending, err := client.TransactionByHash(context.Background(), txHash)
	if err != nil {
		log.Printf("获取交易失败: %v", err)
		c.JSON(200, gin.H{"message": "error", "data": "获取交易失败"})
		return
	}

	if isPending {
		c.JSON(200, gin.H{"message": "pending", "data": "交易确认中"})
		return
	}

	// 获取交易收据
	receipt, err := client.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		log.Printf("获取交易收据失败: %v", err)
		c.JSON(200, gin.H{"message": "error", "data": "获取交易收据失败"})
		return
	}

	// 检查交易是否成功
	if receipt.Status != types.ReceiptStatusSuccessful {
		c.JSON(200, gin.H{"message": "error", "data": "交易失败"})
		return
	}

	// 获取收款地址
	receiverAddress := getWeb3ReceiverAddress()
	if receiverAddress == "" {
		c.JSON(200, gin.H{"message": "error", "data": "管理员未配置收款地址"})
		return
	}

	// 获取 USDC 合约地址
	usdcContract := common.GetUSDCContract(req.Chain)
	if usdcContract == "" {
		c.JSON(200, gin.H{"message": "error", "data": "无法获取该链的USDC合约地址"})
		return
	}

	// 反作弊 1：防止复用旧交易 - 校验区块时间应不早于订单创建时间
	if receipt.BlockNumber != nil {
		blk, err := client.BlockByNumber(context.Background(), receipt.BlockNumber)
		if err == nil {
			blkTime := int64(blk.Time())
			if blkTime < topUp.CreateTime {
				c.JSON(200, gin.H{"message": "error", "data": "交易早于订单创建，疑似复用旧交易"})
				return
			}
		}
	}

	// 反作弊 2：防止重复使用同一交易哈希
	if existed := model.GetTopUpByTxHash(req.TxHash); existed != nil && existed.TradeNo != req.TradeNo {
		c.JSON(200, gin.H{"message": "error", "data": "该交易哈希已被用于其他订单，无法重复入账"})
		return
	}

	// 验证交易金额和接收地址
	verified, amount, err := verifyUSDCTransfer(receipt, receiverAddress, usdcContract)
	if err != nil {
		log.Printf("验证USDC转账失败: %v", err)
		c.JSON(200, gin.H{"message": "error", "data": err.Error()})
		return
	}

	if !verified {
		c.JSON(200, gin.H{"message": "error", "data": "交易验证失败"})
		return
	}

	// 从合约查询实际的小数位数
	contractDecimals, err := getContractDecimals(client, usdcContract)
	if err != nil {
		log.Printf("获取合约小数位数失败: %v, 使用默认值 %d", err, USDCDecimals)
		contractDecimals = USDCDecimals
	}

	// 检查金额是否匹配
	expectedAmount := decimal.NewFromFloat(topUp.Money)
	actualAmount := decimal.NewFromBigInt(amount, -int32(contractDecimals))

	log.Printf("Web3支付验证: 合约小数位: %d, 期望 %.6f USDC, 实际 %.6f USDC",
		contractDecimals, expectedAmount.InexactFloat64(), actualAmount.InexactFloat64())

	// 更严格：金额必须与订单一致，防止复用更早/更大的旧交易
	if !actualAmount.Equal(expectedAmount) {
		c.JSON(200, gin.H{"message": "error", "data": "支付金额与订单金额不一致，请按订单金额付款"})
		return
	}

	// 更新订单状态并记录链上信息
	topUp.Status = "success"
	topUp.TxHash = &req.TxHash
	topUp.Chain = &req.Chain
	topUp.CompleteTime = time.Now().Unix()
	metadata := map[string]string{
		"chain":   req.Chain,
		"tx_hash": req.TxHash,
	}
	metadataJSON, _ := json.Marshal(metadata)
	_ = metadataJSON // 如果模型有 metadata 字段可以使用

	err = topUp.Update()
	if err != nil {
		log.Printf("更新Web3订单失败: %v", topUp)
		// 可能是 tx_hash 唯一索引冲突
		c.JSON(200, gin.H{"message": "error", "data": "更新订单失败，可能为重复交易哈希"})
		return
	}

	// 增加用户额度
	dAmount := decimal.NewFromInt(topUp.Amount)
	dQuotaPerUnit := decimal.NewFromFloat(common.QuotaPerUnit)
	quotaToAdd := int(dAmount.Mul(dQuotaPerUnit).IntPart())

	err = model.IncreaseUserQuota(topUp.UserId, quotaToAdd, true)
	if err != nil {
		log.Printf("Web3支付增加用户额度失败: %v", topUp)
		c.JSON(200, gin.H{"message": "error", "data": "充值失败"})
		return
	}

	model.RecordLog(topUp.UserId, model.LogTypeTopup,
		fmt.Sprintf("使用Web3 USDC充值成功，链: %s，交易: %s，充值金额: %.6f USDC，获得额度：%s",
			req.Chain, req.TxHash, actualAmount.InexactFloat64(), logger.LogQuota(quotaToAdd)))

	log.Printf("Web3充值成功: 用户 %d, 订单 %s, 金额 %.6f USDC, 额度 %d", topUp.UserId, req.TradeNo, actualAmount.InexactFloat64(), quotaToAdd)

	c.JSON(200, gin.H{"message": "success", "data": "充值成功"})
}

// verifyUSDCTransfer 验证 USDC 转账
func verifyUSDCTransfer(receipt *types.Receipt, receiverAddress string, usdcContract string) (bool, *big.Int, error) {
	contractABI, err := abi.JSON(strings.NewReader(USDCABI))
	if err != nil {
		return false, nil, err
	}

	// 查找 Transfer 事件
	transferEventID := contractABI.Events["Transfer"].ID

	for _, vLog := range receipt.Logs {
		// 检查合约地址
		if !strings.EqualFold(vLog.Address.Hex(), usdcContract) {
			continue
		}

		// 检查是否是 Transfer 事件
		if len(vLog.Topics) < 3 {
			continue
		}

		if vLog.Topics[0] != transferEventID {
			continue
		}

		// 解析 Transfer 事件
		// Topics[0] = Event ID
		// Topics[1] = from address
		// Topics[2] = to address
		// Data = value

		toAddress := ethcommon.BytesToAddress(vLog.Topics[2].Bytes())

		// 验证接收地址
		if strings.EqualFold(toAddress.Hex(), receiverAddress) {
			// 解析金额
			value := new(big.Int).SetBytes(vLog.Data)
			return true, value, nil
		}
	}

	return false, nil, fmt.Errorf("未找到有效的USDC转账事件")
}

func getWeb3PayMoney(amount float64, group string) float64 {
	originalAmount := amount
	if operation_setting.GetQuotaDisplayType() == operation_setting.QuotaDisplayTypeTokens {
		amount = amount / common.QuotaPerUnit
	}

	topupGroupRatio := common.GetTopupGroupRatio(group)
	if topupGroupRatio == 0 {
		topupGroupRatio = 1
	}

	discount := 1.0
	if ds, ok := operation_setting.GetPaymentSetting().AmountDiscount[int(originalAmount)]; ok {
		if ds > 0 {
			discount = ds
		}
	}

	// USDC 价格 (默认 1 USDC = 1 USD，可以从环境变量读取)
	usdcPrice := 1.0

	payMoney := amount * usdcPrice * topupGroupRatio * discount
	return payMoney
}

func getWeb3MinTopup() int64 {
	// 默认最小充值 1 USDC，可通过配置项 Web3MinTopup 自定义（默认建议 10 USDC）
	// 可以通过配置项 Web3MinTopup 自定义
	minTopup := 1

	// 从配置读取自定义的最小充值金额
	if minTopupStr, ok := common.OptionMap["Web3MinTopup"]; ok && minTopupStr != "" {
		if customMinTopup, err := strconv.Atoi(minTopupStr); err == nil && customMinTopup >= 1 {
			minTopup = customMinTopup
		}
	}

	if operation_setting.GetQuotaDisplayType() == operation_setting.QuotaDisplayTypeTokens {
		minTopup = minTopup * int(common.QuotaPerUnit)
	}
	return int64(minTopup)
}

// 辅助函数：获取 Web3 配置
func getWeb3ReceiverAddress() string {
	// 从选项表读取
	if addr, ok := common.OptionMap["Web3ReceiverAddress"]; ok && addr != "" {
		return addr
	}
	return ""
}

func getWeb3RPCURL(chain string) string {
	// 从选项表读取配置的 RPC URL
	rpcKey := common.GetRPCConfigKey(chain)
	if rpcKey != "" {
		if rpc, ok := common.OptionMap[rpcKey]; ok && rpc != "" {
			return rpc
		}
	}

	// 如果没有配置,则使用配置文件中的默认公共 RPC
	chainConfig := common.GetChainConfig(chain)
	if chainConfig != nil {
		return chainConfig.RpcURL
	}

	return ""
}
