package tokens

import (
	"crypto/rand"
	"math/big"
	"strings"
)

var words = []string{
	"apple", "banana", "cherry", "dragon", "eagle", "forest", "ghost", "hammer",
	"island", "joker", "knight", "lemon", "mountain", "nebula", "ocean", "planet",
	"quartz", "river", "silver", "tiger", "ultra", "vortex", "winter", "xray",
	"yellow", "zebra", "alpha", "bravo", "delta", "echo", "foxtrot", "golf",
}

// Generate human-readable join token
func Generate() (string, error) {
	var result []string
	for i := 0; i < 3; i++ {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(words))))
		if err != nil {
			return "", err
		}
		result = append(result, words[idx.Int64()])
	}
	return strings.Join(result, "-"), nil
}
