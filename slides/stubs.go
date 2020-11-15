package slides

type AuthParams struct {
	CardID string
}

type Authorization struct {
	ID string
	AuthParams

	// Calculated while handling
	CustomerID string

	PaymentID     string
	PaymentStatus string
}

type PaymentStatusEvent struct {
	Status string
}

var pms struct {
	GetCustomerID func(cardID string) (customerID string, err error)
	GetCardToken  func(cardID string) (cardToken string, err error)
}

var zooz struct {
	CreatePayment       func(customerID string, auth AuthParams) (status, id string, err error)
	GetPayment          func(id string) (status string, err error)
	CreateAuthorization func(cardToken, paymentID string, auth AuthParams) (status, id string, err error)
	GetAuthorization    func(id string) (status string, err error)
}

var gpm struct {
	Notify func(status string)
}

func a(h func(auth *Authorization) error) func(data interface{}) (modifiedData interface{}, err error) {
	return func(data interface{}) (modifiedData interface{}, err error) {
		err = h(data.(*Authorization))
		return data, err
	}
}

func er(h func(auth *Authorization, pipelineErr error) error) func(data interface{}, pipelineErr error) (modifiedData interface{}, err error) {
	return func(data interface{}, pipelineErr error) (modifiedData interface{}, err error) {
		err = h(data.(*Authorization), pipelineErr)
		return data, err
	}
}

func i(f func(auth Authorization) bool) func(data interface{}) bool {
	return func(data interface{}) bool {
		return f(data.(Authorization))
	}
}

func ev(h func(auth *Authorization, event PaymentStatusEvent) error) func(data interface{}, event interface{}) (modifiedData interface{}, err error) {
	return func(data interface{}, event interface{}) (modifiedData interface{}, err error) {
		err = h(data.(*Authorization), event.(PaymentStatusEvent))
		return data, err
	}
}
