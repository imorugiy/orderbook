package main

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
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
	client     *ethclient.Client
	orderbooks map[Market]*orderbook.Orderbook
	users      map[int]*User
}

func NewExchange() (*Exchange, error) {
	client, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		return nil, err
	}

	ordersbooks := make(map[Market]*orderbook.Orderbook)
	ordersbooks[MarketETH] = orderbook.NewOrderbook()

	return &Exchange{
		client:     client,
		orderbooks: ordersbooks,
		users:      make(map[int]*User),
	}, nil
}

type User struct {
	ID         int
	PrivateKey *ecdsa.PrivateKey
}

func NewUser(id int, privKey string) *User {
	privateKey, err := crypto.HexToECDSA(privKey)
	if err != nil {
		panic(err)
	}

	return &User{
		ID:         id,
		PrivateKey: privateKey,
	}
}

func main() {
	ex, err := NewExchange()
	if err != nil {
		panic(err)
	}
	user6 := NewUser(6, "e485d098507f54e7733a205420dfddbe58db035fa577fc294ebd14db90767a52")
	ex.users[6] = user6
	user7 := NewUser(7, "a453611d9419d0e56f499079478fd72c37b251a94bfde4d19872c44cf65386e3")
	ex.users[7] = user7
	user8 := NewUser(8, "829e924fdf021ba3dbbc4225edfece9aca04b929d6e75613329ca6f1d31c0bb4")
	ex.users[8] = user8

	getBalance(ex.client, crypto.PubkeyToAddress(user6.PrivateKey.PublicKey).Hex())
	getBalance(ex.client, crypto.PubkeyToAddress(user7.PrivateKey.PublicKey).Hex())
	getBalance(ex.client, crypto.PubkeyToAddress(user8.PrivateKey.PublicKey).Hex())

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
	UserID    int
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
				UserID:    order.UserID,
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
				UserID:    order.UserID,
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
	UserID int
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

	market := placeOrderRequest.Market
	o := orderbook.NewOrder(placeOrderRequest.Bid, placeOrderRequest.Size, placeOrderRequest.UserID)

	if placeOrderRequest.Type == LimitOrder {
		if err := ex.handlePlaceLimitOrder(market, placeOrderRequest.Price, o); err != nil {
			return err
		}
		return c.JSON(http.StatusOK, map[string]any{"status": "limit order placed"})
	}

	if placeOrderRequest.Type == MarketOrder {
		matches, matchedOrders := ex.handlePlaceMarketOrder(market, o)
		if err := ex.handleMatches(matches); err != nil {
			return err
		}
		return c.JSON(http.StatusOK, map[string]any{"matches": matchedOrders})
	}

	return nil
}

func (ex *Exchange) handlePlaceLimitOrder(market Market, price float64, order *orderbook.Order) error {
	ob := ex.orderbooks[market]
	ob.PlaceLimitOrder(price, order)
	return nil
}

func (ex *Exchange) handlePlaceMarketOrder(market Market, order *orderbook.Order) ([]orderbook.Match, []MatchedOrder) {
	ob := ex.orderbooks[market]
	matches := ob.PlaceMarketOrder(order)
	matchedOrders := make([]MatchedOrder, len(matches))
	for i := 0; i < len(matches); i++ {
		var id int
		if matches[i].Ask == order {
			id = matches[i].Bid.ID
		} else {
			id = matches[i].Ask.ID
		}
		matchedOrders[i].ID = id
		matchedOrders[i].Price = matches[i].Price
		matchedOrders[i].Size = matches[i].SizeFilled
	}

	return matches, matchedOrders
}

func (ex *Exchange) handleMatches(matches []orderbook.Match) error {
	for _, match := range matches {
		from, ok := ex.users[match.Ask.UserID]
		if !ok {
			return fmt.Errorf("user %d not found", match.Ask.UserID)
		}
		to, ok := ex.users[match.Bid.UserID]
		if !ok {
			return fmt.Errorf("user %d not found", match.Bid.UserID)
		}
		transferEth(ex.client, from.PrivateKey, to.PrivateKey.PublicKey, int64(match.SizeFilled))
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
