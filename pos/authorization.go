package pos

import (
	"time"

	"github.com/drscre/pos_pipeline/pipeline"
)

// Получить запрос от GPM
//
// Сходить в PMS за customer id
//
// Создать payment в Zooz (может вернуть пендинг)
// Подождать конечного ответа
//
// Сходить в PMS за токеном карты
//
// Создать авторизацию в Zooz
// Если пендинг - оповестить GPM
// Подождать окончательный результат
//}
// Вернуть результат в GPM


// Step: do something.
// On successful cb call *checkpoint* and continue to next step.
// On error - *no checkpoint*.
// If error is retryable - retry according to step retry policy. (error is not part of logic flow, that's why no checkpoint).
// If error is not retryable or maximum number of retries reached - abort pipeline with error.
//
// Can have a delayed launch.

// Loop: repeatedly do something until cb returns stop = true.
// *checkpoint* after every successful cb execution.
// If error - *no checkpoint*, retry according to retry policy.
// Successful execution resets retry policy failure count.
// Can have a delayed launch.


func authorizationPipeline() []pipeline.Step {
	// Retrieve customer id from PMS
	pipeline.Do(a(func(auth *Authorization) error {

		customerID, err := pmsAPI.GetCustomerID(auth.CardID)
		if err != nil {
			return pipeline.Retry(err)
		}
		auth.CustomerID = customerID
		return nil

	}))

	// Create authorization
	pipeline.Do(a(func(auth *Authorization) error {

		status, err := zoozAPI.CreateAuthorization(auth)
		if err != nil {
			return pipeline.Retry(err)
		}
		auth.Status = status
		return nil

	}))
	/*
	// Authorization can be pending. Lets poll for status until "success" or "fail"
	isPending := i(func (auth Authorization) bool {  return auth.Status == "pending"  })
	pipeline.While(isPending,
		pipeline.Sleep(1 * time.Minute),
		pipeline.Do(a(func(auth *Authorization) error {
			status, _ := zoozAPI.GetAuthorization(auth.ID)
			auth.Status = status
			return nil
		})),
	)
	 */


	/*
	// Продолжая аналогию с джоба/пайплайн = горутина
	pipeline.Go() // "форкаем" текущий стейт и пуляем соседнюю джобу про него. Что делать с ID? Теперь надо лочить данные??
	Авторизация становится "шареным" ресурсом.
	В горутине - мьютексы. В джобах - локи (или select for update)
	Автоматически шарить и лочить стейт (на уровне либы) - в отличие от горутины не нужы мьютексы
	*/


	// Step
	{
		auth.status := zoozAPI.createAuthorization()
		return auth
	}
	If(status == "pending",
		WaitLoop(func(event) {
			status, err := zoozAPI.getAuthorization()
			if err != nil {
				// log and go for next iteration
			}
		})
		)

	for {
		status := zoozAPI.getAuthorization()
		if status == "success/fail" {
			break
		}
		time.Sleep(1) // <-- stop point
	}
}









func receiveGPMRequest(authParams AuthParams) {
	validate(authParams)
	auth := dbInsert(authParams)

	authorizationPipeline().For(auth.ID)
}

type Authorization struct {
	ID           string // serves also as pipeline id
	CustomerID   string
	Currency string
	Amount int64
	CardID string
	Status string


	PipelineStep string // store pipeline state in the authorization entity itself
}

type AuthParams struct {
	ID       string
	Currency string
	Amount   int64
	Account  string
}


//func a(h func(auth Authorization) (modified Authorization, err error)) func(data interface{}) (modifiedData interface{}, err error) {
//	return func(data interface{}) (modifiedData interface{}, err error) {
//		return h(data.(Authorization))
//	}
//}

func a(h func(auth *Authorization) error) func(data interface{}) (modifiedData interface{}, err error) {
	return func(data interface{}) (modifiedData interface{}, err error) {
		err = h(data.(*Authorization))
		return data, err
	}
}

func i(f func(auth Authorization) bool) func(data interface{}) bool {
	return func(data interface{}) bool {
		return f(data.(Authorization))
	}
}
