package slides

import (
	"github.com/drscre/pos_pipeline/pipeline"
)

func authorizationPipeline2() {
	// Сходить в PMS за customer id
	pipeline.Do(a(func(auth *Authorization) error {

		customerID, err := pms.GetCustomerID(auth.CardID)
		if err != nil {
			return pipeline.Retry(err)
		}
		auth.CustomerID = customerID
		return nil

	})).
		RetryPolicy( /* ... */ ) // <-- Можно задавать политику ретраев для конкретного шага

	// Обработчик фатальных ошибок.
	// Если шаг пайплайна возвращает неретраемую ошибку (нет кредитной карточки, исчерпано максимальное количество попыток, ...)
	// то выполнение прерывается и вызывается обработчки ошибок.
	// После чего пайплайн завершается.
	pipeline.ErrorHandler(er(func(auth *Authorization, err error) error {
		gpm.Notify("error")
		return nil
	}))
}
