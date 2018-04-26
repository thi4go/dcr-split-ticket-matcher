package validations

import (
	"bytes"
	"encoding/hex"

	"github.com/decred/dcrd/blockchain"
	"github.com/decred/dcrd/chaincfg"
	"github.com/decred/dcrd/chaincfg/chainhash"
	"github.com/decred/dcrd/dcrutil"
	"github.com/decred/dcrd/txscript"
	"github.com/decred/dcrd/wire"
	"github.com/matheusd/dcr-split-ticket-matcher/pkg/matcher"
	"github.com/pkg/errors"
)

// VoterLotteryPkScriptSize is the number of bytes in the output script of the
// output that stores the voter lottery commitment. The entry is:
// OP_RETURN
// OP_DATA32
// [32-byte data]
const VoterLotteryPkScriptSize = 1 + 1 + 32

// CheckSplit validates that the given split transaction respects the rules for
// the split ticket matching service
func CheckSplit(split *wire.MsgTx, utxos UtxoMap,
	secretHashes []matcher.SecretNumberHash, mainchainHash *chainhash.Hash,
	params *chaincfg.Params) error {

	var err error

	err = blockchain.CheckTransactionSanity(split, params)
	if err != nil {
		return errors.Wrap(err, "split tx failed sanity check")
	}

	if len(split.TxOut[0].PkScript) != VoterLotteryPkScriptSize {
		return errors.Errorf("size of pkscript of output 0 (%d) of split tx "+
			"doesn't have expected length (%d)", len(split.TxOut[0].PkScript),
			VoterLotteryPkScriptSize)
	}

	if split.TxOut[0].PkScript[0] != txscript.OP_RETURN {
		return errors.Errorf("output 0 of split tx is not an OP_RETURN")
	}

	targetVoterHash := matcher.SecretNumberHashesHash(secretHashes, mainchainHash)

	// pick the range [2:] because the first byte is the OP_RETURN, the second
	// is the push data op
	splitVoterCommitment := split.TxOut[0].PkScript[2:]
	if !bytes.Equal(targetVoterHash, splitVoterCommitment) {
		return errors.Errorf("voter lottery commitment (%s) does not equal "+
			"the expected value (%s)", hex.EncodeToString(splitVoterCommitment),
			hex.EncodeToString(targetVoterHash))
	}

	return nil
}

// CheckSignedSplit validates that the given signed split transaction is
// valid according to split ticket matcher rules. Only safe to be called on
// split transactions that passed CheckSplit
func CheckSignedSplit(split *wire.MsgTx, utxos UtxoMap, params *chaincfg.Params) error {
	var totalAmountIn int64
	for i, in := range split.TxIn {
		utxo, hasUtxo := utxos[in.PreviousOutPoint]
		if !hasUtxo {
			return errors.Errorf("utxo for input %d of split tx not provided", i)
		}

		// TODO: check if utxo is spent

		// ensure the input actually signs the split transaction
		engine, err := txscript.NewEngine(utxo.PkScript, split, i,
			currentScriptFlags, utxo.Version, nil)
		if err != nil {
			return errors.Wrapf(err, "error creating engine to process input "+
				"%d of split tx", i)
		}

		err = engine.Execute()
		if err != nil {
			return errors.Wrapf(err, "error executing script of input %d of "+
				"split tx", i)
		}

		newAmountIn := totalAmountIn + int64(utxo.Value)
		if (newAmountIn < 0) || (newAmountIn > dcrutil.MaxAmount) {
			return errors.Errorf("overflow of total input amount of split tx "+
				"at index %d", i)
		}
		totalAmountIn = newAmountIn
	}

	totalAmountOut := totalOutputAmount(split)
	txFee := totalAmountIn - int64(totalAmountOut)

	serializedSize := int64(split.SerializeSize())
	minFee := (serializedSize * int64(minRelayFeeRate)) / 1000
	if txFee < minFee {
		return errors.Errorf("split tx fee (%s) less than minimum required "+
			"amount (%s)", dcrutil.Amount(txFee), dcrutil.Amount(minFee))
	}

	return nil
}
