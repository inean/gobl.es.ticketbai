// Package bizkaia builds the Bizkaia LROE submission for signed TicketBAI invoices.
package bizkaia

import (
	"errors"

	"github.com/invopop/gobl.es.ticketbai/internal/gateways/ebizkaia"
)

// AltaData contains the ledger details bound to a signed TicketBAI invoice.
type AltaData struct {
	Model    string // "140" for an individual or "240" for an organisation.
	NIF      string // Issuer's tax identifier.
	Name     string // Issuer's legal name.
	Year     string // Four-digit invoice issue year.
	Activity string // IAE epígrafe for model 140.
}

// NewAltaRequest encodes one subchapter 1.1 Alta without performing any network
// I/O. It returns the gzipped XML and single-value HTTP headers, including the
// N3 metadata. The caller supplies signed XML and the matching ledger details.
func NewAltaRequest(signedXML []byte, data AltaData) ([]byte, map[string]string, error) {
	if data.Model != ebizkaia.Modelo140 && data.Model != ebizkaia.Modelo240 {
		return nil, nil, errors.New("unsupported Bizkaia LROE model")
	}
	request, err := ebizkaia.NewCreateRequest(&ebizkaia.Supplier{
		Model: data.Model, NIF: data.NIF, Name: data.Name, Year: data.Year, Activity: data.Activity,
	}, signedXML)
	if err != nil {
		return nil, nil, err
	}
	return request.Payload, request.Headers, nil
}
