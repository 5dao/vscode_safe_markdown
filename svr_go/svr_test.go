package main

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJson(t *testing.T) {
	{
		s1 := "we are floating in space"
		s1b64 := base64.StdEncoding.EncodeToString([]byte(s1))
		assert.Equal(t, "d2UgYXJlIGZsb2F0aW5nIGluIHNwYWNl", s1b64)

		s1Data, err := base64.StdEncoding.DecodeString(s1b64)
		assert.NoError(t, err)

		assert.Equal(t, s1, string(s1Data))
	}

	{
		d := &CiphertextJSON{
			Key:       "kk",
			Body:      "bb",
			MakeTs:    "12312312311112222345555555",
			EditCount: 1,
			LastTs:    "1111",
		}
		dData, err := json.Marshal(d)
		assert.NoError(t, err)
		b64 := base64.StdEncoding.EncodeToString(dData)
		assert.Equal(t, "eyJrZXkiOiJrayIsImJvZHkiOiJiYiIsIm1ha2VfdHMiOiIxMjMxMjMxMjMxMTExMjIyMjM0NTU1NTU1NSIsImVkaXRfY291bnQiOjEsImxhc3RfdHMiOiIxMTExIn0=", b64)

		b64Data, err := base64.StdEncoding.DecodeString(b64)
		assert.NoError(t, err)

		d2 := &CiphertextJSON{}
		err = json.Unmarshal(b64Data, d2)
		assert.NoError(t, err)
		assert.Equal(t, d, d2)
	}
}
