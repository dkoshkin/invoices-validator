# Invoice Validator

## Using It

Set secrets:

```
export DROPBOX_TOKEN=''
export SENDGRID_API_KEY=''
export TWILIO_ACCOUNT_SID=''
export TWILIO_API_KEY=''
```

Modify `hack/env/vars.env` and run:

```
source hack/env/vars.env
./bin/invoices-validator-darwin-amd64 -log-level=debug
```

## Development

```
make test
# build a docker container
make build-container
# or build a binary to bin/
make build-binary
```