package slides

import "time"

// Сейчас для простоты опущены коллбэки.
// О них поговорим позже, линейности происходящего они не отменяют.
func processAuthorizationJobHandler(auth AuthParams) {
	// Сходить в PMS за customer id
	customerID, _ := pms.GetCustomerID(auth.CardID) // retries внутри с помощью `for {  if err != nil time.Sleep }`

	// Создать payment в Zooz
	paymentStatus, paymentID, _ := zooz.CreatePayment(customerID, auth)
	// Если pending - подождать финального результата
	for paymentStatus == "pending" {
		time.Sleep(10 * time.Minute)
		paymentStatus, _ = zooz.GetPayment(paymentID)
	}

	// Сходить в PMS за токеном карты (offtopic: почему сразу вместе с customerID не получить)
	cardToken, _ := pms.GetCardToken(auth.CardID)

	// Создать авторизацию в Zooz
	status, authID, _ := zooz.CreateAuthorization(cardToken, paymentID, auth)
	// Если pending - оповестить GPM
	if status == "pending" {
		gpm.Notify("pending")
	}
	// и подождать финального результата
	for status == "pending" {
		time.Sleep(10 * time.Minute)
		status, _ = zooz.GetAuthorization(authID)
	}

	// Отправить результат в GPM
	gpm.Notify(status)
}


























/*

Так может в продакшн?

Ну и что, что пендинг может висеть потенциально несколько дней.
Горутины дешёвые - для этого и создавались (у себя локально на маке я создал миллион, работающих со структрой  Authorize - 2.5GB ).
Для джобов добавить хэлсчеки (если их нет), чтобы мониторить крэши длинных джобов.


Проблема длинных джобов:
 * Увеличивается шанс словить даунскейл в ходе обработки.
   Хоть всё и идемпотентно, не хочется повторять заново уже прошедшие шаги.
   ! Можно после каждого шага персистить в базу а перед шагов делать проверку (ручной чек выполнен шаг или нет)
 * Надо быть осторожным с памятью (например, не держать сырые респонсы от зуза где-нибудь в кэше).
   ! С памятью неплохо бы быть всегда осторожным

Проблемы нескольких инстансов:
 * Коллбек приходить на один инстанс, а горутина работает на другом
   ! Надо уметь размножать коллбэки на все инстансы.
 * Когда приходит capture, может захотеться прочекать состояние авторизации
   ! Должно решится промежуточными персистами в базу.
*/
