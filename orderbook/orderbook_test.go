package orderbook

import (
	"fmt"
	"reflect"
	"testing"
)

func assert(t *testing.T, a, b any) {
	if !reflect.DeepEqual(a, b) {
		t.Errorf("%+v != %+v", a, b)
	}
}

func TestDeleteOrder(t *testing.T) {
	l := NewLimit(19_000)
	buyOrder := NewOrder(true, 5.0)
	l.AddOrder(buyOrder)

	assert(t, buyOrder.Limit, l)
	assert(t, len(l.Orders), 1)
	assert(t, l.TotalVolume, 5.0)

	l.DeleteOrder(buyOrder)
	assert(t, buyOrder.Limit, (*Limit)(nil))
	assert(t, len(l.Orders), 0)
	assert(t, l.TotalVolume, 0.0)

	buyOrderA := NewOrder(true, 5.0)
	buyOrderB := NewOrder(true, 2.0)
	buyOrderC := NewOrder(true, 3.0)
	l.AddOrder(buyOrderA)
	l.AddOrder(buyOrderB)
	l.AddOrder(buyOrderC)
	l.DeleteOrder(buyOrderA)

	fmt.Println(l.Orders)
}

func TestPlaceLimitOrder(t *testing.T) {
	ob := NewOrderbook()
	buyOrder := NewOrder(true, 5.0)
	ob.PlaceLimitOrder(19_000, buyOrder)

	assert(t, len(ob.asks), 0)
	assert(t, len(ob.bids), 1)
	assert(t, ob.bids[0].TotalVolume, 5.0)
	assert(t, ob.bids[0].Price, 19_000.0)
	assert(t, ob.Orders[buyOrder.ID], buyOrder)
}

func TestPlaceMarketOrder(t *testing.T) {
	ob := NewOrderbook()
	sellOrderA := NewOrder(false, 12.0)
	sellOrderB := NewOrder(false, 5.0)
	ob.PlaceLimitOrder(5_000, sellOrderA)
	ob.PlaceLimitOrder(7_000, sellOrderB)
	buyOrder := NewOrder(true, 14.0)
	matches := ob.PlaceMarketOrder(buyOrder)

	assert(t, len(ob.asks), 1)
	assert(t, ob.asks[0].TotalVolume, 3.0)

	fmt.Println(matches)
}

func TestPlaceMarketOrder2(t *testing.T) {
	ob := NewOrderbook()
	buyOrderA := NewOrder(true, 5.0)
	buyOrderB := NewOrder(true, 3.0)
	buyOrderC := NewOrder(true, 7.0)
	ob.PlaceLimitOrder(19_000, buyOrderA)
	ob.PlaceLimitOrder(15_000, buyOrderB)
	ob.PlaceLimitOrder(12_000, buyOrderC)
	sellOrder := NewOrder(false, 10)
	ob.PlaceMarketOrder(sellOrder)

	assert(t, len(ob.bids), 1)
	assert(t, ob.bids[0].TotalVolume, 5.0)
	assert(t, len(ob.bids[0].Orders), 1)
	fmt.Println(ob.bids)
}

func TestCancelOrder(t *testing.T) {
	ob := NewOrderbook()
	buyOrder := NewOrder(true, 5.0)
	ob.PlaceLimitOrder(5_000, buyOrder)

	assert(t, ob.BidTotalVolume(), 5.0)

	ob.CancelOrder(buyOrder)

	assert(t, ob.BidTotalVolume(), 0.0)

	_, ok := ob.Orders[buyOrder.ID]
	assert(t, ok, false)
}

// func TestLimitSort(t *testing.T) {
// 	ob := NewOrderbook()
// 	sellOrderA := NewOrder(false, 5.0)
// 	sellOrderB := NewOrder(false, 8.0)
// 	sellOrderC := NewOrder(false, 10.0)
// 	ob.PlaceLimitOrder(19_000, sellOrderA)
// 	ob.PlaceLimitOrder(10_000, sellOrderB)
// 	ob.PlaceLimitOrder(5_000, sellOrderC)
// 	ob.Asks()
// 	fmt.Println(ob.asks)
// }
