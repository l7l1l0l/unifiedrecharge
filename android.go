package unifiedrecharge

import (
	"context"
	ap "google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"math"
	"net/http"
	"sync"
	"time"
	"unsafe"
)

//call androidpushliser v3
var (
	once            sync.Once
	purchaseService *ap.PurchasesProductsService
	lastError       error
)

const (
	maxTryCount = 5
)

func getService(configFile string) *ap.PurchasesProductsService {
	once.Do(func() {
		ctx := context.Background()
		service, err := ap.NewService(ctx, option.WithCredentialsFile(configFile))
		if err != nil {
			lastError = err
			return
		}

		purchaseService = ap.NewPurchasesProductsService(service)
	})

	return purchaseService
}

func CheckGoogleAndroidPurchase(configFile, packageName, productId, token string) (productPurchase *ap.ProductPurchase, err error) {
	service := getService(configFile)
	if service == nil {
		return nil, lastError
	}

	for i := 0; i < maxTryCount; i++ {
		productPurchase, err = service.Get(packageName, productId, token).Do()
		if err != nil {
			//check code, if server error
			googleErr := (*googleapi.Error)(unsafe.Pointer(&err))
			if googleErr.Code == http.StatusInternalServerError || googleErr.Code == http.StatusServiceUnavailable {
				//use exponential backoff algorithm
				wait := math.Pow(2, float64(i))
				time.Sleep(time.Duration(wait)*time.Second + time.Duration(time.Now().Nanosecond()%1000)*time.Microsecond)
				continue
			} else {
				//return err
				return productPurchase, err
			}
		} else {
			break
		}
	}

	return productPurchase, err
}
