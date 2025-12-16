package business

import (
	"fmt"
	"math/rand"
	"time"
)

func GenerateTenantID() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("BIZ%04d", rand.Intn(10000))
}
