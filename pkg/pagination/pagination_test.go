package pagination

import (
	"context"
	"testing"
)

func TestStreamPages(t *testing.T) {
	calls := 0
	fn := func(ctx context.Context, cursor string) ([]string, string, error) {
		calls++
		switch cursor {
		case "":
			return []string{"a", "b"}, "page2", nil
		case "page2":
			return []string{"c"}, "", nil
		default:
			return nil, "", nil
		}
	}

	var all []string
	for result := range StreamPages(context.Background(), fn) {
		if result.Err != nil {
			t.Fatal(result.Err)
		}
		all = append(all, result.Items...)
	}
	if len(all) != 3 {
		t.Fatalf("expected 3 items: %v", all)
	}
	if calls != 2 {
		t.Fatalf("expected 2 calls: %d", calls)
	}
}

func TestCollectAll(t *testing.T) {
	fn := func(ctx context.Context, cursor string) ([]int, string, error) {
		if cursor == "" {
			return []int{1, 2, 3}, "", nil
		}
		return nil, "", nil
	}

	all, err := CollectAll(context.Background(), fn)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 3 {
		t.Fatalf("expected 3: %v", all)
	}
}

func TestCollectOffset(t *testing.T) {
	fn := func(ctx context.Context, offset, limit int) ([]int, int, error) {
		if offset == 0 {
			return []int{1, 2}, 2, nil
		}
		if offset == 1 {
			return []int{3}, 1, nil
		}
		return nil, 0, nil
	}

	all, err := CollectOffset(context.Background(), fn, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 3 {
		t.Fatalf("expected 3: %v", all)
	}
}

func TestBatch(t *testing.T) {
	items := []string{"a", "b", "c", "d", "e"}
	results, err := Batch(context.Background(), items, 2, func(ctx context.Context, batch []string) (string, error) {
		return batch[0], nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 batches: %v", results)
	}
}
