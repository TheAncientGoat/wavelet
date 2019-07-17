package server

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"io/ioutil"
	"strconv"

	"github.com/perlin-network/wavelet"
	"github.com/perlin-network/wavelet/cmd/cli/tui/logger"
	"github.com/perlin-network/wavelet/sys"
	"github.com/pkg/errors"
)

// TODO(diamond): Sometime in the future, maybe when I'm 70, I would like all
// this to return structs instead of calling logger directly.

func (s *Server) Status() {
	preferredID := "N/A"

	if preferred := s.Ledger.Finalizer().Preferred(); preferred != nil {
		preferredID = hex.EncodeToString(preferred.ID[:])
	}

	count := s.Ledger.Finalizer().Progress()

	snapshot := s.Ledger.Snapshot()
	publicKey := s.Keys.PublicKey()

	accountsLen := wavelet.ReadAccountsLen(snapshot)

	balance, _ := wavelet.ReadAccountBalance(snapshot, publicKey)
	stake, _ := wavelet.ReadAccountStake(snapshot, publicKey)
	reward, _ := wavelet.ReadAccountReward(snapshot, publicKey)
	nonce, _ := wavelet.ReadAccountNonce(snapshot, publicKey)

	round := s.Ledger.Rounds().Latest()
	rootDepth := s.Ledger.Graph().RootDepth()

	peers := s.Client.ClosestPeerIDs()
	peerIDs := make([]string, 0, len(peers))

	for _, id := range peers {
		peerIDs = append(peerIDs, id.String())
	}

	// Add the root ID and self ID to the history
	s.History.add(hex.EncodeToString(round.End.ID[:]), HistoryEntryRoot)
	s.History.add(hex.EncodeToString(publicKey[:]), HistoryEntrySelf)

	s.logger.Level(logger.WithInfo("Node status:").
		F("difficulty", "%v", round.ExpectedDifficulty(
			sys.MinDifficulty, sys.DifficultyScaleFactor)).
		F("round", "%d", round.Index).
		F("root_id", "%x", round.End.ID).
		F("height", "%d", s.Ledger.Graph().Height()).
		F("id", "%x", publicKey).
		F("balance", "%d", balance).
		F("stake", "%d", stake).
		F("reward", "%d", reward).
		F("nonce", "%d", nonce).
		F("peers", "%v", peerIDs).
		F("num_tx", "%d", s.Ledger.Graph().DepthLen(&rootDepth, nil)).
		F("num_missing_tx", "%d", s.Ledger.Graph().MissingLen()).
		F("num_tx_in_store", "%d", s.Ledger.Graph().Len()).
		F("num_accounts_in_store", "%d", accountsLen).
		F("preferred_id", preferredID).
		F("preferred_votes", "%d", count))
}

// gasLimit is optional
func (s *Server) Pay(recipient [wavelet.SizeAccountID]byte,
	amount, gasLimit int, additional []byte) (*wavelet.Transaction, error) {

	// Create a new payload and write the recipient
	payload := bytes.NewBuffer(nil)
	payload.Write(recipient[:])

	// Make an int64 bytes buffer
	var intBuf = make([]byte, 8)

	// Write the amount
	binary.LittleEndian.PutUint64(intBuf, uint64(amount))
	payload.Write(intBuf)

	// Write the gas limit
	binary.LittleEndian.PutUint64(intBuf, uint64(gasLimit))
	payload.Write(intBuf)

	if additional != nil {
		payload.Write(additional)
	}

	tx, err := s.sendTx(wavelet.NewTransaction(
		s.Keys, sys.TagTransfer, payload.Bytes(),
	))

	if err != nil {
		return nil, err
	}

	s.logger.Level(logger.WithSuccess("Paid").
		F("tx_id", "%x", tx.ID))

	return tx, nil
}

// Call calls a smart contract function
func (s *Server) Call(recipient [wavelet.SizeAccountID]byte,
	amount, gasLimit int, fn FunctionCall) (*wavelet.Transaction, error) {

	// Make an int buffer
	var intBuf = make([]byte, 8)

	// Make a payload buffer
	var payload bytes.Buffer

	// Write the function name length and name
	binary.LittleEndian.PutUint32(intBuf[:4], uint32(len(fn.Name)))
	payload.Write(intBuf[:4])
	payload.WriteString(fn.Name)

	// Make a function parameters buffer
	var params bytes.Buffer

	for i, arg := range fn.Params {
		b, err := arg.Encode()
		if err != nil {
			e := logger.WithError(err).
				Wrap("Can't decode function " + strconv.Itoa(i))

			s.logger.Level(e)

			return nil, e
		}

		params.Write(b)
	}

	// Write the parameters buffer into the payload
	buf := params.Bytes()
	binary.LittleEndian.PutUint32(intBuf[:4], uint32(len(buf)))
	payload.Write(intBuf[:4]) // write len
	payload.Write(buf)        // write body

	tx, err := s.Pay(recipient, amount, gasLimit, payload.Bytes())
	if err != nil {
		return nil, err
	}

	s.logger.Level(logger.WithSuccess("Called function "+fn.Name).
		F("tx_id", "%x", tx.ID))

	return tx, nil
}

var (
	ErrInvalidID       = errors.New("Invalid transaction/account ID")
	ErrInvalidIDLength = errors.New("Invalid transaction/account ID length")
)

func (s *Server) Find(addr string) error {
	addrHex, err := hex.DecodeString(addr)
	if err != nil {
		return err
	}

	if len(addrHex) != wavelet.SizeTransactionID &&
		len(addrHex) != wavelet.SizeAccountID {

		err := logger.WithError(ErrInvalidIDLength)
		err.F("length", "%d", len(addrHex))

		s.logger.Level(err)

		return err
	}

	snapshot := s.Ledger.Snapshot()

	// See if the ID given is an account ID
	var accountID wavelet.AccountID
	copy(accountID[:], addrHex)

	// Read account info
	balance, _ := wavelet.ReadAccountBalance(snapshot, accountID)
	stake, _ := wavelet.ReadAccountStake(snapshot, accountID)
	nonce, _ := wavelet.ReadAccountNonce(snapshot, accountID)
	reward, _ := wavelet.ReadAccountReward(snapshot, accountID)

	_, isContract := wavelet.ReadAccountContractCode(snapshot, accountID)
	numPages, _ := wavelet.ReadAccountContractNumPages(snapshot, accountID)

	// The ID is an account
	if balance > 0 || stake > 0 || nonce > 0 || isContract || numPages > 0 {
		s.logger.Level(logger.WithSuccess("Found account: "+addr).
			F("balance", "%d", balance).
			F("stake", "%d", stake).
			F("nonce", "%d", nonce).
			F("reward", "%d", reward).
			F("is_contract", "%v", isContract).
			F("num_pages", "%d", numPages))

		s.History.add(addr, HistoryEntryAccount)

		// Exit, not a transaction
		return nil
	}

	// ID was not an account, probably a transaction then
	var txID wavelet.TransactionID
	copy(txID[:], addrHex)

	tx := s.Ledger.Graph().FindTransaction(txID)

	if tx != nil {
		var parents = make([]string, 0, len(tx.ParentIDs))
		for _, parentID := range tx.ParentIDs {
			parents = append(parents, hex.EncodeToString(parentID[:]))
		}

		s.logger.Level(logger.WithSuccess("Found transaction "+addr).
			F("parents", "%v", parents).
			F("sender", "%x", tx.Sender).
			F("creator", "%x", tx.Creator).
			F("nonce", "%d", tx.Nonce).
			F("tag", "%d", tx.Tag).
			F("depth", "%d", tx.Depth).
			F("seed", "%x", tx.Seed[:]).
			F("seed_zero_prefix_len", "%d", tx.SeedLen))

		s.History.add(addr, HistoryEntryTransaction)

		return nil
	}

	s.logger.Level(logger.WithError(ErrInvalidID))
	return ErrInvalidID
}

// Spawn spawns a smart contract
func (s *Server) Spawn(pathToContract string) error {
	code, err := ioutil.ReadFile(pathToContract)
	if err != nil {
		e := logger.WithError(err)
		e.Wrap("Failed to find the smart contract")
		e.F("path", pathToContract)

		s.logger.Level(e)
		return e
	}

	var buf [8]byte

	w := bytes.NewBuffer(nil)

	// Write a fake gas fee
	binary.LittleEndian.PutUint64(buf[:], 100000000)
	w.Write(buf[:])

	// Write the payload size
	binary.LittleEndian.PutUint64(buf[:], uint64(len(code)))
	w.Write(buf[:])

	w.Write(code) // Smart contract code.

	newTx := wavelet.NewTransaction(s.Keys, sys.TagContract, w.Bytes())

	tx, err := s.sendTx(newTx)
	if err != nil {
		return err
	}

	s.logger.Level(logger.WithSuccess("Smart contract deployed").
		F("id", "%x", tx.ID))

	return nil
}

// PlaceStake places a stake
func (s *Server) PlaceStake(amount int) {
	var intBuf [8]byte
	var payload bytes.Buffer

	payload.WriteByte(sys.PlaceStake)

	// Write the amount
	binary.LittleEndian.PutUint64(intBuf[:8], uint64(amount))
	payload.Write(intBuf[:8])

	tx, err := s.sendTx(wavelet.NewTransaction(
		s.Keys, sys.TagStake, payload.Bytes(),
	))

	if err != nil {
		return
	}

	s.logger.Level(logger.WithSuccess("Stake placed").
		F("id", "%x", tx.ID))
}

// WithdrawStake withdraws the stake
func (s *Server) WithdrawStake(amount int) {
	var intBuf [8]byte
	var payload bytes.Buffer

	payload.WriteByte(sys.WithdrawStake)

	// Write the amount
	binary.LittleEndian.PutUint64(intBuf[:8], uint64(amount))
	payload.Write(intBuf[:8])

	tx, err := s.sendTx(wavelet.NewTransaction(
		s.Keys, sys.TagStake, payload.Bytes(),
	))

	if err != nil {
		return
	}

	s.logger.Level(logger.WithSuccess("Stake withdrew").
		F("id", "%x", tx.ID))
}

// WithdrawReward withdraws the reward
func (s *Server) WithdrawReward(amount int) {
	var intBuf [8]byte
	var payload bytes.Buffer

	payload.WriteByte(sys.WithdrawReward)

	// Write the amount
	binary.LittleEndian.PutUint64(intBuf[:8], uint64(amount))
	payload.Write(intBuf[:8])

	tx, err := s.sendTx(wavelet.NewTransaction(
		s.Keys, sys.TagStake, payload.Bytes(),
	))

	if err != nil {
		return
	}

	s.logger.Level(logger.WithSuccess("Reward withdrew").
		F("id", "%x", tx.ID))
}

func (s *Server) sendTx(tx wavelet.Transaction) (*wavelet.Transaction, error) {
	tx = wavelet.AttachSenderToTransaction(
		s.Keys, tx,
		s.Ledger.Graph().FindEligibleParents()...,
	)

	if err := s.Ledger.AddTransaction(tx); err != nil {
		if errors.Cause(err) != wavelet.ErrMissingParents {
			e := logger.WithError(err)
			e.Wrap("Failed to create transaction")
			e.F("tx_id", "%x", tx.ID)

			s.logger.Level(e)

			return nil, e
		}
	}

	// Add the ID into history
	s.History.add(hex.EncodeToString(tx.ID[:]), HistoryEntryTransaction)

	return &tx, nil
}
