package notifier

import (
	"fmt"
	"github.com/dkoshkin/invoices-validator/pkg/stringsx"
	"github.com/dkoshkin/invoices-validator/pkg/validator"
	"github.com/thoas/go-funk"
	log "k8s.io/klog"
	"os"
)

const (
	enabledNotifiersEnv = "ENABLED_NOTIFIERS"

	notifierEmailMapEnv   = "NOTIFIER_NAME_EMAIL_PAIRS"

	notifierSMSNumbersEnv = "NOTIFIER_SMS_PHONE_NUMBERS"

	emailNotifier = "email"
	smsNotifier = "sms"
)

var defaultEmailContactsGetter = func() []Contact {
	notifierContactsRaw := os.Getenv(notifierEmailMapEnv)
	// split "name1=email1:name2=email2" to a slice of notifier.Contact
	contactsRaw := stringsx.Split(notifierContactsRaw, ":")
	contacts := make([]Contact, 0)

	funk.ForEach(contactsRaw, func(contact string) {
		split := stringsx.Split(contact, "=")
		if len(split) != 2 {
			log.Errorf("invalid name=email pair: %q", contact)
			return
		}
		c := Contact{split[0], split[1]}
		contacts = append(contacts, c)
	})

	return contacts
}

var defaultSMSContactGetter = func() []Contact {
	notifierSMSNumbers := os.Getenv(notifierSMSNumbersEnv)
	contacts := make([]Contact, 0)
	for _, number := range stringsx.Split(notifierSMSNumbers, ":") {
		contacts = append(contacts, Contact{Address: number})
	}

	return contacts
}

type Notifier interface {
	// SetContactsGetter allows to customize how the contacts are populated when Sending
	SetContactsGetter(f func() []Contact)
	Send(subject string, content string) error
	FormatContent(errs []validator.ValidationError) (string, error)
}

type Contact struct {
	Name    string
	Address string
}

func ConfiguredNotifiers() ([]Notifier, error) {
	var notifiers []Notifier

	enabledNotifiersString := os.Getenv(enabledNotifiersEnv)
	if len(enabledNotifiersString) == 0 {
		return notifiers, nil
	}

	enabledNotifiers := stringsx.Split(enabledNotifiersString, ":")

	// get all initialized notifiers
	for _, n := range enabledNotifiers{
		switch n {
		case emailNotifier:
			if emailNotifier, err := NewEmailNotifier(); err != nil {
				log.Errorf("could not initialize email notifier: %v", err)
			} else {
				notifiers = append(notifiers, emailNotifier)
			}
		case smsNotifier:
			if smsNotifier, err := NewSMSNotifier(); err != nil {
				log.Errorf("could not initialize SMS notifier: %v", err)
			} else {
				notifiers = append(notifiers, smsNotifier)
			}
		}
	}

	if len(notifiers) == 0  {
		return notifiers, fmt.Errorf("no notifiers were enabled")
	}

	return notifiers, nil
}
