// Package pagination provides generic helpers for paginating cursor-based
// and offset-based HTTP APIs and for parallelizing batch work.
//
// Use pagination when a Polymarket (or any) endpoint returns paged
// results and you want a tight loop around either the streamed pages or
// the fully collected slice. Helpers are generic over the page item type;
// callers supply the per-page fetch function.
//
// When not to use this package:
//   - For single-page fetches — call the underlying API directly.
//   - When concurrency is not desired and a simple for-loop is clearer.
//
// Stability: Page, OffsetPage, StreamResult, StreamPages, CollectAll,
// CollectOffset, and Batch are part of the polygolem public SDK and
// follow semver.
package pagination

import (
	"context"
	"sync"
)

// Page fetches one page of cursor-based data given a cursor.
// Returns the page's items, the next cursor (empty string ends iteration),
// and any error. The first call receives an empty cursor.
type Page[T any] func(ctx context.Context, cursor string) ([]T, string, error)

// StreamResult is a single page result emitted on the channel returned by
// StreamPages. Exactly one of Items or Err is meaningful per result.
type StreamResult[T any] struct {
	Items []T
	Err   error
}

// StreamPages iterates through all pages of a cursor-based endpoint.
// Calls pageFn until the next cursor is empty or an error occurs.
// Returns a channel that closes when iteration completes or ctx is
// cancelled. Errors are delivered as a final StreamResult before close.
func StreamPages[T any](ctx context.Context, pageFn Page[T]) <-chan StreamResult[T] {
	ch := make(chan StreamResult[T])

	go func() {
		defer close(ch)
		cursor := ""
		for {
			items, next, err := pageFn(ctx, cursor)
			if err != nil {
				select {
				case ch <- StreamResult[T]{Err: err}:
				case <-ctx.Done():
				}
				return
			}
			if len(items) > 0 {
				select {
				case ch <- StreamResult[T]{Items: items}:
				case <-ctx.Done():
					return
				}
			}
			if next == "" {
				return
			}
			cursor = next
		}
	}()

	return ch
}

// CollectAll consumes a cursor-paged stream and returns all items.
// Returns the first error encountered; partial results are discarded.
func CollectAll[T any](ctx context.Context, pageFn Page[T]) ([]T, error) {
	var all []T
	stream := StreamPages(ctx, pageFn)
	for result := range stream {
		if result.Err != nil {
			return nil, result.Err
		}
		all = append(all, result.Items...)
	}
	return all, nil
}

// OffsetPage fetches one page of an offset-based API.
// Returns the page's items, the count returned (used to detect the last
// page), and any error.
type OffsetPage[T any] func(ctx context.Context, offset, limit int) ([]T, int, error)

// CollectOffset iterates through all pages of an offset-based endpoint.
// Stops when pageFn returns fewer than limit items. Returns the first
// error encountered; partial results are discarded.
func CollectOffset[T any](ctx context.Context, pageFn OffsetPage[T], limit int) ([]T, error) {
	var all []T
	offset := 0
	for {
		items, count, err := pageFn(ctx, offset, limit)
		if err != nil {
			return nil, err
		}
		all = append(all, items...)
		if count < limit {
			break
		}
		offset += limit
	}
	return all, nil
}

// Batch splits items into chunks of at most maxBatchSize and runs fn on
// each batch concurrently. Returns the per-batch results in input order.
// Returns the first error encountered. fn may be invoked concurrently;
// callers are responsible for synchronizing any shared state.
func Batch[T, R any](ctx context.Context, items []T, maxBatchSize int, fn func(context.Context, []T) (R, error)) ([]R, error) {
	type result struct {
		idx int
		r   R
		err error
	}

	batches := make([][]T, 0, (len(items)+maxBatchSize-1)/maxBatchSize)
	for i := 0; i < len(items); i += maxBatchSize {
		end := i + maxBatchSize
		if end > len(items) {
			end = len(items)
		}
		batches = append(batches, items[i:end])
	}

	var wg sync.WaitGroup
	results := make([]result, len(batches))

	for i, batch := range batches {
		wg.Add(1)
		go func(idx int, b []T) {
			defer wg.Done()
			r, err := fn(ctx, b)
			results[idx] = result{idx: idx, r: r, err: err}
		}(i, batch)
	}
	wg.Wait()

	out := make([]R, 0, len(results))
	for _, res := range results {
		if res.err != nil {
			return nil, res.err
		}
		out = append(out, res.r)
	}
	return out, nil
}
