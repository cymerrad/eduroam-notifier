package automaton

import "eduroam-notifier/app/models"

type A struct {
}

func New(settings models.NotifierSettings, rules []models.NotifierRule, templates []models.NotifierTemplate) (*A, error) {
	a := &A{}

	return a, nil
}
