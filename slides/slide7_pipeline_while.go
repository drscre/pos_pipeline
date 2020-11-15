package slides

import (
	"time"

	"github.com/drscre/pos_pipeline/pipeline"
)

func authorizationPipeline3() {
	// Создать payment в Zooz
	pipeline.Step(a(func(auth *Authorization) error {

		status, id, _ := zooz.CreatePayment(auth.CustomerID, auth.AuthParams)
		auth.PaymentID = id
		auth.PaymentStatus = status
		return nil

	}))

	// Если pending - подождать финального результата.
	// Будем проверять результат раз в 10 минут

	// 	for paymentStatus == "pending" {
	//		time.Sleep(10*time.Minute)
	//		paymentStatus, _ = zooz.GetPayment(paymentID)
	//	}
	isPending := i(func(auth Authorization) bool { return auth.PaymentStatus == "pending" })
	pipeline.While(isPending,
		pipeline.Sleep(10*time.Minute), // <-- здесь пайплайн "запаркуется": джоба либо завершится и запланирует новую на будущее, либо пойдёт на ретрай через механизм ретраев
		pipeline.Step(a(func(auth *Authorization) error {

			status, _ := zooz.GetPayment(auth.PaymentID)
			auth.PaymentStatus = status
			return nil

		})),
	)
}
