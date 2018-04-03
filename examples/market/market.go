package main

import (
	"log"
	"net/http"
	"sync"
	"time"
)

// Fruit represents a sweet, fresh fruit.
type Fruit struct {
	Name    string    `json:"name" validate:"required"`
	Origin  string    `json:"origin" validate:"required" description:"Country of origin of the fruit" enum:"ecuador,france,senegal,china,spain"`
	Price   float64   `json:"price" validate:"required" description:"Price in euros"`
	AddedAt time.Time `json:"-" binding:"-" description:"Date of addition of the fruit to the market"`
}

// Type implements openapi.TypeNamer for Fruit.
func (f *Fruit) Type() string { return "RottenFruit" }

// Market is a fruit market.
type Market struct {
	fruits map[string]*Fruit
	sync.RWMutex
}

var market *Market

func init() {
	market = &Market{
		fruits:  make(map[string]*Fruit),
		RWMutex: sync.RWMutex{},
	}
	for _, f := range fruits {
		market.fruits[f.Name] = f
	}
}

var fruits = []*Fruit{
	{"banana", "ecuador", 2.99, time.Now()},
	{"apricot", "france", 4.50, time.Now()},
	{"mango", "senegal", 6.99, time.Now()},
	{"litchi", "china", 5.65, time.Now()},
	{"apple", "france", 2.49, time.Now()},
	{"peach", "spain", 3.20, time.Now()},
	{"peach", "spain", 3.20, time.Now()},
}

func main() {
	router, err := NewRouter()
	if err != nil {
		log.Fatal(err)
	}
	srv := &http.Server{
		Addr:    ":4242",
		Handler: router,
	}
	srv.ListenAndServe()
}
