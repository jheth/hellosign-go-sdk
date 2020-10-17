package hellosign

import (
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/dnaeon/go-vcr/recorder"
	"github.com/stretchr/testify/assert"
)

func TestCreateEmbeddedSignatureRequestSuccess(t *testing.T) {
	// Start our recorder
	vcr := fixture("fixtures/embedded_signature_request")
	defer vcr.Stop() // Make sure recorder is stopped once done with it

	client := createVcrClient(vcr)

	embReq := createEmbeddedRequest()
	res, err := client.CreateEmbeddedSignatureRequest(embReq)

	assert.NotNil(t, res, "Should return response")
	assert.Nil(t, err, "Should not return error")

	assert.Equal(t, "6d7ad140141a7fe6874fec55931c363e0301c353", res.SignatureRequestID)
	assert.Equal(t, "awesome", res.Subject)
	assert.Equal(t, true, res.TestMode)
	assert.Equal(t, false, res.IsComplete)
	assert.Equal(t, false, res.IsDeclined)
}

func TestCreateEmbeddedSignatureRequestSuccess2(t *testing.T) {
	// Start our recorder
	vcr := fixture("fixtures/embedded_signature_request_more_fields")
	defer vcr.Stop() // Make sure recorder is stopped once done with it

	client := createVcrClient(vcr)

	embReq := createEmbeddedRequest()
	res, err := client.CreateEmbeddedSignatureRequest(embReq)

	assert.NotNil(t, res, "Should return response")
	assert.Nil(t, err, "Should not return error")

	assert.Equal(t, "0afd5e3ac99a19a7e2aa68740faf9bd32441dc11", res.SignatureRequestID)
	assert.Equal(t, "awesome", res.Subject)
	assert.Equal(t, true, res.TestMode)
	assert.Equal(t, false, res.IsComplete)
	assert.Equal(t, false, res.IsDeclined)
}

func TestCreateEmbeddedSignatureRequestMissingSigners(t *testing.T) {
	// Start our recorder
	vcr := fixture("fixtures/embedded_signature_request_missing_signers")
	defer vcr.Stop() // Make sure recorder is stopped once done with it

	client := createVcrClient(vcr)

	embReq := createEmbeddedRequest()
	embReq.Signers = []Signer{}

	res, err := client.CreateEmbeddedSignatureRequest(embReq)

	assert.Nil(t, res, "Should not return response")
	assert.NotNil(t, err, "Should return error")

	assert.Equal(t, err.Error(), "bad_request: Must specify a name for each signer")
}

func TestCreateEmbeddedSignatureRequestFileURL(t *testing.T) {
	// Start our recorder
	vcr := fixture("fixtures/embedded_signature_request_file_url")
	defer vcr.Stop() // Make sure recorder is stopped once done with it

	client := createVcrClient(vcr)

	request := EmbeddedRequest{
		TestMode: true,
		ClientID: os.Getenv("HELLOSIGN_CLIENT_ID"),
		FileURL:  []string{"http://www.pdf995.com/samples/pdf.pdf"},
		Title:    "My First Document",
		Subject:  "Contract",
		Signers: []Signer{
			{
				Email: "jane@example.com",
				Name:  "Jane Doe",
			},
		},
	}

	res, err := client.CreateEmbeddedSignatureRequest(request)
	assert.NotNil(t, res, "Should return response")
	assert.Nil(t, err, "Should not return error")

	assert.Equal(t, "c9af885443fad587aa2a4698086c08c64233df64", res.SignatureRequestID)
	assert.Equal(t, "My First Document", res.Title)
	assert.Equal(t, true, res.TestMode)
	assert.Equal(t, false, res.IsComplete)
	assert.Equal(t, false, res.IsDeclined)
}

func TestGetSignatureRequest(t *testing.T) {
	vcr := fixture("fixtures/get_signature_request")
	defer vcr.Stop() // Make sure recorder is stopped once done with it

	client := createVcrClient(vcr)

	res, err := client.GetSignatureRequest("6d7ad140141a7fe6874fec55931c363e0301c353")

	assert.NotNil(t, res, "Should return response")
	assert.Nil(t, err, "Should not return error")

	assert.Equal(t, "6d7ad140141a7fe6874fec55931c363e0301c353", res.SignatureRequestID)
	assert.Equal(t, "awesome", res.Subject)
	assert.Equal(t, true, res.TestMode)
	assert.Equal(t, false, res.IsComplete)
	assert.Equal(t, false, res.IsDeclined)
}

func TestGetSignatureRequests(t *testing.T) {
	vcr := fixture("fixtures/list_signature_requests")
	defer vcr.Stop() // Make sure recorder is stopped once done with it

	client := createVcrClient(vcr)

	res, err := client.ListSignatureRequests()

	assert.NotNil(t, res, "Should return response")
	assert.Nil(t, err, "Should not return error")

	assert.Equal(t, 1, res.ListInfo.NumPages)
	assert.Equal(t, 1, res.ListInfo.Page)
	assert.Equal(t, 19, res.ListInfo.NumResults)
	assert.Equal(t, 20, res.ListInfo.PageSize)

	assert.Equal(t, 19, len(res.SignatureRequests))
}

func TestGetEmbeddedSignURL(t *testing.T) {
	vcr := fixture("fixtures/get_embedded_sign_url")
	defer vcr.Stop() // Make sure recorder is stopped once done with it

	client := createVcrClient(vcr)

	res, err := client.GetEmbeddedSignURL("deaf86bfb33764d9a215a07cc060122d")

	assert.NotNil(t, res, "Should return response")
	assert.Nil(t, err, "Should not return error")

	assert.Contains(t, res.SignURL, "embeddedSign?signature_id=deaf86bfb33764d9a215a07cc060122d&token=")
	assert.Equal(t, 1505259198, res.ExpiresAt)
}

func TestSaveFile(t *testing.T) {
	vcr := fixture("fixtures/get_pdf")
	defer vcr.Stop() // Make sure recorder is stopped once done with it

	client := createVcrClient(vcr)

	fileInfo, err := client.SaveFile("6d7ad140141a7fe6874fec55931c363e0301c353", "pdf", "/tmp/download.pdf")

	assert.NotNil(t, fileInfo, "Should return response")
	assert.Nil(t, err, "Should not return error")

	assert.Equal(t, int64(98781), fileInfo.Size())
	assert.Equal(t, "download.pdf", fileInfo.Name())
}

func TestGetPDF(t *testing.T) {
	vcr := fixture("fixtures/get_pdf")
	defer vcr.Stop() // Make sure recorder is stopped once done with it

	client := createVcrClient(vcr)

	data, err := client.GetPDF("6d7ad140141a7fe6874fec55931c363e0301c353")

	assert.NotNil(t, data, "Should return response")
	assert.Nil(t, err, "Should not return error")

	assert.Equal(t, 98781, len(data))
}

func TestCancelSignatureRequests(t *testing.T) {
	vcr := fixture("fixtures/cancel_signature_request")
	defer vcr.Stop() // Make sure recorder is stopped once done with it

	client := createVcrClient(vcr)

	res, err := client.CancelSignatureRequest("5c002b65dfefab79795a521bef312c45914cc48d")

	assert.NotNil(t, res, "Should return response")
	assert.Nil(t, err, "Should not return error")

	assert.Equal(t, 200, res.StatusCode)
}

func TestUpdateSignatureRequestSuccess(t *testing.T) {
	vcr := fixture("fixtures/update_signature_request")
	defer vcr.Stop() // Make sure recorder is stopped once done with it

	client := createVcrClient(vcr)

	res, err := client.UpdateSignatureRequest(
		"9040be434b1301e31019b3dad895ed580f8ca890",
		"deaf86bfb33764d9a215a07cc060122d",
		"franky@hellosign.com",
	)

	assert.Nil(t, err, "Should not return error")
	assert.NotNil(t, res, "Should return response")

	assert.Equal(t, "9040be434b1301e31019b3dad895ed580f8ca890", res.SignatureRequestID)
	assert.Equal(t, "franky@hellosign.com", res.Signatures[0].SignerEmailAddress)
}

func TestUpdateSignatureRequestFails(t *testing.T) {
	vcr := fixture("fixtures/update_signature_request_deleted")
	defer vcr.Stop() // Make sure recorder is stopped once done with it

	client := createVcrClient(vcr)

	res, err := client.UpdateSignatureRequest(
		"5c002b65dfefab79795a521bef312c45914cc48d",
		"d82212e10dcf71ad465e033907074423",
		"franky@hellosign.com",
	)

	assert.Nil(t, res, "Should not return response")
	assert.NotNil(t, err, "Should return error")

	assert.Equal(t, "deleted: This resource has been deleted", err.Error())
}

// Private Functions

func fixture(path string) *recorder.Recorder {
	vcr, err := recorder.New(path)
	if err != nil {
		log.Fatal(err)
	}
	return vcr
}

func createVcrClient(transport *recorder.Recorder) Client {
	httpClient := &http.Client{Transport: transport}

	client := Client{
		APIKey:     os.Getenv("HELLOSIGN_API_KEY"),
		HTTPClient: httpClient,
	}
	return client
}

func createEmbeddedRequest() EmbeddedRequest {

	return EmbeddedRequest{
		TestMode: true,
		ClientID: os.Getenv("HELLOSIGN_CLIENT_ID"),
		File: []string{
			"fixtures/offer_letter.pdf",
			"fixtures/offer_letter.pdf",
		},
		Title:   "cool title",
		Subject: "awesome",
		Message: "cool message bro",
		// SigningRedirectURL: "example signing redirect url",
		Signers: []Signer{
			{
				Email: "freddy@hellosign.com",
				Name:  "Freddy Rangel",
			},
			{
				Email: "frederick.rangel@gmail.com",
				Name:  "Frederick Rangel",
			},
		},
		CCEmailAddresses: []string{
			"no@cats.com",
			"no@dogs.com",
		},
		UseTextTags:  false,
		HideTextTags: true,
		Metadata: map[string]string{
			"no":   "cats",
			"more": "dogs",
		},
		FormFieldsPerDocument: [][]DocumentFormField{
			[]DocumentFormField{
				{
					APIId:    "api_id",
					Name:     "display name",
					Type:     "text",
					X:        123,
					Y:        456,
					Width:    678,
					Required: true,
					Signer:   0,
				},
			},
			[]DocumentFormField{
				{
					APIId:    "api_id_2",
					Name:     "display name 2",
					Type:     "text",
					X:        123,
					Y:        456,
					Width:    678,
					Required: true,
					Signer:   1,
				},
			},
		},
	}
}
