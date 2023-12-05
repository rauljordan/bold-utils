# Mock Stake Token Approver

Mints and approves spending of a mock stake token implementation on Ethereum testnets for use in the Arbitrum BOLD challenge protocol.

Usage:
```
-gwei-to-deposit uint
    tokens to deposit (default 10000)
-l1-chain-id string
    l1 chain id (sepolia default) (default "11155111")
-l1-endpoint string
    l1 endpoint
-rollup-address string
    rollup address
-stake-token-address string
    rollup address
-validator-priv-keys string
    comma-separated, validator private keys to fund and approve mock ERC20 stake token
```