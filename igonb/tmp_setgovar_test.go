package igonb

import (
    "testing"

    "github.com/HazelnutParadise/idensyra/internal"
)

func TestSetGoVariableSliceConversion(t *testing.T) {
    exec, err := NewExecutorWithSymbols(internal.Symbols)
    if err != nil {
        t.Fatalf("new executor: %v", err)
    }
    if _, err := exec.runGoSegment("b := []int{1,2,3}", false); err != nil {
        t.Fatalf("define b: %v", err)
    }
    if err := exec.setGoVariable("b", []any{float64(1), float64(2), float64(3)}); err != nil {
        t.Fatalf("setGoVariable: %v", err)
    }
}
