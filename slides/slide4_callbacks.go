package slides

import "time"

func processAuthorization4(auth AuthParams, callback chan string) {
	// Создать payment в Zooz
	paymentStatus, paymentID, _ := zooz.CreatePayment("customerID", auth)
	// Если pending - подождать финального результата
	for paymentStatus == "pending" {
		select {
		case paymentStatus = <-callback: // в коллбэках только финальный статус для простоты
		case <-time.After(1 * time.Hour): // polling
			paymentStatus, _ = zooz.GetPayment(paymentID)
		}
	}

	// было
	//for paymentStatus == "pending" {
	//	time.Sleep(1*time.Hour)
	//	paymentStatus, _ = zooz.GetPayment(paymentID)
	//}

}
