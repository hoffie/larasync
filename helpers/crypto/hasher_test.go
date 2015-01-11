package crypto

import (
	. "gopkg.in/check.v1"
)

type HasherTests struct {
	hashKey [HashingKeySize]byte
}

var _ = Suite(&HasherTests{})

func (t *HasherTests) SetUpTest(c *C) {
	t.hashKey = [HashingKeySize]byte{}
}

func (t *HasherTests) getHasher() *Hasher {
	return NewHasher(t.hashKey)
}

func (t *HasherTests) TestHashing(c *C) {
	t.getHasher().Hash([]byte("test"))
}

func (t *HasherTests) TestDifferntKeys(c *C) {
	testBytes := []byte("test")
	firstCheck := t.getHasher().StringHash(testBytes)
	t.hashKey[0] = 200
	secondCheck := t.getHasher().StringHash(testBytes)
	c.Assert(firstCheck, Not(Equals), secondCheck)
}
