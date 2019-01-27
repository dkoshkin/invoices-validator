package notifier

import (
	"github.com/dkoshkin/invoices-validator/pkg/validator"
	"github.com/thoas/go-funk"
	"log"
	"os"
	"strings"
)

const (
	notifierEmailMapEnv   = "NOTIFIER_NAME_EMAIL_PAIRS"

	notifierSMSNumbersEnv = "NOTIFIER_SMS_PHONE_NUMBERS"
)

var defaultEmailContactsGetter = func() []Contact {
	notifierContactsRaw := os.Getenv(notifierEmailMapEnv)
	// split "name1=email1:name2=email2" to a slice of notifier.Contact
	contactsRaw := strings.Split(notifierContactsRaw, ":")
	contacts := make([]Contact, 0)

	funk.ForEach(contactsRaw, func(contact string) {
		split := strings.Split(contact, "=")
		if len(split) != 2 {
			log.Fatalf("invalid name=email pair: %q", contact)
		}
		c := Contact{split[0], split[1]}
		contacts = append(contacts, c)
	})

	return contacts
}

var defaultSMSContactGetter = func() []Contact {
	notifierSMSNumbers := os.Getenv(notifierSMSNumbersEnv)
	contacts := make([]Contact, 0)
	for _, number := range strings.Split(notifierSMSNumbers, ":") {
		contacts = append(contacts, Contact{Address: number})
	}

	return contacts
}

type Notifier interface {
	Initialized() bool
	// SetContactsGetter allows to customize how the contacts are populated when Sending
	SetContactsGetter(f func() []Contact)
	Send(subject string, content string) error
	FormatContent(errs []validator.ValidationError) (string, error)
}

type Contact struct {
	Name    string
	Address string
}

func ConfiguredNotifiers() []Notifier {
	var notifiers []Notifier

	// get all initialized notifiers
	if emailNotifier := NewEmailNotifier(); emailNotifier.Initialized() {
		notifiers = append(notifiers, emailNotifier)
	}
	if smsNotifier := NewSMSNotifier(); smsNotifier.Initialized() {
		notifiers = append(notifiers, smsNotifier)
	}

	return notifiers
}
