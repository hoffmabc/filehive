package fil

import (
	"encoding/json"
	"errors"
	addr "github.com/filecoin-project/go-address"
	"github.com/ipfs/go-cid"
	"io"
	"math/big"
	"time"
)

func init() {
	addr.CurrentNetwork = addr.Mainnet
}

// ErrInsuffientFunds is an error that should be returned by WalletBackend.Send method
// if the address does not have enough funds.
var ErrInsuffientFunds = errors.New("insufficient funds")

const attoFilPerFilecoin = 1000000000000000000

// FilecoinBackend is an interface to a Filecoin backend that interacts with the
// Filecoin network and handles storage deals and retrieval.
type FilecoinBackend interface {
	// Store will put a file to Filecoin and pay for it out of the provided
	// address. A jobID is return or an error.
	Store(data io.Reader, addr addr.Address, userToken string) (jobID, contentID string, size int64, err error)

	// TODO
	JobStatus(jobID cid.Cid) (string, error)

	Get(cid string, userToken string) (io.Reader, error)

	CreateUser() (id string, token string, error error)
}

// WalletBackend is an interface for a Filecoin wallet that can hold the keys
// for multiple addresses and can make transactions.
type WalletBackend interface {
	// NewAddress generates a new address and store the key in the backend.
	NewAddress(userToken string) (string, error)

	// Send filecoin from one address to another. Returns the cid of the
	// transaction.
	Send(from, to string, amount *big.Int, userToken string) (string, error)

	// Balance returns the balance for an address.
	Balance(address string, userToken string) (*big.Int, error)

	// Transactions returns the list of transactions for an address.
	Transactions(addr string, limit, offset int) ([]Transaction, error)
}

// Transaction represents a Filecoin transaction.
type Transaction struct {
	ID        string
	From      string
	To        string
	Amount    *big.Int
	Timestamp time.Time
}

// MarshalJSON is a custom JSON marshaller for Transaction.
func (t *Transaction) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID        string    `json:"transactionID"`
		From      string    `json:"from"`
		To        string    `json:"to"`
		Amount    float64   `json:"amount"`
		Timestamp time.Time `json:"timestamp"`
	}{
		ID:        t.ID,
		From:      t.From,
		To:        t.To,
		Amount:    AttoFILToFIL(t.Amount),
		Timestamp: t.Timestamp,
	})
}

// FILtoAttoFIL converts a float containing an amount of Filecoin to
// a big.Int representation in the attoFil base unit.
func FILtoAttoFIL(fil float64) *big.Int {
	bigAtto := big.NewFloat(fil)
	bigAtto = bigAtto.Mul(bigAtto, big.NewFloat(attoFilPerFilecoin))
	ret, acc := bigAtto.Int(nil)

	if acc == big.Above {
		ret.Add(ret, big.NewInt(-1))
	}
	if acc == big.Below {
		ret.Add(ret, big.NewInt(1))
	}

	return ret
}

// AttoFILtoFIL converts a big.Int containing the attoFIL base unit to
// a float of the amount of Filecoin.
func AttoFILToFIL(attoFIL *big.Int) float64 {
	bigAtto := new(big.Float).SetInt(attoFIL)
	bigAtto = bigAtto.Quo(bigAtto, new(big.Float).SetInt(big.NewInt(attoFilPerFilecoin)))
	ret, _ := bigAtto.Float64()
	return ret
}
