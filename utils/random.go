package utils

import (
	"math/rand"

	"github.com/google/uuid"
)

func RNG(min, max int) int {
	return rand.Intn(max-min+1) + min
}

func RandomString(strs []string) string {
	return strs[RNG(0, len(strs)-1)]
}

func RandomItem[T any](items []T) T {
	return items[RNG(0, len(items)-1)]
}

func RandomBool() bool {
	return rand.Intn(2) == 1
}

func RandomFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func GenerateOrderID() string {
	return uuid.New().String()
}
