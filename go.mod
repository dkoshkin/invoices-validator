module github.com/dkoshkin/invoices-validator

// +heroku goVersion go1.11
// +heroku install ./cmd/... ./pkg/...

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/dropbox/dropbox-sdk-go-unofficial v5.4.0+incompatible
	github.com/gorilla/schema v1.0.2 // indirect
	github.com/sendgrid/rest v2.4.1+incompatible // indirect
	github.com/sendgrid/sendgrid-go v3.4.1+incompatible
	github.com/sfreiberg/gotwilio v0.0.0-20181223013140-ccf5c3cb3e06
	github.com/thoas/go-funk v0.0.0-20181020164546-fbae87fb5b5c
	golang.org/x/net v0.0.0-20190125091013-d26f9f9a57f3 // indirect
	golang.org/x/oauth2 v0.0.0-20190115181402-5dab4167f31c // indirect
	k8s.io/klog v0.1.0
)
