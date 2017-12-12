package Responder

import ()

type Responder struct {
	Result func()
	Status func()
}

func New(result func(), status func()) (*Responder, error) {
	var res Responder
	res.Result = result
	res.Status = status

	return &res, nil
}
