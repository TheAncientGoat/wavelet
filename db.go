package wavelet

var (
	keyAccounts       = [...]byte{0x1}
	keyAccountBalance = [...]byte{0x2}
	keyAccountStake   = [...]byte{0x3}

	keyAccountContractCode     = [...]byte{0x4}
	keyAccountContractNumPages = [...]byte{0x5}
	keyAccountContractPages    = [...]byte{0x6}

	keyCriticalTimestampHistory = [...]byte{0x7}

	keyLedgerDifficulty = [...]byte{0x8}
	keyLedgerViewID     = [...]byte{0x9}
	keyLedgerGenesis    = [...]byte{0x10}

	keyGraphRoot = [...]byte{0x11}
)
