package slides

import (
	"github.com/drscre/pos_pipeline/pipeline"
)

func authorizationPipeline1() {
	// Сходить в PMS за customer id
	pipeline.Do(a(func(auth *Authorization) error { // <-- auth это наши "переменные стэка". Сохраняются атомарно в момент завершения шага без ошибки.

		customerID, _ := pms.GetCustomerID(auth.CardID)
		auth.CustomerID = customerID
		return nil

	}))
}
