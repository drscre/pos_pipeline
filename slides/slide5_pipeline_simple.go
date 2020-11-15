package slides

import (
	"github.com/drscre/pos_pipeline/pipeline"
)

func declarePipeline() []pipeline.IStep {
	return []pipeline.IStep{

		// Сходить в PMS за customer id
		pipeline.Step(a(func(auth *Authorization) error { // <-- auth это наши "переменные стэка". Сохраняются атомарно в момент завершения шага без ошибки.

			customerID, _ := pms.GetCustomerID(auth.CardID)
			auth.CustomerID = customerID
			return nil

		})).
			Name("pms_customer_id"), // У каждого шага есть уникальное имя, чтобы можно было сохранить в БД какой последний шаг был выполнен
								     // (можно смотреть как на синтаксический сахар для "стейт-машины", где все шаги друг за другом)
	}

}

// В базе данных хранится контекcт выполнения пайплайна.
type State struct {
	ID                string      // Уникальный идентификатор пайплайна (пусть будет id авторизации). По нему берётся лок при выполнении шага.
	Data              interface{} // Данные ("переменные стэка горутины")
	LastCompletedStep string      // IP instruction pointer - последний выполненный шаг
}
