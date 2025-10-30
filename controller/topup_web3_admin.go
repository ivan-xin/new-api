package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// AdminWeb3PayRequest 管理员为指定用户创建Web3充值订单的请求
type AdminWeb3PayRequest struct {
	UserId        int    `json:"user_id"`                           // 目标用户ID（user_id 和 username 二选一）
	Username      string `json:"username"`                          // 目标用户名（user_id 和 username 二选一）
	Amount        int64  `json:"amount" binding:"required"`         // 充值额度
	PaymentMethod string `json:"payment_method" binding:"required"` // 支付方式
	Chain         string `json:"chain" binding:"required"`          // 区块链网络
}

// AdminWeb3VerifyRequest 管理员验证Web3交易的请求
type AdminWeb3VerifyRequest struct {
	TradeNo string `json:"trade_no" binding:"required"` // 订单号
	TxHash  string `json:"tx_hash" binding:"required"`  // 交易哈希
	Chain   string `json:"chain" binding:"required"`    // 区块链网络
}

// AdminRequestWeb3Pay 管理员为指定用户创建 Web3 充值订单
func AdminRequestWeb3Pay(c *gin.Context) {
	var req AdminWeb3PayRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "参数错误: " + err.Error()})
		return
	}

	// 验证必须提供 user_id 或 username 之一
	if req.UserId == 0 && req.Username == "" {
		c.JSON(200, gin.H{"message": "error", "data": "必须提供 user_id 或 username 参数"})
		return
	}

	var targetUser model.User

	// 优先使用 user_id，如果没有则使用 username 查询
	if req.UserId != 0 {
		err = model.DB.Where("id = ?", req.UserId).First(&targetUser).Error
		if err != nil {
			c.JSON(200, gin.H{"message": "error", "data": fmt.Sprintf("用户ID %d 不存在", req.UserId)})
			return
		}
	} else {
		// 通过用户名查询用户
		err = model.DB.Where("username = ?", req.Username).First(&targetUser).Error
		if err != nil {
			c.JSON(200, gin.H{"message": "error", "data": fmt.Sprintf("用户名 '%s' 不存在", req.Username)})
			return
		}
		// 使用查询到的用户ID
		req.UserId = targetUser.Id
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

	// 获取用户分组
	group, err := model.GetUserGroup(req.UserId, true)
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
	tradeNo := fmt.Sprintf("WEB3%dNO%s%d", req.UserId, common.GetRandomString(6), time.Now().Unix())

	metadata := map[string]string{
		"chain":   req.Chain,
		"tx_hash": "",
	}
	metadataJSON, _ := json.Marshal(metadata)

	topUp := &model.TopUp{
		UserId:        req.UserId,
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

	log.Printf("管理员为用户创建Web3订单: %s, 用户: %d (%s), 金额: %.6f USDC",
		tradeNo, req.UserId, targetUser.Username, payMoney)

	c.JSON(200, gin.H{
		"message": "success",
		"data": gin.H{
			"trade_no":         tradeNo,
			"user_id":          req.UserId,
			"username":         targetUser.Username,
			"chain":            req.Chain,
			"usdc_contract":    common.GetUSDCContract(req.Chain),
			"receiver_address": receiverAddress,
			"amount_usdc":      fmt.Sprintf("%.6f", payMoney),
			"metadata":         string(metadataJSON),
		},
	})
}

// AdminVerifyWeb3Transaction 管理员验证 Web3 交易
func AdminVerifyWeb3Transaction(c *gin.Context) {
	var req AdminWeb3VerifyRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "参数错误: " + err.Error()})
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
		c.JSON(200, gin.H{"message": "error", "data": "订单状态错误，当前状态: " + topUp.Status})
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
		c.JSON(200, gin.H{"message": "error", "data": "获取交易失败: " + err.Error()})
		return
	}

	if isPending {
		c.JSON(200, gin.H{"message": "pending", "data": "交易确认中，请稍后再试"})
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
		c.JSON(200, gin.H{"message": "error", "data": "交易失败，请检查链上状态"})
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

	// 验证交易金额和接收地址
	verified, amount, err := verifyUSDCTransfer(receipt, receiverAddress, usdcContract)
	if err != nil {
		log.Printf("验证USDC转账失败: %v", err)
		c.JSON(200, gin.H{"message": "error", "data": err.Error()})
		return
	}

	if !verified {
		c.JSON(200, gin.H{"message": "error", "data": "交易验证失败，未找到有效的USDC转账"})
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

	log.Printf("Web3支付验证: 订单 %s, 合约小数位: %d, 期望 %.6f USDC, 实际 %.6f USDC",
		req.TradeNo, contractDecimals, expectedAmount.InexactFloat64(), actualAmount.InexactFloat64())

	log.Printf("Web3支付验证: 订单 %s, 期望 %.6f USDC, 实际 %.6f USDC",
		req.TradeNo, expectedAmount.InexactFloat64(), actualAmount.InexactFloat64())

	if actualAmount.LessThan(expectedAmount) {
		c.JSON(200, gin.H{"message": "error", "data": fmt.Sprintf("支付金额不足，需要 %.6f USDC，实际 %.6f USDC",
			expectedAmount.InexactFloat64(), actualAmount.InexactFloat64())})
		return
	}

	// 更新订单状态
	topUp.Status = "success"
	metadata := map[string]string{
		"chain":   req.Chain,
		"tx_hash": req.TxHash,
	}
	metadataJSON, _ := json.Marshal(metadata)
	_ = metadataJSON

	err = topUp.Update()
	if err != nil {
		log.Printf("更新Web3订单失败: %v", err)
		c.JSON(200, gin.H{"message": "error", "data": "更新订单失败"})
		return
	}

	// 增加用户额度
	dAmount := decimal.NewFromInt(topUp.Amount)
	dQuotaPerUnit := decimal.NewFromFloat(common.QuotaPerUnit)
	quotaToAdd := int(dAmount.Mul(dQuotaPerUnit).IntPart())

	err = model.IncreaseUserQuota(topUp.UserId, quotaToAdd, true)
	if err != nil {
		log.Printf("Web3支付增加用户额度失败: 用户ID %d, %v", topUp.UserId, err)
		c.JSON(200, gin.H{"message": "error", "data": "充值失败，增加额度时出错"})
		return
	}

	// 获取用户信息用于日志
	user, _ := model.GetUserById(topUp.UserId, false)
	username := "unknown"
	if user != nil {
		username = user.Username
	}

	model.RecordLog(topUp.UserId, model.LogTypeTopup,
		fmt.Sprintf("管理员为用户充值(Web3 USDC)，链: %s，交易: %s，充值金额: %.6f USDC，获得额度：%s",
			req.Chain, req.TxHash, actualAmount.InexactFloat64(), logger.LogQuota(quotaToAdd)))

	log.Printf("Web3充值成功: 用户 %d (%s), 订单 %s, 金额 %.6f USDC, 额度 %d",
		topUp.UserId, username, req.TradeNo, actualAmount.InexactFloat64(), quotaToAdd)

	c.JSON(200, gin.H{
		"message": "success",
		"data": gin.H{
			"trade_no":    req.TradeNo,
			"user_id":     topUp.UserId,
			"username":    username,
			"quota":       quotaToAdd,
			"amount_usdc": fmt.Sprintf("%.6f", actualAmount.InexactFloat64()),
		},
	})
}

// getContractDecimals 从合约获取小数位数
func getContractDecimals(client *ethclient.Client, contractAddress string) (int, error) {
	// ERC20 decimals() 方法的 ABI
	decimalsABI := `[{"constant":true,"inputs":[],"name":"decimals","outputs":[{"name":"","type":"uint8"}],"type":"function"}]`

	contractABIParsed, err := abi.JSON(strings.NewReader(decimalsABI))
	if err != nil {
		return 0, fmt.Errorf("解析 decimals ABI 失败: %v", err)
	}

	// 编码 decimals() 调用
	data, err := contractABIParsed.Pack("decimals")
	if err != nil {
		return 0, fmt.Errorf("编码 decimals 调用失败: %v", err)
	}

	// 调用合约
	contractAddr := ethcommon.HexToAddress(contractAddress)

	msg := ethereum.CallMsg{
		To:   &contractAddr,
		Data: data,
	}

	result, err := client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return 0, fmt.Errorf("调用合约 decimals 失败: %v", err)
	}

	// 解析返回值
	var decimals uint8
	err = contractABIParsed.UnpackIntoInterface(&decimals, "decimals", result)
	if err != nil {
		return 0, fmt.Errorf("解析 decimals 返回值失败: %v", err)
	}

	return int(decimals), nil
}

// verifyUSDCTransfer 验证 USDC 转账（内部辅助函数）
func verifyUSDCTransferInternal(receipt *types.Receipt, receiverAddress string, usdcContract string) (bool, *big.Int, error) {
	contractABI, err := abi.JSON(strings.NewReader(USDCABI))
	if err != nil {
		return false, nil, err
	}

	transferEventID := contractABI.Events["Transfer"].ID

	for _, vLog := range receipt.Logs {
		if !strings.EqualFold(vLog.Address.Hex(), usdcContract) {
			continue
		}

		if len(vLog.Topics) < 3 {
			continue
		}

		if vLog.Topics[0] != transferEventID {
			continue
		}

		toAddress := ethcommon.BytesToAddress(vLog.Topics[2].Bytes())

		if strings.EqualFold(toAddress.Hex(), receiverAddress) {
			value := new(big.Int).SetBytes(vLog.Data)
			return true, value, nil
		}
	}

	return false, nil, fmt.Errorf("未找到有效的USDC转账事件")
}
