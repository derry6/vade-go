package source

import (
    "math"
    "sort"
    "testing"
)

// TestConfigPathUpdate tests config path Update
func TestConfigPathUpdate(t *testing.T) {
    type TestCase struct {
        v1      map[string]interface{}
        v2      map[string]interface{}
        total   int
        created int
        updated int
        deleted int
    }
    var testCases = []TestCase{
        {
            v1: map[string]interface{}{},
            v2: map[string]interface{}{"a":1, "b":2},
            total: 2,
            created: 2,
        },
        {
            v1: map[string]interface{}{"a":1, "b": 2},
            v2: map[string]interface{}{},
            total: 2,
            deleted: 2,
        },

        {
            // created
            v1:      map[string]interface{}{"a": 1},
            v2:      map[string]interface{}{"a": 1, "b": 2, "c": 3},
            total:   2,
            created: 2,
        },
        {
            // updated
            v1:      map[string]interface{}{"a": 1, "b": 2, "c": 3},
            v2:      map[string]interface{}{"a": 2, "b": 2, "c": 4},
            total:   2,
            updated: 2,
        },
        {
            // deleted
            v1:      map[string]interface{}{"a": 1, "b": 2, "c": 3},
            v2:      map[string]interface{}{"b": 2},
            total:   2,
            deleted: 2,
        },
        {
            // created, updated
            v1:      map[string]interface{}{"a": 1, "b": 2, "c": 3},
            v2:      map[string]interface{}{"a": 3, "b": 2, "c": 1, "d": 4},
            total:   3,
            created: 1,
            updated: 2,
        },
        {
            // created, updated, deleted
            v1:      map[string]interface{}{"a": 1, "b": 3, "c": 3},
            v2:      map[string]interface{}{"a": 1, "b": 2, "d": 4},
            total:   3,
            created: 1,
            updated: 1,
            deleted: 1,
        },
    }
    // v1, v2, n,
    for i, testCase := range testCases {
        n := &pathStore{values: testCase.v1}
        events := n.Update(testCase.v2, true)
        if total := len(events); total != testCase.total {
            t.Errorf("total events of testCase %d is %d, expect %d", i, total, testCase.total)
        }
        created, updated, deleted := 0, 0, 0
        for _, e := range events {
            switch e.Action {
            case Created:
                created++
            case Updated:
                updated++
            case Deleted:
                deleted++
            default:
                t.Errorf("unknown event action of testCase %d: %v", i, e.Action)
            }
        }
        if created != testCase.created {
            t.Logf("created events of testCase %d is %d, expect %d", i, created, testCase.created)
        }

        if updated != testCase.updated {
            t.Logf("updated events of testCase %d is %d, expect %d", i, updated, testCase.updated)
        }

        if deleted != testCase.deleted {
            t.Logf("deleted events of testCase %d is %d, expect %d", i, deleted, testCase.deleted)
        }
    }
}

func TestConfigPathSort(t *testing.T) {
    namespaces := []*pathStore{
        { "n1", 1,nil, nil },
        { "n2", 5,nil, nil },
        { "n3", 9,nil, nil },
        { "n4", 9,nil, nil },
        { "n5", 4,nil, nil },
        { "n6", 2,nil, nil },
        { "n7", 0,nil, nil },
    }
    sort.Sort(pathHighToLow(namespaces))
    last := math.MaxInt32
    for _, n := range namespaces {
        if n.pri > last {
            t.Errorf("sort error")
        }
        last = n.pri
    }
}