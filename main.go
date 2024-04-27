package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/imorugiy/crypto-exchange/orderbook"
	"github.com/labstack/echo/v4"
)

type OrderType string

const (
	MarketOrder OrderType = "MARKET"
	LimitOrder  OrderType = "LIMIT"
)

type Market string

const (
	MarketETH Market = "ETH"
	MarketBTC Market = "BTC"
)

type Exchange struct {
	orderbooks map[Market]*orderbook.Orderbook
}

func NewExchange() *Exchange {
	ordersbooks := make(map[Market]*orderbook.Orderbook)
	ordersbooks[MarketETH] = orderbook.NewOrderbook()

	return &Exchange{
		orderbooks: ordersbooks,
	}
}

func main() {
	ex := NewExchange()
	e := echo.New()

	e.GET("/health", handleHealth)
	e.POST("/order", ex.handlePlaceOrder)
	e.GET("/books/:market", ex.handleGetBook)
	e.DELETE("/order/:id", ex.handleCancelOrder)

	e.Start(":3000")
}

func handleHealth(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]any{"status": "healthy"})
}

type Order struct {
	ID        int
	Bid       bool
	Size      float64
	Price     float64
	Timestamp int64
}

type OrderbookData struct {
	TotalBidVolume float64
	TotalAskVolume float64
	Asks           []*Order
	Bids           []*Order
}

func (ex *Exchange) handleGetBook(c echo.Context) error {
	market := Market(c.Param("market"))
	ob := ex.orderbooks[market]

	data := OrderbookData{
		TotalAskVolume: ob.AskTotalVolume(),
		TotalBidVolume: ob.BidTotalVolume(),
		Asks:           []*Order{},
		Bids:           []*Order{},
	}

	for _, limit := range ob.Asks() {
		for _, order := range limit.Orders {
			o := Order{
				ID:        order.ID,
				Bid:       order.Bid,
				Size:      order.Size,
				Price:     limit.Price,
				Timestamp: order.Timestamp,
			}
			data.Asks = append(data.Asks, &o)
		}
	}

	for _, limit := range ob.Bids() {
		for _, order := range limit.Orders {
			o := Order{
				ID:        order.ID,
				Bid:       order.Bid,
				Size:      order.Size,
				Price:     limit.Price,
				Timestamp: order.Timestamp,
			}
			data.Bids = append(data.Bids, &o)
		}
	}

	return c.JSON(http.StatusOK, data)
}

type placeOrderRequest struct {
	Market Market
	Type   OrderType // limit or market
	Size   float64
	Price  float64
	Bid    bool
}

type MatchedOrder struct {
	ID    int
	Price float64
	Size  float64
}

func (ex *Exchange) handlePlaceOrder(c echo.Context) error {
	var placeOrderRequest placeOrderRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&placeOrderRequest); err != nil {
		return err
	}

	ob := ex.orderbooks[placeOrderRequest.Market]
	o := orderbook.NewOrder(placeOrderRequest.Bid, placeOrderRequest.Size)

	if placeOrderRequest.Type == LimitOrder {
		ob.PlaceLimitOrder(placeOrderRequest.Price, o)
		return c.JSON(http.StatusOK, map[string]any{"status": "limit order placed"})
	}

	if placeOrderRequest.Type == MarketOrder {
		matches := ob.PlaceMarketOrder(o)
		matchedOrders := make([]MatchedOrder, len(matches))
		for i := 0; i < len(matches); i++ {
			var id int
			if matches[i].Ask == o {
				id = matches[i].Bid.ID
			} else {
				id = matches[i].Ask.ID
			}
			matchedOrders[i].ID = id
			matchedOrders[i].Price = matches[i].Price
			matchedOrders[i].Size = matches[i].SizeFilled
		}
		return c.JSON(http.StatusOK, map[string]any{"matches": matchedOrders})
	}

	return nil
}

func (ex *Exchange) handleCancelOrder(c echo.Context) error {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	ob := ex.orderbooks[MarketETH]
	o := ob.Orders[id]
	ob.CancelOrder(o)

	return c.JSON(http.StatusOK, map[string]any{"status": "order canceled"})
}
