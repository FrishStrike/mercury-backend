package main

import (
	"fmt"
	catalogv1 "github.com/frishstrike/mercury-backend/api/proto/gen/go/catalog/v1"
	orderv1 "github.com/frishstrike/mercury-backend/api/proto/gen/go/order/v1"
	paymentv1 "github.com/frishstrike/mercury-backend/api/proto/gen/go/payment/v1"
)

func main() {
	// Проверка что импорты работают
	_ = &catalogv1.Product{}
	_ = &orderv1.Order{}
	_ = &paymentv1.Payment{}

	fmt.Println("✅ All proto imports working!")
}
