package main

import (
	"sync"
	"time"
)

func main() {
	const routines = 1_000_000
	wg := sync.WaitGroup{}
	for i := 0; i < routines; i++ {
		wg.Add(1)
		go func() {
			auth := makeAuth()
			for j := 0; j < 30; j++ {
				time.Sleep(1 * time.Second)
				auth.Amount = int64(j)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

type Authorization struct {
	ID       string
	Currency string
	Amount   int64

	IdempotencyKey string
	Account        string

	ZoozPaymentID     string
	ZoozTransactionID string

	Status string
}

func makeAuth() Authorization {
	return Authorization{
		ID:       "id",
		Currency: "RUB",
		Amount:   1000,

		IdempotencyKey: "super-duper-long-idempotency-key",
		Account:        "another-surprisingly-long-account",

		ZoozPaymentID:     "cfe1cfee-c476-4e22-9044-736bc7202890",
		ZoozTransactionID: "494a40ad-ec9d-406d-ac13-7a51014df246",

		Status: "Sorry",
	}
}
