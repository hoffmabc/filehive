package fil

import (
	"context"
	"crypto/rand"
	addr "github.com/filecoin-project/go-address"
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
	"github.com/prometheus/common/log"
	pow "github.com/textileio/powergate/api/client"
	"io"
	"math/big"
	"os"
	"sync"
	"time"
)

// PowergateBackend is a mock backend for a Filecoin service using Powergate
type PowergateBackend struct {
	dataDir    string
	jobs       map[cid.Cid]string
	adminToken string
	powClient  *pow.Client

	mtx sync.RWMutex
}

type PowergateUser struct {
	id    string
	token string
}

// NewPowergateBackend instantiates a new FilecoinBackend
func NewPowergateBackend(dataDir string, adminToken string) (*PowergateBackend, error) {
	client, err := pow.NewClient("127.0.0.1:5002")
	if err != nil {
		return nil, nil
	}

	if err := os.MkdirAll(dataDir, os.ModePerm); err != nil {
		return nil, err
	}
	return &PowergateBackend{dataDir: dataDir, jobs: make(map[cid.Cid]string), mtx: sync.RWMutex{}, adminToken: adminToken, powClient: client}, nil
}

// Store will put a file to Filecoin and pay for it out of the provided
// address. A jobID is return or an error.
func (f *PowergateBackend) Store(data io.Reader, addr addr.Address) (jobID, contentID cid.Cid, size int64, err error) {
	return contentID, jobID, size, nil
}

// TODO
func (f *PowergateBackend) JobStatus(jobID cid.Cid) (string, error) {
	return "", nil
}

// TODO
func (f *PowergateBackend) Get(id cid.Cid) (io.Reader, error) {
	return nil, nil
}

func (f *PowergateBackend) CreateUser() (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(60)*time.Second)
	defer cancel()

	response, err := f.powClient.Admin.Users.Create(ctx)
	if err != nil {
		return "", "", err
	}

	return response.User.Id, response.User.Token, nil
}

// MockWalletBackend is a mock backend for the wallet that allows
// for making mock transactions and generating mock blocks.
type PowergateWalletBackend struct {
	transactions map[string][]Transaction
	nextAddr     *addr.Address
	nextTxid     *cid.Cid
	nextTime     *time.Time
	powClient    *pow.Client
	mtx          sync.RWMutex
}

// NewMockWalletBackend instantiates a new WalletBackend.
func NewPowergateWalletBackend() *PowergateWalletBackend {
	client, err := pow.NewClient("127.0.0.1:5002")
	if err != nil {
		return nil
	}

	return &PowergateWalletBackend{
		transactions: make(map[string][]Transaction),
		mtx:          sync.RWMutex{},
		powClient:    client,
	}
}

// GenerateToAddress creates mock coins and sends them to the address.
func (w *PowergateWalletBackend) GenerateToAddress(addr string, amount *big.Int) {
	w.mtx.Lock()
	defer w.mtx.Unlock()

	// Get address for the user from Powergate

	var txid cid.Cid
	if w.nextTxid != nil {
		txid = *w.nextTxid
		w.nextTxid = nil
	} else {
		txid, _ = w.randCid()
	}

	ts := time.Now()
	if w.nextTime != nil {
		ts = *w.nextTime
	}

	tx := Transaction{
		ID:        txid,
		To:        addr,
		Timestamp: ts,
		Amount:    amount,
	}

	w.transactions[addr] = append(w.transactions[addr], tx)
}

// NewAddress generates a new address and store the key in the backend.
func (w *PowergateWalletBackend) NewAddress(userToken string) (string, error) {
	w.mtx.Lock()
	defer w.mtx.Unlock()

	ctx := context.WithValue(context.Background(), pow.AuthKey, userToken)
	log.Debug(ctx.Value(pow.AuthKey))

	address, err := w.powClient.Wallet.NewAddress(ctx, "")
	if err != nil {
		return "", err
	}

	return address.Address, nil
}

func (w *PowergateWalletBackend) SetNextAddress(addr addr.Address) {
	w.mtx.Lock()
	defer w.mtx.Unlock()

	w.nextAddr = &addr
}

func (w *PowergateWalletBackend) SetNextTxid(id cid.Cid) {
	w.mtx.Lock()
	defer w.mtx.Unlock()

	w.nextTxid = &id
}

func (w *PowergateWalletBackend) SetNextTime(timestamp time.Time) {
	w.mtx.Lock()
	defer w.mtx.Unlock()

	w.nextTime = &timestamp
}

// Send filecoin from one address to another. Returns the cid of the
// transaction.
func (w *PowergateWalletBackend) Send(from, to string, amount *big.Int, userToken string) (cid.Cid, error) {
	w.mtx.Lock()
	defer w.mtx.Unlock()

	balance, err := w.balance(from, userToken)
	if err != nil {
		return cid.Cid{}, err
	}

	if amount.Cmp(balance) > 0 {
		return cid.Cid{}, ErrInsuffientFunds
	}

	var txid cid.Cid
	if w.nextTxid != nil {
		txid = *w.nextTxid
		w.nextTxid = nil
	} else {
		txid, err = randCid()
		if err != nil {
			return cid.Cid{}, err
		}
	}

	ts := time.Now()
	if w.nextTime != nil {
		ts = *w.nextTime
	}

	tx := Transaction{
		ID:        txid,
		To:        to,
		From:      from,
		Timestamp: ts,
		Amount:    amount,
	}

	w.transactions[to] = append(w.transactions[to], tx)
	if to != from {
		w.transactions[from] = append(w.transactions[from], tx)
	}

	return txid, nil
}

// Balance returns the balance for an address.
func (w *PowergateWalletBackend) Balance(address string, userToken string) (*big.Int, error) {
	w.mtx.RLock()
	defer w.mtx.RUnlock()

	return w.balance(address, userToken)
}

func (w *PowergateWalletBackend) balance(address string, userToken string) (*big.Int, error) {
	w.mtx.Lock()
	defer w.mtx.Unlock()

	ctx := context.WithValue(context.Background(), pow.AuthKey, userToken)
	log.Debug(ctx.Value(pow.AuthKey))

	balance, err := w.powClient.Wallet.Balance(ctx, address)
	if err != nil {
		return nil, err
	}

	b, ok := new(big.Int).SetString(balance.Balance, 10)
	if !ok {
		return nil, err
	}

	return b, nil
}

// Transactions returns the list of transactions for an address.
func (w *PowergateWalletBackend) Transactions(addr string, limit, offset int) ([]Transaction, error) {
	w.mtx.RLock()
	defer w.mtx.RUnlock()

	txs := w.transactions[addr]

	if limit < 0 {
		limit = len(txs)
	}
	if offset < 0 {
		offset = 0
	}
	if offset+limit > len(txs) {
		limit = len(txs) - offset
	}

	return txs[offset : offset+limit], nil
}

func (w *PowergateWalletBackend) randCid() (cid.Cid, error) {
	r := make([]byte, 32)
	rand.Read(r)

	mh, err := multihash.Encode(r, multihash.SHA2_256)
	if err != nil {
		return cid.Cid{}, err
	}

	return cid.NewCidV1(cid.Raw, mh), nil
}