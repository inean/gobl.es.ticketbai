package bizkaia_test

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/xml"
	"io"
	"testing"

	"github.com/invopop/gobl.es.ticketbai/bizkaia"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAltaRequest(t *testing.T) {
	signed := []byte("<?xml version=\"1.0\"?>\n<TicketBai>\n  <Signature/>\n</TicketBai>\n")
	for _, tc := range []struct{ model, nif, activity string }{
		{model: "240", nif: "B64847106"},
		{model: "140", nif: "12345678Z", activity: "722300"},
	} {
		t.Run(tc.model, func(t *testing.T) {
			data := bizkaia.AltaData{
				Model: tc.model, NIF: tc.nif, Name: "Test issuer", Year: "2026", Activity: tc.activity,
			}
			body, headers, err := bizkaia.NewAltaRequest(signed, data)
			require.NoError(t, err)
			assert.Equal(t, "gzip", headers["Accept-Encoding"])
			assert.Equal(t, "gzip", headers["Content-Encoding"])
			assert.Equal(t, "application/octet-stream", headers["Content-Type"])
			assert.Equal(t, "application/xml", headers["eus-bizkaia-n3-content-type"])
			assert.Equal(t, "1.0", headers["eus-bizkaia-n3-version"])
			assert.JSONEq(t, `{"con":"LROE","apa":"1.1","inte":{"nif":"`+tc.nif+
				`","nrs":"Test issuer"},"drs":{"mode":"`+tc.model+`","ejer":"2026"}}`,
				headers["eus-bizkaia-n3-data"])

			reader, err := gzip.NewReader(bytes.NewReader(body))
			require.NoError(t, err)
			outer, err := io.ReadAll(reader)
			require.NoError(t, err)
			require.NoError(t, reader.Close())
			var request struct {
				Company    string `xml:"FacturasEmitidas>FacturaEmitida>TicketBai"`
				Individual string `xml:"Ingresos>Ingreso>TicketBai"`
				Activity   string `xml:"Ingresos>Ingreso>Renta>DetalleRenta>Epigrafe"`
			}
			require.NoError(t, xml.Unmarshal(outer, &request))
			encoded := request.Company
			if tc.model == "140" {
				assert.Empty(t, request.Company)
				encoded = request.Individual
			} else {
				assert.Empty(t, request.Individual)
			}
			assert.Equal(t, tc.activity, request.Activity)
			embedded, err := base64.StdEncoding.DecodeString(encoded)
			require.NoError(t, err)
			assert.Equal(t, signed, embedded)
		})
	}
}

func TestNewAltaRequestRejectsUnsupportedModel(t *testing.T) {
	_, _, err := bizkaia.NewAltaRequest([]byte("<TicketBai/>"), bizkaia.AltaData{Model: "999"})
	require.Error(t, err)
}
