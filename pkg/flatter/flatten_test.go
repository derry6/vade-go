package flatter

import (
    "testing"
)

var (
    testMap = map[string]interface{}{
        "a": map[string]interface{}{
            "b": "90-13",
            "c": 100,
            "d": map[string]string{
                "x": "3841083",
                "y": "234d",
            },
            "e": 12.0,
        },
    }
)

func TestFlatten(t *testing.T) {
    out := map[string]interface{}{}
    if err := Flatten(testMap, out, "", false); err != nil {
        t.Error(err)
    }
    if len(out) != 5 {
        t.Errorf("length of out is %d, expect 5", len(out))
    }
}

func TestFlattenReflect(t *testing.T) {
    out := map[string]interface{}{}
    if err := Flatten(testMap, out, "",true); err != nil {
        t.Error(err)
    }
    if len(out) != 5 {
        t.Errorf("length of out is %d, expect 5", len(out))
    }
}

func BenchmarkFlatten(b *testing.B) {
    out := map[string]interface{}{}
    b.Run("Normal", func(b *testing.B) {
        b.ResetTimer()
        for i := 0; i < b.N; i++ {
            if err := Flatten(testMap, out, "", false); err != nil {
                b.Error(err)
            }
        }
    })
    b.Run("Reflect", func(b *testing.B) {
        b.ResetTimer()
        for i := 0; i < b.N; i++ {
            if err := Flatten(testMap, out, "", true); err != nil {
                b.Error(err)
            }
        }
    })

}