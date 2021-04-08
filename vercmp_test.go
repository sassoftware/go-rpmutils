package rpmutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVercmp(t *testing.T) {
	// from rpm/tests/rpmvercmp.at
	var values = [][]string{
		{"1.0", "2.0"},
		{"2.0", "2.0.1"},
		{"2.0.1", "2.0.1a"},
		{"5.5p1", "5.5p2"},
		{"5.5p1", "5.5p10"},
		{"10xyz", "10.1xyz"},
		{"xyz10", "xyz10.1"},
		{"xyz.4", "8"},
		{"xyz.4", "2"},
		{"5.5p2", "5.6p1"},
		{"5.6p1", "6.5p1"},
		{"6.0", "6.0.rc1"},
		{"10a2", "10b2"},
		{"1.0a", "1.0aa"},
		{"10.0001", "10.0039"},
		{"4.999.9", "5.0"},
		{"20101121", "20101122"},
		{"1.0~rc1", "1.0"},
		{"1.0~rc1", "1.0~rc2"},
		{"1.0~rc1~git123", "1.0~rc1"},
		// {"1.0", "1.0^"},
		{"1.0", "1.0^git1"},
		{"1.0^git1", "1.0^git2"},
		{"1.0^git1", "1.01"},
		// {"1.0^20160101", "1.0.1"},
		{"1.0^20160101^git1", "1.0^20160102"},
		{"1.0~rc1", "1.0~rc1^git1"},
		{"1.0^git1~pre", "1.0^git1"},
	}
	for _, v := range values {
		assert.Equal(t, -1, Vercmp(v[0], v[1]), "expected: %s should be less than %s", v[0], v[1])
		assert.Equal(t, 1, Vercmp(v[1], v[0]), "expected: %s should be greater than %s", v[1], v[0])
	}
}
