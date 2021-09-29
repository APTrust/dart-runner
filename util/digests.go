package util

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"hash"

	"github.com/APTrust/dart-runner/constants"
)

func GetHashes(algs []string) map[string]hash.Hash {
	hashes := make(map[string]hash.Hash)
	if StringListContains(algs, constants.AlgMd5) {
		hashes[constants.AlgMd5] = md5.New()
	}
	if StringListContains(algs, constants.AlgSha1) {
		hashes[constants.AlgSha1] = sha1.New()
	}
	if StringListContains(algs, constants.AlgSha256) {
		hashes[constants.AlgSha256] = sha256.New()
	}
	if StringListContains(algs, constants.AlgSha512) {
		hashes[constants.AlgSha512] = sha512.New()
	}
	return hashes
}
