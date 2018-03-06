package main

import (
	"net/http"
	"sync"
	"time"
)

// Fruit represents a sweet, fresh fruit.
type Fruit struct {
	Name    string    `json:"name" binding:"required"`
	Origin  string    `json:"origin" description:"Country of origin of the fruit"`
	Price   float64   `json:"price" binding:"required" description:"Price in euros"`
	AddedAt time.Time `json:"added_at" binding:"-" description:"Date of addition of the fruit to the market"`
}

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
	&Fruit{"banana", "ecuador", 2.99, time.Now()},
	&Fruit{"apricot", "france", 4.50, time.Now()},
	&Fruit{"mango", "senegal", 6.99, time.Now()},
	&Fruit{"litchi", "china", 5.65, time.Now()},
	&Fruit{"apple", "france", 2.49, time.Now()},
	&Fruit{"peach", "spain", 3.20, time.Now()},
	&Fruit{"peach", "spain", 3.20, time.Now()},
}

func main() {
	srv := &http.Server{
		Addr:    ":48879",
		Handler: NewRouter(),
	}
	srv.ListenAndServe()
}
