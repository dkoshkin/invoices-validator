package notifier

import (
	"bytes"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/dkoshkin/invoices-validator/pkg/validator"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	log "github.com/sirupsen/logrus"
	"html/template"
	"os"
)

const (
	sendgridAPIKeyEnv      = "SENDGRID_API_KEY"
	notifierSenderNameEnv  = "NOTIFIER_SENDER_NAME"
	notifierSenderEmailEnv = "NOTIFIER_SENDER_EMAIL"
)

type sendGridNotifier struct {
	client *sendgrid.Client

	senderName  string
	senderEmail string

	contacts func() []Contact
}

func NewEmailNotifier() (Notifier, error) {
	log.Info("Initializing Email notifier...")
	notifier := &sendGridNotifier{}

	sendgridAPIKey := os.Getenv(sendgridAPIKeyEnv)
	notifierSenderName := os.Getenv(notifierSenderNameEnv)
	notifierSenderEmail := os.Getenv(notifierSenderEmailEnv)

	for env, val := range map[string]string{
		sendgridAPIKeyEnv:      sendgridAPIKey,
		notifierSenderNameEnv:  notifierSenderName,
		notifierSenderEmailEnv: notifierSenderEmail,
	} {
		if val == "" {
			return nil, fmt.Errorf("%s variable must be set", env)
		}
	}

	notifier.client = sendgrid.NewSendClient(sendgridAPIKey)
	notifier.senderName = notifierSenderName
	notifier.senderEmail = notifierSenderEmail
	notifier.contacts = defaultEmailContactsGetter
	log.Info("Email notifier initialized successfully")

	return notifier, nil
}

func (n *sendGridNotifier) SetContactsGetter(f func() []Contact) {
	n.contacts = f
}

func (n sendGridNotifier) Send(subject string, content string) error {
	log.Info("Notifying using Email notifier...")
	from := mail.NewEmail(n.senderName, n.senderEmail)

	contacts := n.contacts()
	if len(contacts) == 0 {
		return fmt.Errorf("empty email list to send to")
	}

	toEmails := make([]*mail.Email, 0)
	for _, contact := range contacts {
		toEmails = append(toEmails, mail.NewEmail(contact.Name, contact.Address))
	}

	personalization := mail.NewPersonalization()
	personalization.AddTos(toEmails...)

	message := &mail.SGMailV3{
		From:             from,
		Subject:          subject,
		Content:          []*mail.Content{mail.NewContent("text/html", content)},
		Personalizations: []*mail.Personalization{personalization},
	}

	log.Debugf("Email Request:\n%s", spew.Sdump(message))

	// send email
	resp, err := n.client.Send(message)
	if resp != nil {
		log.Debugf("Email Response:\n%s", spew.Sdump(resp))
	}
	if err != nil {
		return fmt.Errorf("error sending email: %v", err)
	}

	log.Infof("Email sent successfully to: %v", contacts)

	return nil
}

func (n sendGridNotifier) FormatContent(errs []validator.ValidationError) (string, error) {
	t, err := template.New("email-template").Parse(emailTemplate)
	if err != nil {
		return "", fmt.Errorf("could not create tempalate: %v", err)
	}

	data := struct {
		Items []validator.ValidationError
	}{
		Items: errs,
	}

	var buf bytes.Buffer

	err = t.Execute(&buf, data)
	if err != nil {
		return "", fmt.Errorf("could not execute tempalate: %v", err)
	}

	return buf.String(), nil
}

var emailTemplate = `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd">
<html data-editor-version="2" class="sg-campaigns" xmlns="http://www.w3.org/1999/xhtml">

<head>
  <meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1, minimum-scale=1, maximum-scale=1" />
  <!--[if !mso]><!-->
  <meta http-equiv="X-UA-Compatible" content="IE=Edge" />
  <!--<![endif]-->
  <!--[if (gte mso 9)|(IE)]>
    <xml>
    <o:OfficeDocumentSettings>
    <o:AllowPNG/>
    <o:PixelsPerInch>96</o:PixelsPerInch>
    </o:OfficeDocumentSettings>
    </xml>
    <![endif]-->
  <!--[if (gte mso 9)|(IE)]>
    <style type="text/css">
      body {width: 600px;margin: 0 auto;}
      table {border-collapse: collapse;}
      table, td {mso-table-lspace: 0pt;mso-table-rspace: 0pt;}
      img {-ms-interpolation-mode: bicubic;}
    </style>
    <![endif]-->

  <style type="text/css">
    body,
    p,
    div {
      font-family: arial;
      font-size: 14px;
    }

    body {
      color: #000000;
    }

    body a {
      color: #1188E6;
      text-decoration: none;
    }

    p {
      margin: 0;
      padding: 0;
    }

    table.wrapper {
      width: 100% !important;
      table-layout: fixed;
      -webkit-font-smoothing: antialiased;
      -webkit-text-size-adjust: 100%;
      -moz-text-size-adjust: 100%;
      -ms-text-size-adjust: 100%;
    }

    img.max-width {
      max-width: 100% !important;
    }

    .column.of-2 {
      width: 50%;
    }

    .column.of-3 {
      width: 33.333%;
    }

    .column.of-4 {
      width: 25%;
    }

    @media screen and (max-width:480px) {

      .preheader .rightColumnContent,
      .footer .rightColumnContent {
        text-align: left !important;
      }

      .preheader .rightColumnContent div,
      .preheader .rightColumnContent span,
      .footer .rightColumnContent div,
      .footer .rightColumnContent span {
        text-align: left !important;
      }

      .preheader .rightColumnContent,
      .preheader .leftColumnContent {
        font-size: 80% !important;
        padding: 5px 0;
      }

      table.wrapper-mobile {
        width: 100% !important;
        table-layout: fixed;
      }

      img.max-width {
        height: auto !important;
        max-width: 480px !important;
      }

      a.bulletproof-button {
        display: block !important;
        width: auto !important;
        font-size: 80%;
        padding-left: 0 !important;
        padding-right: 0 !important;
      }

      .columns {
        width: 100% !important;
      }

      .column {
        display: block !important;
        width: 100% !important;
        padding-left: 0 !important;
        padding-right: 0 !important;
        margin-left: 0 !important;
        margin-right: 0 !important;
      }
    }
  </style>
  <!--user entered Head Start-->

  <!--End Head user entered-->
</head>

<body>
  <center class="wrapper" data-link-color="#1188E6" data-body-style="font-size: 14px; font-family: arial; color: #000000; background-color: #ffffff;">
    <div class="webkit">
      <table cellpadding="0" cellspacing="0" border="0" width="100%" class="wrapper" bgcolor="#ffffff">
        <tr>
          <td valign="top" bgcolor="#ffffff" width="100%">
            <table width="100%" role="content-container" class="outer" align="left" cellpadding="0" cellspacing="0"
              border="0">
              <tr>
                <td width="100%">
                  <table width="100%" cellpadding="0" cellspacing="0" border="0">
                    <tr>
                      <td>
                        <!--[if mso]>
                          <center>
                          <table><tr><td width="600">
                          <![endif]-->
                        <table width="100%" cellpadding="0" cellspacing="0" border="0" style="width: 100%; max-width:600px;"
                          align="left">
                          <tr>
                            <td role="modules-container" style="padding: 0px 0px 0px 0px; color: #000000; text-align: left;"
                              bgcolor="#ffffff" width="100%" align="left">

                              <table class="module preheader preheader-hide" role="module" data-type="preheader" border="0"
                                cellpadding="0" cellspacing="0" width="100%" style="display: none !important; mso-hide: all; visibility: hidden; opacity: 0; color: transparent; height: 0; width: 0;">
                                <tr>
                                  <td role="module-content">
                                    <p></p>
                                  </td>
                                </tr>
                              </table>

                              <table class="module" role="module" data-type="text" border="0" cellpadding="0"
                                cellspacing="0" width="100%" style="table-layout: fixed;">
                                <tr>
                                  <td style="padding:18px 0px 18px 0px;line-height:22px;text-align:inherit;" height="100%"
                                    valign="top" bgcolor="">
                                    <div>Below is the list of failed validators:</div>
									<div>
										<ul>
										{{range .Items}}
											<li>
												{{ .AdditionalInfo }}<br />
												<span style="padding-left: 20px">Actual: {{ .Actual }}</span><br />
												<span style="padding-left: 20px">Expected: {{ .Expected }}</span>
											</li>
										{{end}}
										<ul>
									</div>
                                  </td>
                                </tr>
                              </table>

                            </td>
                          </tr>
                        </table>
                        <!--[if mso]>
                          </td></tr></table>
                          </center>
                          <![endif]-->
                      </td>
                    </tr>
                  </table>
                </td>
              </tr>
            </table>
          </td>
        </tr>
      </table>
    </div>
  </center>
</body>

</html>`
