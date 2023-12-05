package main

import (
	"crypto/ecdsa"
	"log"
	"os"

	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/spf13/cobra"

	retry "github.com/OffchainLabs/bold/runtime"
	"github.com/OffchainLabs/bold/solgen/go/mocksgen"
	"github.com/OffchainLabs/bold/solgen/go/rollupgen"
	challenge_testing "github.com/OffchainLabs/bold/testing"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
)

// Define the flags as global variables
var (
	valPrivKeys       string
	l1ChainIdStr      string
	l1EndpointUrl     string
	rollupAddrStr     string
	stakeTokenAddrStr string
	inboxAddrStr      string
	gweiToDeposit     uint64
)

var rootCmd = &cobra.Command{
	Use:   "bold-utils",
	Short: "Arbitrum BOLD command-line utilities",
}

var mintStakeTokenCmd = &cobra.Command{
	Use:   "mint-stake-token",
	Short: "Mint stake token description",
	Run: func(cmd *cobra.Command, args []string) {
		mintStakeToken()
	},
}

var bridgeEthCmd = &cobra.Command{
	Use:   "bridge-eth",
	Short: "Bridge Ethereum description",
	Run: func(cmd *cobra.Command, args []string) {
		bridgeEth()
	},
}

func init() {
	// Bind flags for mint stake token
	mintStakeTokenCmd.Flags().StringVarP(&valPrivKeys, "validator-priv-keys", "", "", "comma-separated validator private keys")
	mintStakeTokenCmd.Flags().StringVarP(&l1ChainIdStr, "l1-chain-id", "", "11155111", "l1 chain id (sepolia default)")
	mintStakeTokenCmd.Flags().StringVarP(&l1EndpointUrl, "l1-endpoint", "", "", "l1 endpoint")
	mintStakeTokenCmd.Flags().StringVarP(&rollupAddrStr, "rollup-address", "", "", "rollup address")
	mintStakeTokenCmd.Flags().StringVarP(&stakeTokenAddrStr, "stake-token-address", "", "", "stake token address")
	mintStakeTokenCmd.Flags().Uint64VarP(&gweiToDeposit, "gwei-to-deposit", "", 10000, "eth to deposit into tokens, in gwei")

	// Bind flags for bridge eth
	bridgeEthCmd.Flags().StringVarP(&valPrivKeys, "validator-priv-keys", "", "", "comma-separated validator private keys")
	bridgeEthCmd.Flags().StringVarP(&l1ChainIdStr, "l1-chain-id", "", "11155111", "l1 chain id (sepolia default)")
	bridgeEthCmd.Flags().StringVarP(&l1EndpointUrl, "l1-endpoint", "", "", "l1 endpoint")
	bridgeEthCmd.Flags().StringVarP(&inboxAddrStr, "inbox-address", "", "", "inbox address")
	bridgeEthCmd.Flags().Uint64VarP(&gweiToDeposit, "gwei-to-deposit", "", 2000000, "eth to bridge over, in gwei (2M default, or 0.002 ETH)")
}

func main() {
	rootCmd.AddCommand(mintStakeTokenCmd)
	rootCmd.AddCommand(bridgeEthCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func mintStakeToken() {
	// Your existing logic goes here
	// You can use the flag variables like valPrivKeys, l1ChainIdStr, etc., directly here
	ctx := context.Background()
	endpoint, err := rpc.DialContext(ctx, l1EndpointUrl)
	if err != nil {
		panic(err)
	}
	client := ethclient.NewClient(endpoint)
	l1ChainId, ok := new(big.Int).SetString(l1ChainIdStr, 10)
	if !ok {
		panic("not big int")
	}
	if valPrivKeys == "" {
		panic("no validator private keys set")
	}
	privKeyStrings := strings.Split(valPrivKeys, ",")
	for _, privKeyStr := range privKeyStrings {
		validatorPrivateKey, err := crypto.HexToECDSA(privKeyStr)
		if err != nil {
			panic(err)
		}
		txOpts, err := bind.NewKeyedTransactorWithChainID(validatorPrivateKey, l1ChainId)
		if err != nil {
			panic(err)
		}

		rollupAddr := common.HexToAddress(rollupAddrStr)
		rollupBindings, err := rollupgen.NewRollupUserLogicCaller(rollupAddr, client)
		if err != nil {
			panic(err)
		}
		chalManagerAddr, err := rollupBindings.ChallengeManager(&bind.CallOpts{})
		if err != nil {
			panic(err)
		}
		stakeTokenAddr := common.HexToAddress(stakeTokenAddrStr)

		tokenBindings, err := mocksgen.NewTestWETH9(stakeTokenAddr, client)
		if err != nil {
			panic(err)
		}
		allow, err := tokenBindings.Allowance(&bind.CallOpts{}, txOpts.From, rollupAddr)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Addr %#x gave rollup %#x allowance of %#x\n", txOpts.From, rollupAddr, allow.Bytes())

		allow, err = tokenBindings.Allowance(&bind.CallOpts{}, txOpts.From, chalManagerAddr)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Addr %#x gave chal manager addr %#x allowance of %#x\n", txOpts.From, chalManagerAddr, allow.Bytes())

		depositAmount := new(big.Int).SetUint64(gweiToDeposit * params.GWei)
		txOpts.Value = depositAmount
		if _, err = retry.UntilSucceeds[bool](ctx, func() (bool, error) {
			tx, err2 := tokenBindings.Deposit(txOpts)
			if err2 != nil {
				return false, err2
			}
			if err2 = challenge_testing.WaitForTx(ctx, client, tx); err2 != nil {
				return false, err2
			}
			return true, nil
		}); err != nil {
			panic(err)
		}
		txOpts.Value = big.NewInt(0)
		maxUint256 := new(big.Int)
		maxUint256.Exp(big.NewInt(2), big.NewInt(256), nil).Sub(maxUint256, big.NewInt(1))
		if _, err = retry.UntilSucceeds[bool](ctx, func() (bool, error) {
			tx, err2 := tokenBindings.Approve(txOpts, rollupAddr, maxUint256)
			if err2 != nil {
				return false, err2
			}
			if err2 = challenge_testing.WaitForTx(ctx, client, tx); err2 != nil {
				return false, err2
			}
			return true, nil
		}); err != nil {
			panic(err)
		}
		if _, err = retry.UntilSucceeds[bool](ctx, func() (bool, error) {
			tx, err2 := tokenBindings.Approve(txOpts, chalManagerAddr, maxUint256)
			if err2 != nil {
				return false, err2
			}
			if err2 = challenge_testing.WaitForTx(ctx, client, tx); err2 != nil {
				return false, err2
			}
			return true, nil
		}); err != nil {
			panic(err)
		}

	}
}

func bridgeEth() {
	ctx := context.Background()
	l1ChainId, ok := new(big.Int).SetString(l1ChainIdStr, 10)
	if !ok {
		panic("not big int")
	}
	client, err := ethclient.Dial(l1EndpointUrl)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}
	privKeyStrings := strings.Split(valPrivKeys, ",")
	toAddress := common.HexToAddress(inboxAddrStr)
	data := common.Hex2Bytes("0f4d14e9000000000000000000000000000000000000000000000000000082f79cd90000")
	depositAmount := new(big.Int).SetUint64(gweiToDeposit * params.GWei)
	gasLimit := uint64(150000)
	for _, privKeyStr := range privKeyStrings {
		privateKey, err := crypto.HexToECDSA(privKeyStr)
		if err != nil {
			log.Fatalf("Failed to create private key: %v", err)
		}
		publicKey := privateKey.Public()
		publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
		if !ok {
			log.Fatal("Error casting public key to ECDSA")
		}
		fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
		nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
		if err != nil {
			log.Fatalf("Failed to get nonce: %v", err)
		}
		gasPrice, err := client.SuggestGasPrice(context.Background())
		if err != nil {
			log.Fatalf("Failed to suggest gas price: %v", err)
		}
		tx := types.NewTransaction(nonce, toAddress, depositAmount, gasLimit, gasPrice, data)
		signer := types.NewCancunSigner(l1ChainId)
		signedTx, err := types.SignTx(tx, signer, privateKey)
		if err != nil {
			log.Fatalf("Failed to sign transaction: %v", err)
		}
		err = client.SendTransaction(ctx, signedTx)
		if err != nil {
			log.Fatalf("Failed to send transaction: %v", err)
		}
		fmt.Printf("Sent transaction: %s", signedTx.Hash().Hex())
	}
}
