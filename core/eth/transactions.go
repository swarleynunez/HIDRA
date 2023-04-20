package eth

import (
	"context"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/swarleynunez/hidra/core/utils"
	"math/big"
	"strconv"
)

func Transactor(ctx context.Context, ethc *ethclient.Client, ks *keystore.KeyStore, from accounts.Account, gasLimit uint64) *bind.TransactOpts {

	// Get and parse chain ID (transaction replay protection)
	chainId, err := strconv.ParseUint(utils.GetEnv("CHAIN_ID"), 10, 64)
	utils.CheckError(err, utils.FatalMode)

	// Auth transactor type
	auth, err := bind.NewKeyStoreTransactorWithChainID(ks, from, big.NewInt(int64(chainId)))
	utils.CheckError(err, utils.FatalMode)

	// Set nonce
	nonce, err := ethc.PendingNonceAt(ctx, from.Address) // Get loaded Ethereum account current nonce
	utils.CheckError(err, utils.FatalMode)
	auth.Nonce = big.NewInt(int64(nonce))

	// Set gas limit
	auth.GasLimit = gasLimit

	return auth
}

func SignedEtherTransaction(ctx context.Context, ethc *ethclient.Client, ks *keystore.KeyStore, from accounts.Account, passphrase string, to common.Address, value int64) *types.Transaction {

	// Set nonce
	nonce, err := ethc.PendingNonceAt(ctx, from.Address) // Get loaded Ethereum account current nonce
	utils.CheckError(err, utils.FatalMode)

	// Suggest gas price
	gasPrice, err := ethc.SuggestGasPrice(ctx)
	utils.CheckError(err, utils.FatalMode)

	// Create transaction
	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		GasPrice: gasPrice,
		Gas:      21000,
		To:       &to,
		Value:    big.NewInt(value),
	})

	// Get and parse chain ID (transaction replay protection)
	chainId, err := strconv.ParseUint(utils.GetEnv("CHAIN_ID"), 10, 64)
	utils.CheckError(err, utils.FatalMode)

	// Sign transaction
	stx, err := ks.SignTxWithPassphrase(from, passphrase, tx, big.NewInt(int64(chainId)))
	utils.CheckError(err, utils.FatalMode)

	return stx
}
