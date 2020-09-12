package web

import "testing"

func TestDigestsMatch(t *testing.T) {
	testBody := "foobar"

	testKey := "79a0f8f74072fa834a7251ffa1315536dd4c1e0d"

	// computed from: `echo -n "foobar" | openssl dgst -sha1 -hmac "79a0f8f74072fa834a7251ffa1315536dd4c1e0d"`
	knownHash := "sha1=ac1d15d169dd3ef41142da74ebc71bdb4c6d2aff"

	err := DigestsMatch([]byte(testBody), testKey, knownHash)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
}
