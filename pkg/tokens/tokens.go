// Package tokens provides secure token generation for GOSP authentication.
// It generates human-readable join tokens for worker-cluster authentication.
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
	"hotel", "india", "juliett", "kilo", "lima", "mike", "november", "oscar",
	"papa", "quebec", "romeo", "sierra", "tango", "uniform", "victor", "whiskey",
	"xray", "yankee", "zulu", "amber", "bronze", "copper", "diamond", "emerald",
	"flame", "glimmer", "heaven", "indigo", "jade", "karate", "lunar", "magic",
	"night", "opal", "pearl", "ruby", "storm", "titan", "unity", "valley",
	"whale", "zenith", "arc", "bolt", "crane", "drift", "edge", "flux", "glow",
	"haze", "ion", "jolt", "kite", "lynx", "moth", "nova", "orb", "pulse",
	"quark", "rift", "snag", "tide", "urge", "vibe", "warp", "zone",
}

// Generate creates a human-readable join token consisting of three random words.
// The token is cryptographically secure and suitable for worker authentication.
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
