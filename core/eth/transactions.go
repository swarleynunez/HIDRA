package eth

import (
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/swarleynunez/superfog/core/utils"
	"math/big"
	"os"
	"strconv"
)

/*func SendEther(ethc *ethclient.Client, from accounts.Account, to common.Address, value *big.Int, ks *keystore.KeyStore, passphrase string) {

	// Get the next nonce (sender)
	nonce, err := ethc.PendingNonceAt(context.Background(), from.Address)
	utils.CheckError(err, utils.WarningMode)

	// Gas price
	gasPrice, err := ethc.SuggestGasPrice(context.Background())
	utils.CheckError(err, utils.WarningMode)

	// Create raw transaction
	tx := types.NewTransaction(nonce, to, value, 21000, gasPrice, nil)

	// Get ChainId for transaction replay protection
	chainId, err := ethc.ChainID(context.Background())
	utils.CheckError(err, utils.WarningMode)

	// Signs the raw transaction
	tx, err = ks.SignTxWithPassphrase(from, passphrase, tx, chainId)
	utils.CheckError(err, utils.WarningMode)

	// Send the transaction
	err = ethc.SendTransaction(context.Background(), tx)
	utils.CheckError(err, utils.WarningMode)
}*/

func GetTransactor(ks *keystore.KeyStore, from accounts.Account, nonce, gasLimit uint64) *bind.TransactOpts {

	// Get and parse chain ID (transaction replay protection)
	chainId, err := strconv.ParseUint(os.Getenv("CHAIN_ID"), 10, 64)
	utils.CheckError(err, utils.WarningMode)

	// Auth transactor type
	auth, err := bind.NewKeyStoreTransactorWithChainID(ks, from, big.NewInt(int64(chainId)))
	utils.CheckError(err, utils.WarningMode)

	// Set nonce
	auth.Nonce = big.NewInt(int64(nonce))

	// Set gas limit
	auth.GasLimit = gasLimit

	return auth
}
