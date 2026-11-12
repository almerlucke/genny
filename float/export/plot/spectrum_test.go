package plot

import (
	"math"
	"testing"
)

func TestFFTSinePeaksAtItsBin(t *testing.T) {
	const n, bin = 256, 13
	re, im := make([]float64, n), make([]float64, n)
	for i := range re {
		re[i] = math.Sin(2 * math.Pi * bin * float64(i) / n)
	}
	fft(re, im)
	for k := 0; k < n; k++ {
		mag := math.Hypot(re[k], im[k])
		want := 0.0
		if k == bin || k == n-bin {
			want = n / 2
		}
		if math.Abs(mag-want) > 1e-9 {
			t.Fatalf("bin %d: magnitude %g, want %g", k, mag, want)
		}
	}
}

func TestFFTParseval(t *testing.T) {
	const n = 128
	re, im := make([]float64, n), make([]float64, n)
	energy := 0.0
	for i := range re {
		re[i] = math.Sin(float64(i)*0.37) + 0.3*math.Cos(float64(i*i)*0.01)
		energy += re[i] * re[i]
	}
	fft(re, im)
	spec := 0.0
	for k := range re {
		spec += re[k]*re[k] + im[k]*im[k]
	}
	if math.Abs(spec/n-energy) > 1e-9*energy {
		t.Errorf("spectral energy %g, time energy %g", spec/n, energy)
	}
}

func TestMagnitudeSpectrum(t *testing.T) {
	const n, rate = 512, 8000.0
	samples := make([]float64, n)
	for i := range samples {
		samples[i] = 0.5 * math.Sin(2*math.Pi*1000*float64(i)/rate) // bin 64
	}
	for _, hann := range []bool{false, true} {
		freqs, mags := MagnitudeSpectrum(samples, rate, hann)
		if len(freqs) != n/2+1 || freqs[len(freqs)-1] != rate/2 {
			t.Fatalf("hann=%v: %d bins, last %g", hann, len(freqs), freqs[len(freqs)-1])
		}
		peak := 0
		for k := range mags {
			if mags[k] > mags[peak] {
				peak = k
			}
		}
		if peak != 64 || freqs[peak] != 1000 || math.Abs(mags[peak]-0.5) > 1e-9 {
			t.Errorf("hann=%v: peak at bin %d (%g Hz) of %g, want bin 64, 1000 Hz, 0.5", hann, peak, freqs[peak], mags[peak])
		}
	}
	// zero padding to a power of two keeps the amplitude scale
	if f, _ := MagnitudeSpectrum(make([]float64, 100), rate, false); len(f) != 65 {
		t.Errorf("100 samples should pad to 128: %d bins", len(f))
	}
	if f, m := MagnitudeSpectrum(nil, rate, false); f != nil || m != nil {
		t.Error("empty input should give nothing")
	}
}

func TestNewSpectrumOptions(t *testing.T) {
	samples := make([]float64, 64)
	for i := range samples {
		samples[i] = math.Sin(2 * math.Pi * 4 * float64(i) / 64)
	}
	l, err := NewSpectrum(samples, 64, SpectrumOptions{DB: true, LogFreq: true})
	if err != nil {
		t.Fatal(err)
	}
	if x, _ := l.XY(0); x != 1 {
		t.Errorf("log spectrum should start at bin 1, got %g Hz", x)
	}
	if _, y := l.XY(3); math.Abs(y) > 1e-9 { // 4 Hz, amplitude 1 = 0 dB
		t.Errorf("sine of amplitude 1 should be 0 dB, got %g", y)
	}
	if _, y := l.XY(10); y != -120 {
		t.Errorf("silence should sit at the floor, got %g", y)
	}
	for _, bad := range []struct {
		s    []float64
		rate float64
	}{{nil, 1}, {samples, 0}, {[]float64{math.NaN()}, 1}} {
		if _, err := NewSpectrum(bad.s, bad.rate, SpectrumOptions{}); err == nil {
			t.Errorf("%v at %g should fail", bad.s, bad.rate)
		}
	}
}

func TestLogTicks(t *testing.T) {
	var labels []string
	for _, tk := range (LogTicks{}).Ticks(1, 10000) {
		if tk.Label != "" {
			labels = append(labels, tk.Label)
		}
	}
	want := []string{"1", "10", "100", "1k", "10k"}
	if len(labels) != len(want) {
		t.Fatalf("labels %v, want %v", labels, want)
	}
	for i := range want {
		if labels[i] != want[i] {
			t.Errorf("labels %v, want %v", labels, want)
		}
	}
	// minor ticks are unlabelled
	if ticks := (LogTicks{}).Ticks(1, 100); len(ticks) != 19 {
		t.Errorf("1..100 should have 19 ticks, got %d", len(ticks))
	}
	// less than a decade: 2 and 5 are labelled
	n := 0
	for _, tk := range (LogTicks{}).Ticks(150, 900) {
		if tk.Label != "" {
			n++
		}
	}
	if n != 2 { // 200 and 500
		t.Errorf("a narrow range should label more ticks, got %d", n)
	}
	for _, s := range []struct {
		v    float64
		want string
	}{{0.1, "0.1"}, {0.01, "0.01"}, {2500, "2.5k"}, {1e6, "1M"}, {1e-5, "1e-05"}} {
		if got := formatLog(s.v); got != s.want {
			t.Errorf("formatLog(%g) = %s, want %s", s.v, got, s.want)
		}
	}
	if (LogTicks{}).Ticks(0, 10) != nil {
		t.Error("a range from zero has no log ticks")
	}
}

func TestLogMapping(t *testing.T) {
	if f := frac(100, 1, 10000, true); math.Abs(f-0.5) > 1e-12 {
		t.Errorf("100 is halfway between 1 and 10000: %g", f)
	}
	lo, hi := resolve(Axis{Scale: Log, Min: math.Inf(1), Max: math.Inf(-1)}, 20, 20000, true)
	if lo >= 20 || hi <= 20000 || lo <= 0 {
		t.Errorf("log range %g..%g", lo, hi)
	}
}
