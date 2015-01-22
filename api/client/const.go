package client

import (
	common "github.com/hoffie/larasync/api/common"
)

const (
	// PrivateKeySize denotes how many bytes a private key needs (binary encoded)
	PrivateKeySize = common.PrivateKeySize
	// PublicKeySize denotes how many bytes a pubkey needs (binary encoded)
	PublicKeySize = common.PublicKeySize
	// SignatureSize denotes how many bytes a sig needs (binary encoded)
	SignatureSize = common.SignatureSize
)
