package dga

import (
	"strings"
	"testing"
	"time"
)

func TestGenerateDomain(t *testing.T) {
	domain := GenerateDomain(time.Now())
	if !strings.HasSuffix(domain, ".com") {
		t.Errorf("domain should end with .com, got: %s", domain)
	}
	if !strings.HasPrefix(domain, "charm-") {
		t.Errorf("domain should start with charm-, got: %s", domain)
	}
}

func TestGenerateDomainDeterministic(t *testing.T) {
	now := time.Now()
	d1 := GenerateDomain(now)
	d2 := GenerateDomain(now)
	if d1 != d2 {
		t.Errorf("same time should produce same domain: %s vs %s", d1, d2)
	}
}

func TestGenerateDomainDifferentDays(t *testing.T) {
	d1 := GenerateDomain(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	d2 := GenerateDomain(time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC))
	if d1 == d2 {
		t.Error("different days should produce different domains")
	}
}

func TestToday(t *testing.T) {
	domain := Today()
	if !strings.Contains(domain, ".com") {
		t.Errorf("invalid domain: %s", domain)
	}
}

func TestGenerateDomains(t *testing.T) {
	domains := GenerateDomains(7)
	if len(domains) != 7 {
		t.Errorf("expected 7 domains, got %d", len(domains))
	}

	seen := make(map[string]bool)
	for _, d := range domains {
		if seen[d] {
			t.Errorf("duplicate domain: %s", d)
		}
		seen[d] = true
	}
}

func TestIsValidDomain(t *testing.T) {
	today := Today()
	if !IsValidDomain(today) {
		t.Errorf("today's domain should be valid: %s", today)
	}

	if IsValidDomain("evil.com") {
		t.Error("evil.com should not be valid")
	}
}

func TestDomainFormat(t *testing.T) {
	domain := Today()
	parts := strings.Split(domain, "-")
	if len(parts) < 3 {
		t.Errorf("domain should have multiple parts: %s", domain)
	}
}
