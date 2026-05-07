// Package pagination provides auto-pagination helpers for cursor and offset-based APIs.
// Stolen from polymarket-go-sdk's StreamData and MarketsAll patterns.
package pagination

import (
	"context"
	"sync"
)

// Page fetches one page of data given a cursor.
// Returns the items, next cursor, and error.
type Page[T any] func(ctx context.Context, cursor string) ([]T, string, error)

// StreamResult is a single page result in a stream.
type StreamResult[T any] struct {
	Items []T
	Err   error
}

// StreamPages iterates through all pages of a cursor-based endpoint.
// Calls Page until nextCursor is empty or an error occurs.
// Returns a channel that closes when all pages are consumed or ctx is cancelled.
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

// CollectAll consumes a page stream and returns all items.
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
type OffsetPage[T any] func(ctx context.Context, offset, limit int) ([]T, int, error)

// CollectOffset iterates through all pages of an offset-based endpoint.
// Returns items and total count when the page returns fewer than limit items.
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

// Batch splits a slice into batches of maxBatchSize and processes each batch concurrently.
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
