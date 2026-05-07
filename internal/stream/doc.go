// Package stream provides typed WebSocket clients for Polymarket CLOB
// market streams with reconnect and event deduplication.
//
// The market client subscribes to public order book and trade updates for
// a set of asset IDs. Authenticated user streams are not implemented here.
// Start with MarketClient and SubscribeAssets for orientation.
//
// This package is internal and not part of the polygolem public SDK.
package stream
