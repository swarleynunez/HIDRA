package eth

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/swarleynunez/superfog/core/utils"
	"math/big"
)

func SendTxn(c *ethclient.Client, ks *keystore.KeyStore, to common.Address, passphrase string, value int64) (tx *types.Transaction) {

	// Selects the sender account
	if len(ks.Accounts()) == 0 {
		err := errors.New("eth: no keystore accounts loaded")
		utils.CheckError(err, utils.FatalMode)
	}
	from := ks.Accounts()[0]

	// Get the next nonce (sender)
	nonce, err := c.PendingNonceAt(context.Background(), from.Address)
	utils.CheckError(err, utils.WarningMode)

	// Gas price
	gasPrice, err := c.SuggestGasPrice(context.Background())
	utils.CheckError(err, utils.WarningMode)

	// Call needed fields
	msg := ethereum.CallMsg{
		From:     from.Address,
		To:       &to,
		Gas:      21000,
		GasPrice: gasPrice,
		Value:    big.NewInt(value),
		Data:     nil,
	}

	// Gas limit
	gasLimit, err := c.EstimateGas(context.Background(), msg)
	utils.CheckError(err, utils.WarningMode)

	// Create raw transaction
	tx = types.NewTransaction(nonce, *msg.To, msg.Value, gasLimit, msg.GasPrice, msg.Data)

	// Get ChainId for transaction replay protection
	chainId, err := c.ChainID(context.Background())
	utils.CheckError(err, utils.WarningMode)

	// Signs the raw transaction
	tx, err = ks.SignTxWithPassphrase(from, passphrase, tx, chainId)
	utils.CheckError(err, utils.WarningMode)

	// Send the transaction
	err = c.SendTransaction(context.Background(), tx)
	utils.CheckError(err, utils.WarningMode)

	return
}
