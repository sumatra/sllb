package sllb

import (
	"bytes"
	"encoding/gob"
	"errors"
	"math"

	metro "github.com/dgryski/go-metro"
)

var (
	exp32 = math.Pow(2, 32)
)

/*
Sketch adapts the LogLog-Beta algorithm basen on Flajolet et. al to the
data stream processing by adding a sliding window mechanism. It has the
advantage to estimate at any time the number of flows seen over any duration
bounded by the length of the sliding window.
*/
type Sketch struct {
	window          uint64
	alpha           float64
	p               uint
	m               uint
	regs            []*reg
	latestTimestamp uint64
}

/*
New return a new Sketch with errorRate
*/
func New(errRate float64) (*Sketch, error) {
	if !(0 < errRate && errRate < 1) {
		return nil, errors.New("errRate must be between 0 and 1")
	}
	p := uint(math.Ceil(math.Log2(math.Pow((1.04 / errRate), 2))))
	m := uint(1) << p
	sk := &Sketch{
		p:     p,
		m:     m,
		regs:  make([]*reg, m, m),
		alpha: alpha(float64(m)),
	}
	for i := range sk.regs {
		sk.regs[i] = newReg()
	}
	return sk, nil
}

// NewDefault returns a sketch with errorRate 0.008
func NewDefault() *Sketch {
	sk, _ := New(0.008)
	return sk
}

func (sk *Sketch) valAndPos(value []byte) (uint8, uint64) {
	val := metro.Hash64(value, 32)
	k := 64 - sk.p
	j := val >> uint(k)
	R := rho(val<<sk.p, 6)
	return R, j
}

/*
Insert a value with a timestamp to the Sketch.
*/
func (sk *Sketch) Insert(timestamp uint64, value []byte) {
	R, j := sk.valAndPos(value)
	sk.regs[j].insert(tR{timestamp, R})
	if timestamp > sk.latestTimestamp {
		sk.latestTimestamp = timestamp
	}
}

/*
Estimate returns the estimated cardinality since a given timestamp
*/
func (sk *Sketch) Estimate(timestamp uint64) uint64 {
	m := float64(sk.m)
	sum := regSumSince(sk.regs, timestamp)
	ez := zerosSince(sk.regs, timestamp)
	beta := beta(ez)
	return uint64(sk.alpha * m * (m - ez) / (beta + sum))
}

// Encode Sketch into a gob
func (h *Sketch) GobEncode() ([]byte, error) {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(h.window); err != nil {
		return nil, err
	}
	if err := enc.Encode(h.alpha); err != nil {
		return nil, err
	}
	if err := enc.Encode(h.p); err != nil {
		return nil, err
	}
	if err := enc.Encode(h.m); err != nil {
		return nil, err
	}

	for i := range h.regs {
		if err := enc.Encode(h.regs[i].lfpm); err != nil {
			return nil, err
		}
	}
	if err := enc.Encode(h.latestTimestamp); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Decode gob into a Sketch structure
func (h *Sketch) GobDecode(b []byte) error {
	dec := gob.NewDecoder(bytes.NewBuffer(b))
	if err := dec.Decode(&h.window); err != nil {
		return err
	}
	if err := dec.Decode(&h.alpha); err != nil {
		return err
	}
	if err := dec.Decode(&h.p); err != nil {
		return err
	}
	if err := dec.Decode(&h.m); err != nil {
		return err
	}
	h.regs = make([]*reg, h.m, h.m)
	for idx := uint(0); idx < h.m; idx++ {
		h.regs[idx] = newReg()
		if err := dec.Decode(&h.regs[idx].lfpm); err != nil {
			return err
		}
	}
	if err := dec.Decode(&h.latestTimestamp); err != nil {
		return err
	}

	return nil
}