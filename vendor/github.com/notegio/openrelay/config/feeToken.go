package config

import (
	"context"
	"encoding/hex"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
	orCommon "github.com/notegio/openrelay/common"
	"github.com/notegio/openrelay/exchangecontract"
	"github.com/notegio/openrelay/types"
	"gopkg.in/redis.v3"
	"log"
	"time"
)

type FeeToken interface {
	Get(order *types.Order) (*types.Address, error)
	Set(*types.Address) error
}

type staticFeeToken struct {
	value *types.Address
}

func (feeToken *staticFeeToken) Get(order *types.Order) (*types.Address, error) {
	return feeToken.value, nil
}

func (feeToken *staticFeeToken) Set(address *types.Address) error {
	feeToken.value = address
	return nil
}

type redisFeeToken struct {
	redisClient     *redis.Client
	cachedValue     *types.Address
	cacheExpiration int64
}

type rpcFeeToken struct {
	conn             bind.ContractBackend
	exchangeTokenMap map[types.Address]*types.Address
}

func (feeToken *rpcFeeToken) Get(order *types.Order) (*types.Address, error) {
	feeTokenAddress := &types.Address{}
	if feeTokenAddress, ok := feeToken.exchangeTokenMap[*order.ExchangeAddress]; ok {
		return feeTokenAddress, nil
	}
	exchange, err := exchangecontract.NewExchange(orCommon.ToGethAddress(order.ExchangeAddress), feeToken.conn)
	if err != nil {
		log.Printf("Error intializing exchange contract '%v': '%v'", hex.EncodeToString(order.ExchangeAddress[:]), err.Error())
		return feeTokenAddress, err
	}
	feeTokenGethAddress, err := exchange.ZRX_TOKEN_CONTRACT(nil)
	if err != nil {
		log.Printf("Error getting fee token address for exhange %#x", order.ExchangeAddress)
		return nil, err
	}
	copy(feeTokenAddress[:], feeTokenGethAddress[:])
	feeToken.exchangeTokenMap[*order.ExchangeAddress] = feeTokenAddress
	return feeTokenAddress, nil
}

func (feeToken *rpcFeeToken) Set(value *types.Address) error {
	// the rpcFeeToken looks up from the RPC server, so we can't actually set
	// the value.
	return nil
}

func (feeToken *redisFeeToken) Get(order *types.Order) (*types.Address, error) {
	if feeToken.cacheExpiration > time.Now().Unix() {
		// The fee token is unlikely to change, so caching it should be fine.
		// Doesn't hurt to check periodically just in case though.
		return feeToken.cachedValue, nil
	}
	result := &types.Address{}
	val, err := feeToken.redisClient.Get("feeToken::address").Result()
	if err != nil {
		return result, err
	}
	addressSlice, err := hex.DecodeString(val)
	if err != nil {
		return result, err
	}
	copy(result[:], addressSlice[:])
	return result, nil
}

func (feeToken *redisFeeToken) Set(value *types.Address) error {
	return feeToken.redisClient.Set("feeToken::address", hex.EncodeToString(value[:]), 0).Err()
}

func NewRpcFeeToken(rpcURL string) (FeeToken, error) {
	conn, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}
	if _, err = conn.SyncProgress(context.Background()); err != nil {
		// This is just here so that an NewRpcFeeToken can't be instantiated
		// successfully if the RPC server isn't responding properly. What RPC
		// function we call isn't important, but SyncProgress is pretty cheap.
		return nil, err
	}
	return &rpcFeeToken{conn, make(map[types.Address]*types.Address)}, nil
}

func NewFeeToken(client *redis.Client) FeeToken {
	return &redisFeeToken{client, &types.Address{}, 0}
}

func StaticFeeToken(address *types.Address) FeeToken {
	return &staticFeeToken{address}
}
