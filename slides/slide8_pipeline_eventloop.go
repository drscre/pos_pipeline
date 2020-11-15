package slides

import (
	"time"

	"github.com/drscre/pos_pipeline/pipeline"
)

func authorizationPipeline4() {
	// Если pending - подождать финального результата.

	//for paymentStatus == "pending" {
	//	select {
	//	case paymentStatus = <-callback:
	//	case <-time.After(1*time.Hour): // polling
	//		paymentStatus, _ = zooz.GetPayment(paymentID)
	//	}
	//}

	isPending := i(func(auth Authorization) bool { return auth.PaymentStatus == "pending" })
	pipeline.EventLoop(isPending, // for paymentStatus == "pending"
		pipeline.OnEvent(ev(func(auth *Authorization, event PaymentStatusEvent) error { // case paymentStatus = <-callback:

			auth.PaymentStatus = event.Status
			return nil

		})),
		pipeline.OnTimerTick(10*time.Minute, a(func(auth *Authorization) error { // case <-time.After(1*time.Hour):

			status, _ := zooz.GetPayment(auth.PaymentID)
			auth.PaymentStatus = status
			return nil

		})),
	)
}

// События реализованы в виде постановки джобы на выполнение пайплайна с параметром (где событие в параметре)
// Когда джоба запускается то смотрим - ждёт ли на текущем шаге пайплайн этот ивент (в будущем можно делать что-то вроде pipeline.Listen)
// TimerTick - это по сути тоже событие, то есть джоба, которая просто дёрнет пайплайн через указанный интервал.
