package client

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
	confirm "github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/gagliardetto/solana-go/text"
)

const (
	privateKey = "45TzHcRSR3ps9eWgLDSCDpB8kVh1F2JXMLmE9T5qDi1Wb2bVMo4LuXUWeAo2Uv8HwVmrEjGsCUkwE62fz2wz6czy"
	publicKey  = "8fGaCBDAC9nW8x2gpsBRVfoe97t1L8Jptvp7w7Difu1P"

	privateKey1 = "5VTwsprzQ55JLPRwPnq86eLBbn5ZcCCcZPzR9kAVvzKNXqHzGCQHs1GowUdvSZLuZzLCdoojCHAsY14ZCaTBoHXf"
	publicKey1  = "Dtr8PEdWyAAmwWkppiE76UBY41y9qxEJciaogDeVadny"
)

func TestRequestAirdrop(t *testing.T) {
	// Create a new account:
	//account := solana.NewWallet()
	account, err := solana.PrivateKeyFromBase58(privateKey)
	if err != nil {
		panic(err)
	}
	fmt.Println("account private key:", account)
	fmt.Println("account public key:", account.PublicKey())
	fmt.Println(rpc.DevNet_RPC)
	// Create a new RPC client:
	client := rpc.New(rpc.DevNet_RPC)

	// Airdrop 100 SOL to the new account:
	out, err := client.RequestAirdrop(
		context.Background(),
		account.PublicKey(),
		solana.LAMPORTS_PER_SOL*5, //max is 5
		rpc.CommitmentFinalized,
	)
	if err != nil {
		panic(err)
	}
	fmt.Println("airdrop transaction signature:", out)
}

func TestTransfer(t *testing.T) {
	// Create a new RPC client:
	rpcClient := rpc.New(rpc.DevNet_RPC)

	// Create a new WS client (used for confirming transactions)
	wsClient, err := ws.Connect(context.Background(), rpc.DevNet_WS)
	if err != nil {
		panic(err)
	}

	// Load the account that you will send funds FROM:
	accountFrom, err := solana.PrivateKeyFromBase58(privateKey)
	if err != nil {
		panic(err)
	}
	fmt.Println("accountFrom private key:", accountFrom)
	fmt.Println("accountFrom public key:", accountFrom.PublicKey())

	// The public key of the account that you will send sol TO:
	accountTo := solana.MustPublicKeyFromBase58(publicKey1)
	// The amount to send (in lamports);
	// 1 sol = 1000000000 lamports
	amount := uint64(1e9)

	recent, err := rpcClient.GetRecentBlockhash(context.TODO(), rpc.CommitmentFinalized)
	if err != nil {
		panic(err)
	}

	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			system.NewTransferInstruction(
				amount,
				accountFrom.PublicKey(),
				accountTo,
			).Build(),
		},
		recent.Value.Blockhash,
		solana.TransactionPayer(accountFrom.PublicKey()),
	)
	if err != nil {
		panic(err)
	}

	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if accountFrom.PublicKey().Equals(key) {
				return &accountFrom
			}
			return nil
		},
	)
	if err != nil {
		panic(fmt.Errorf("unable to sign transaction: %w", err))
	}
	spew.Dump(tx)
	// Pretty print the transaction:
	tx.EncodeTree(text.NewTreeEncoder(os.Stdout, "Transfer SOL"))

	// Send transaction, and wait for confirmation:
	sig, err := confirm.SendAndConfirmTransaction(
		context.TODO(),
		rpcClient,
		wsClient,
		tx,
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(sig)

	// Or just send the transaction WITHOUT waiting for confirmation:
	// sig, err := rpcClient.SendTransactionWithOpts(
	//   context.TODO(),
	//   tx,
	//   false,
	//   rpc.CommitmentFinalized,
	// )
	// if err != nil {
	//   panic(err)
	// }
	// spew.Dump(sig)
}

func TestGetBalance(t *testing.T) {
	//endpoint := rpc.MainNetBeta_RPC
	endpoint := rpc.DevNet_RPC
	client := rpc.New(endpoint)

	pubKey := solana.MustPublicKeyFromBase58(publicKey)
	out, err := client.GetBalance(
		context.TODO(),
		pubKey,
		rpc.CommitmentFinalized,
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(out)
	spew.Dump(out.Value) // total lamports on the account; 1 sol = 1000000000 lamports

	var lamportsOnAccount = new(big.Float).SetUint64(uint64(out.Value))
	// Convert lamports to sol:
	var solBalance = new(big.Float).Quo(lamportsOnAccount, new(big.Float).SetUint64(solana.LAMPORTS_PER_SOL))

	// WARNING: this is not a precise conversion.
	fmt.Println("◎", solBalance.Text('f', 10))
}
