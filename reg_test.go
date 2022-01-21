package sllb

import (
	"testing"
)

type testStruct struct {
	name  string
	tr    tR
	exptr []tR
}

var testData = []*testStruct{
	&testStruct{name: "A", tr: tR{T: 0, R: 5}, exptr: []tR{tR{T: 0, R: 5}}},
	&testStruct{name: "B", tr: tR{T: 1, R: 3}, exptr: []tR{tR{T: 0, R: 5}, tR{T: 1, R: 3}}},
	&testStruct{name: "C", tr: tR{T: 2, R: 4}, exptr: []tR{tR{T: 0, R: 5}, tR{T: 2, R: 4}}},
	&testStruct{name: "D", tr: tR{T: 3, R: 2}, exptr: []tR{tR{T: 0, R: 5}, tR{T: 2, R: 4}, tR{T: 3, R: 2}}},
	&testStruct{name: "E", tr: tR{T: 4, R: 1}, exptr: []tR{tR{T: 0, R: 5}, tR{T: 2, R: 4}, tR{T: 3, R: 2}, tR{T: 4, R: 1}}},
	&testStruct{name: "F", tr: tR{T: 5, R: 6}, exptr: []tR{tR{T: 5, R: 6}}},
}

func testEq(a, b []tR) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestRegInsert(t *testing.T) {
	r := newReg()
	for _, td := range testData {
		r.insert(td.tr)
		if !testEq(r.lfpm, td.exptr) {
			t.Errorf("expected %v, got %v", td.exptr, r.lfpm)
		}
	}
}
