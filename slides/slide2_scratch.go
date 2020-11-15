package slides



// обработать запрос от GPM
func handleRequest(params AuthParams) {
	jobs.Enqueue("process_authorization", params)
}

func processAuthorizationJobHandle(auth AuthParams) {
	// Сходить в PMS за customer id

	// Создать payment в Zooz
	// Если pending - подождать финального результата

	// Сходить в PMS за токеном карты

	// Создать авторизацию в Zooz
	// Если pending - оповестить GPM и подождать финального результата

	// Отправить результат в GPM
}
