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

	"github.com/OffchainLabs/bold/solgen/go/mocksgen"
	"github.com/OffchainLabs/bold/solgen/go/rollupgen"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
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
	weiToDeposit      string
	weiToMint         string
	bumpPricePercent  int64
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
	mintStakeTokenCmd.Flags().StringVarP(&weiToMint, "wei-to-mint", "", "100000000000000000000", "eth to mint into erc20 WETH tokens, in wei, default (50 WETH)")
	mintStakeTokenCmd.Flags().Int64VarP(&bumpPricePercent, "bump-price-percent", "", 100, "percent to increase the suggested gas price by")

	// Bind flags for bridge eth
	bridgeEthCmd.Flags().StringVarP(&valPrivKeys, "validator-priv-keys", "", "", "comma-separated validator private keys")
	bridgeEthCmd.Flags().StringVarP(&l1ChainIdStr, "l1-chain-id", "", "11155111", "l1 chain id (sepolia default)")
	bridgeEthCmd.Flags().StringVarP(&l1EndpointUrl, "l1-endpoint", "", "", "l1 endpoint")
	bridgeEthCmd.Flags().StringVarP(&inboxAddrStr, "inbox-address", "", "", "inbox address")
	bridgeEthCmd.Flags().StringVarP(&weiToDeposit, "wei-to-deposit", "", "2000000000000000", "eth to bridge over, in wei (0.002 ETH)")
	bridgeEthCmd.Flags().Int64VarP(&bumpPricePercent, "bump-price-percent", "", 100, "percent to increase the suggested gas price by")
}

func main() {
	rootCmd.AddCommand(mintStakeTokenCmd)
	rootCmd.AddCommand(bridgeEthCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func bumpGasPrice(suggested *big.Int) *big.Int {
	bumpMultiplier := new(big.Int).SetInt64(bumpPricePercent)
	increase := new(big.Int).Mul(suggested, bumpMultiplier)
	increasedGasPrice := new(big.Int).Div(increase, big.NewInt(100))
	return new(big.Int).Add(suggested, increasedGasPrice)
}

func weiToGwei(wei *big.Int) *big.Int {
	// 1 Gwei = 10^9 Wei
	gwei := new(big.Int).Div(wei, big.NewInt(1000000000))
	return gwei
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
	fmt.Println("Now minting the stake token required for BOLD assertion posting and challenge participation")
	fmt.Println("This command will convert the specified amount of testnet ETH into a WETH ERC-20 stake token")
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
		suggested, err := client.SuggestGasPrice(ctx)
		if err != nil {
			panic(err)
		}
		nonce, err := client.PendingNonceAt(ctx, txOpts.From)
		if err != nil {
			panic(err)
		}
		txOpts.Nonce = new(big.Int).SetUint64(nonce)
		fmt.Printf("Suggested gas price: %s gwei, bumping by %d percent\n", weiToGwei(suggested).String(), bumpPricePercent)
		txOpts.GasPrice = bumpGasPrice(suggested)
		fmt.Printf("Bumped to price: %s gwei\n", weiToGwei(txOpts.GasPrice).String())

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

		depositAmount, ok := new(big.Int).SetString(weiToMint, 10)
		if !ok {
			panic("Not ok deposit amount")
		}
		txOpts.Value = depositAmount
		tx, err := tokenBindings.Deposit(txOpts)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Sent token minting tx with hash %#x\n", tx.Hash())
		txOpts.Value = big.NewInt(0)
		maxUint256 := new(big.Int)
		maxUint256.Exp(big.NewInt(2), big.NewInt(256), nil).Sub(maxUint256, big.NewInt(1))

		suggested, err = client.SuggestGasPrice(ctx)
		if err != nil {
			panic(err)
		}
		txOpts.Nonce = new(big.Int).Add(txOpts.Nonce, big.NewInt(1))
		fmt.Printf("Suggested gas price: %s gwei, bumping by %d percent\n", weiToGwei(suggested).String(), bumpPricePercent)
		txOpts.GasPrice = bumpGasPrice(suggested)
		fmt.Printf("Bumped to price: %s gwei\n", weiToGwei(txOpts.GasPrice).String())
		fmt.Printf("Your %#x address is giving the BOLD rollup contract %#x a full allowance\n", txOpts.From, rollupAddr)
		tx, err = tokenBindings.Approve(txOpts, rollupAddr, maxUint256)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Sent tx that approves the rollup contract's spending of your WETH ERC-20 with hash %#x\n", tx.Hash())

		fmt.Printf("Now your %#x address is giving the BOLD challenge manager contract %#x a full allowance\n", txOpts.From, chalManagerAddr)
		suggested, err = client.SuggestGasPrice(ctx)
		if err != nil {
			panic(err)
		}
		txOpts.Nonce = new(big.Int).Add(txOpts.Nonce, big.NewInt(1))
		fmt.Printf("Suggested gas price: %s gwei, bumping by %d percent\n", weiToGwei(suggested).String(), bumpPricePercent)
		txOpts.GasPrice = bumpGasPrice(suggested)
		fmt.Printf("Bumped to price: %s gwei\n", weiToGwei(txOpts.GasPrice).String())
		tx, err = tokenBindings.Approve(txOpts, chalManagerAddr, maxUint256)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Sent tx that approves challenge manager's spending of your WETH ERC-20 with hash %#x\n", tx.Hash())
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
	depositAmount, ok := new(big.Int).SetString(weiToDeposit, 10)
	if !ok {
		panic("not ok deposit amount")
	}
	gasLimit := uint64(150000)
	fmt.Println("Now bridging ETH from Sepolia to the Arbitrum BOLD L2 rollup")
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
		fmt.Printf("Suggested gas price: %s gwei, bumping by %d percent\n", weiToGwei(gasPrice).String(), bumpPricePercent)
		suggested := bumpGasPrice(gasPrice)
		fmt.Printf("Bumped to price: %s gwei\n", weiToGwei(suggested).String())
		tx := types.NewTransaction(nonce, toAddress, depositAmount, gasLimit, suggested, data)
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
