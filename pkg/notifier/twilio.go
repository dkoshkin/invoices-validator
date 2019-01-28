package notifier

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/dkoshkin/invoices-validator/pkg/validator"
	"github.com/sfreiberg/gotwilio"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
)

const (
	twilioAPIKeyEnv             = "TWILIO_API_KEY"
	twilioAccountSIDEnv         = "TWILIO_ACCOUNT_SID"
	twillioSenderPhoneNumberEnv = "NOTIFIER_SENDER_PHONE_NUMBER"
)

type twilioNotifier struct {
	client       *gotwilio.Twilio
	senderNumber string

	contacts func() []Contact
}

func NewSMSNotifier() (Notifier, error) {
	log.Info("Initializing SMS notifier...")
	notifier := &twilioNotifier{}

	twilioAPIKey := os.Getenv(twilioAPIKeyEnv)
	twilioAccountSID := os.Getenv(twilioAccountSIDEnv)
	twillioSenderPhoneNumber := os.Getenv(twillioSenderPhoneNumberEnv)

	for env, val := range map[string]string{
		twilioAPIKeyEnv:             twilioAPIKey,
		twilioAccountSIDEnv:         twilioAccountSID,
		twillioSenderPhoneNumberEnv: twillioSenderPhoneNumber,
	} {
		if val == "" {
			return nil, fmt.Errorf("%s variable must be set", env)
		}
	}

	notifier.client = gotwilio.NewTwilioClient(twilioAccountSID, twilioAPIKey)
	notifier.senderNumber = twillioSenderPhoneNumber
	notifier.contacts = defaultSMSContactGetter
	log.Info("SMS notifier initialized successfully")

	return notifier, nil
}

func (n *twilioNotifier) SetContactsGetter(f func() []Contact) {
	n.contacts = f
}

func (n twilioNotifier) Send(subject string, content string) error {
	log.Info("Notifying using SMS notifier...")
	contacts := n.contacts()
	if len(contacts) == 0 {
		return fmt.Errorf("empty phone number list to send to")
	}

	for _, contact := range contacts {
		resp, exception, err := n.client.SendSMS(n.senderNumber, contact.Address, content, "", "")
		if resp != nil {
			log.Debugf("SMS Response:\n%s", spew.Sdump(resp))
		}
		if exception != nil {
			log.Errorf("SMS Exception:\n%s", spew.Sdump(exception))
			return fmt.Errorf("exception sending SMS")
		}
		if err != nil {
			return fmt.Errorf("error sending SMS: %v", err)
		}
		log.Infof("SMS sent successfully to: %q", contact.Address)
	}

	return nil
}

func (n twilioNotifier) FormatContent(errs []validator.ValidationError) (string, error) {
	var content string
	content = "Below is the list of failed validators:\n\n"
	for _, e := range errs {
		content = fmt.Sprintf("%s%s\nActual: %s\nExpected: %s", content, e.AdditionalInfo, e.Actual, e.Expected)
		content = fmt.Sprintf("%s\n%s\n", content, strings.Repeat("-", 30))
	}
	return content, nil
}
