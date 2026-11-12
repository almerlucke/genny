package counterpoint

import "testing"

// FuzzGenerate: any cantus firmus (bytes as pitches) never panics Generate or Check, and a line that
// comes without error passes the rules
func FuzzGenerate(f *testing.F) {
	f.Add([]byte{60, 62, 64, 62, 60}, byte(0), byte(0))
	f.Add([]byte{}, byte(1), byte(0))
	f.Add([]byte{60}, byte(0), byte(3))
	f.Add([]byte{0, 127, 0, 127, 0, 127}, byte(1), byte(200))
	f.Fuzz(func(t *testing.T, cantus []byte, side byte, leap byte) {
		if len(cantus) > 40 {
			cantus = cantus[:40]
		}
		cf := make([]int, len(cantus))
		for i, b := range cantus {
			cf[i] = int(b)
		}
		cfg := Config{Side: Side(side % 2), MaxLeap: int(leap) % 13, Budget: 20000}
		cp, err := Generate(cf, cfg)
		if err == nil {
			if len(cp) != len(cf) {
				t.Fatalf("length %d for a cantus of %d", len(cp), len(cf))
			}
			if v := cfg.Check(cf, cp); len(v) > 0 {
				t.Fatalf("own line breaks the rules: %v", v)
			}
		}
		GenerateSecond(cf, cfg)
		Check(cf, cp)
		CheckSecond(cf, cp)
	})
}
