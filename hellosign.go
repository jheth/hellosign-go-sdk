package hellosign

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"reflect"
	"strconv"
)

const (
	baseURL string = "https://api.hellosign.com/v3/"
)

// Client contains APIKey and optional http.client
type Client struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

// EmbeddedRequest contains the request parameters for create_embedded
type EmbeddedRequest struct {
	TestMode           bool     `form_field:"test_mode"`
	ClientID           string   `form_field:"client_id"`
	FileURL            []string `form_field:"file_url"`
	File               []string `form_field:"file"`
	Title              string   `form_field:"title"`
	Subject            string   `form_field:"subject"`
	Message            string   `form_field:"message"`
	SigningRedirectURL string   `form_field:"signing_redirect_url"`
	Signers            []Signer `form_field:"signers"`
	// Attachments            []Attachment `form_field:"attachments"`
	CustomFields     []CustomField     `form_field:"custom_fields"`
	CCEmailAddresses []string          `form_field:"cc_email_addresses"`
	UseTextTags      bool              `form_field:"use_text_tags"`
	HideTextTags     bool              `form_field:"hide_text_tags"`
	Metadata         map[string]string `form_field:"metadata"`
	// AllowDecline          int                   `form_field:"allow_decline"`
	// AllowReassign         int                   `form_field:"allow_reassign"`
	FormFieldsPerDocument [][]DocumentFormField `form_field:"form_fields_per_document"`
	// FieldOptions map[string]string `form_field:"field_options"``
}

type Signer struct {
	Name  string `field:"name"`
	Email string `field:"email_address"`
	Order int    `field:"order"`
	Pin   string `field:"pin"`
}

type DocumentFormField struct {
	APIId    string `json:"api_id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	X        int    `json:"x"`
	Y        int    `json:"y"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	Required bool   `json:"required"`
	Signer   int    `json:"signer"`
}

type SignatureRequestResponse struct {
	SignatureRequest *SignatureRequest `json:"signature_request"`
}

type SignatureRequest struct {
	TestMode              bool                     `json:"test_mode"`               // Whether this is a test signature request. Test requests have no legal value. Defaults to 0.
	SignatureRequestID    string                   `json:"signature_request_id"`    // The id of the SignatureRequest.
	RequesterEmailAddress string                   `json:"requester_email_address"` // The email address of the initiator of the SignatureRequest.
	Title                 string                   `json:"title"`                   // The title the specified Account uses for the SignatureRequest.
	OriginalTitle         string                   `json:"original_title"`          // Default Label for account.
	Subject               string                   `json:"subject"`                 // The subject in the email that was initially sent to the signers.
	Message               string                   `json:"message"`                 // The custom message in the email that was initially sent to the signers.
	Metadata              map[string]interface{}   `json:"metadata"`                // The metadata attached to the signature request.
	CreatedAt             int                      `json:"created_at"`              // Time the signature request was created.
	IsComplete            bool                     `json:"is_complete"`             // Whether or not the SignatureRequest has been fully executed by all signers.
	IsDeclined            bool                     `json:"is_declined"`             // Whether or not the SignatureRequest has been declined by a signer.
	HasError              bool                     `json:"has_error"`               // Whether or not an error occurred (either during the creation of the SignatureRequest or during one of the signings).
	FilesURL              string                   `json:"files_url"`               // The URL where a copy of the request's documents can be downloaded.
	SigningURL            string                   `json:"signing_url"`             // The URL where a signer, after authenticating, can sign the documents. This should only be used by users with existing HelloSign accounts as they will be required to log in before signing.
	DetailsURL            string                   `json:"details_url"`             // The URL where the requester and the signers can view the current status of the SignatureRequest.
	CCEmailAddress        []*string                `json:"cc_email_addresses"`      // A list of email addresses that were CCed on the SignatureRequest. They will receive a copy of the final PDF once all the signers have signed.
	SigningRedirectURL    string                   `json:"signing_redirect_url"`    // The URL you want the signer redirected to after they successfully sign.
	CustomFields          []map[string]interface{} `json:"custom_fields"`           // An array of Custom Field objects containing the name and type of each custom field.
	ResponseData          []*ResponseData          `json:"response_data"`           // An array of form field objects containing the name, value, and type of each textbox or checkmark field filled in by the signers.
	Signatures            []*Signature             `json:"signatures"`              // An array of signature objects, 1 for each signer.
	Warnings              []*Warning               `json:"warnings"`                // An array of warning objects.
}

type CustomField struct {
	Name     string      `json:"name"`     // The name of the Custom Field.
	Type     string      `json:"type"`     // The type of this Custom Field. Only 'text' and 'checkbox' are currently supported.
	Value    interface{} `json:"value"`    // A text string for text fields or true/false for checkbox fields
	Required bool        `json:"required"` // A boolean value denoting if this field is required.
	ApiID    string      `json:"api_id"`   // The unique ID for this field.
	Editor   *string     `json:"editor"`   // The name of the Role that is able to edit this field.
}

type ResponseData struct {
	ApiID       string `json:"api_id"`       // The unique ID for this field.
	SignatureID string `json:"signature_id"` // The ID of the signature to which this response is linked.
	Name        string `json:"name"`         // The name of the form field.
	Value       string `json:"value"`        // The value of the form field.
	Required    bool   `json:"required"`     // A boolean value denoting if this field is required.
	Type        string `json:"type"`         // The type of this form field. See field types
}

type Signature struct {
	SignatureID        string  `json:"signature_id"`         // Signature identifier.
	SignerEmailAddress string  `json:"signer_email_address"` // The email address of the signer.
	SignerName         string  `json:"signer_name"`          // The name of the signer.
	Order              int     `json:"order"`                // If signer order is assigned this is the 0-based index for this signer.
	StatusCode         string  `json:"status_code"`          // The current status of the signature. eg: awaiting_signature, signed, declined
	DeclineReason      string  `json:"decline_reason"`       // The reason provided by the signer for declining the request.
	SignedAt           int     `json:"signed_at"`            // Time that the document was signed or null.
	LastViewedAt       int     `json:"last_viewed_at"`       // The time that the document was last viewed by this signer or null.
	LastRemindedAt     int     `json:"last_reminded_at"`     // The time the last reminder email was sent to the signer or null.
	HasPin             bool    `json:"has_pin"`              // Boolean to indicate whether this signature requires a PIN to access.
	ReassignedBy       string  `json:"reassigned_by"`        // Email address of original signer who reassigned to this signer.
	ReassignmentReason string  `json:"reassignment_reason"`  // Reason provided by original signer who reassigned to this signer.
	Error              *string `json:"error"`                // Error message pertaining to this signer, or null.
}

type Warning struct {
	Message string `json:"warning_msg"`
	Name    string `json:"warning_name"`
}

type ListResponse struct {
	ListInfo          *ListInfo           `json:"list_info"`
	SignatureRequests []*SignatureRequest `json:"signature_requests"`
}

type ListInfo struct {
	NumPages   int `json:"num_pages"`   // Total number of pages available
	NumResults int `json:"num_results"` // Total number of objects available
	Page       int `json:"page"`        // Number of the page being returned
	PageSize   int `json:"page_size"`   // Objects returned per page
}

type ErrorResponse struct {
	Error *Error `json:"error"`
}

type Error struct {
	Message string `json:"error_msg"`
	Name    string `json:"error_name"`
}

type EmbeddedResponse struct {
	Embedded *SignURLResponse `json:"embedded"`
}

type SignURLResponse struct {
	SignURL   string `json:"sign_url"`   // URL of the signature page to display in the embedded iFrame.
	ExpiresAt int    `json:"expires_at"` // When the link expires.
}

// CreateEmbeddedSignatureRequest creates a new embedded signature
func (m *Client) CreateEmbeddedSignatureRequest(
	embeddedRequest EmbeddedRequest) (*SignatureRequest, error) {

	params, writer, err := m.marshalMultipartRequest(embeddedRequest)
	if err != nil {
		return nil, err
	}

	response, err := m.post("signature_request/create_embedded", params, *writer)
	if err != nil {
		return nil, err
	}

	return m.sendSignatureRequest(response)
}

// GetSignatureRequest - Gets a SignatureRequest that includes the current status for each signer.
func (m *Client) GetSignatureRequest(signatureRequestID string) (*SignatureRequest, error) {
	path := fmt.Sprintf("signature_request/%s", signatureRequestID)
	response, err := m.get(path)
	if err != nil {
		return nil, err
	}
	return m.sendSignatureRequest(response)
}

// GetEmbeddedSignURL - Retrieves an embedded signing object.
func (m *Client) GetEmbeddedSignURL(signatureRequestID string) (*SignURLResponse, error) {
	path := fmt.Sprintf("embedded/sign_url/%s", signatureRequestID)
	response, err := m.get(path)
	if err != nil {
		return nil, err
	}

	data := &EmbeddedResponse{}
	err = json.NewDecoder(response.Body).Decode(data)
	if err != nil {
		return nil, err
	}

	return data.Embedded, nil
}

func (m *Client) SaveFile(signatureRequestID, fileType, destFilePath string) (os.FileInfo, error) {
	bytes, err := m.GetFiles(signatureRequestID, fileType)

	out, err := os.Create(destFilePath)
	if err != nil {
		return nil, err
	}
	out.Write(bytes)
	out.Close()

	info, err := os.Stat(destFilePath)
	if err != nil {
		return nil, err
	}
	return info, nil
}

// GetPDF - Obtain a copy of the current pdf specified by the signature_request_id parameter.
func (m *Client) GetPDF(signatureRequestID string) ([]byte, error) {
	return m.GetFiles(signatureRequestID, "pdf")
}

// GetFiles - Obtain a copy of the current documents specified by the signature_request_id parameter.
// signatureRequestID - The id of the SignatureRequest to retrieve.
// fileType - Set to "pdf" for a single merged document or "zip" for a collection of individual documents.
func (m *Client) GetFiles(signatureRequestID, fileType string) ([]byte, error) {
	path := fmt.Sprintf("signature_request/files/%s", signatureRequestID)

	var params bytes.Buffer
	writer := multipart.NewWriter(&params)

	signatureIDField, err := writer.CreateFormField("file_type")
	if err != nil {
		return nil, err
	}
	signatureIDField.Write([]byte(fileType))

	emailField, err := writer.CreateFormField("get_url")
	if err != nil {
		return nil, err
	}
	emailField.Write([]byte("false"))

	response, err := m.request("GET", path, &params, *writer)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// ListSignatureRequests - Lists the SignatureRequests (both inbound and outbound) that you have access to.
func (m *Client) ListSignatureRequests() (*ListResponse, error) {
	path := fmt.Sprintf("signature_request/list")
	response, err := m.get(path)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	listResponse := &ListResponse{}
	err = json.NewDecoder(response.Body).Decode(listResponse)
	if err != nil {
		return nil, err
	}

	return listResponse, err
}

// UpdateSignatureRequest - Update an email address on a signature request.
func (m *Client) UpdateSignatureRequest(signatureRequestID string, signatureID string, email string) (*SignatureRequest, error) {
	path := fmt.Sprintf("signature_request/update/%s", signatureRequestID)

	var params bytes.Buffer
	writer := multipart.NewWriter(&params)

	signatureIDField, err := writer.CreateFormField("signature_id")
	if err != nil {
		return nil, err
	}
	signatureIDField.Write([]byte(signatureID))

	emailField, err := writer.CreateFormField("email_address")
	if err != nil {
		return nil, err
	}
	emailField.Write([]byte(email))

	response, err := m.post(path, &params, *writer)
	if err != nil {
		return nil, err
	}

	return m.sendSignatureRequest(response)
}

// CancelSignatureRequest - Cancels an incomplete signature request. This action is not reversible.
func (m *Client) CancelSignatureRequest(signatureRequestID string) (*http.Response, error) {
	path := fmt.Sprintf("signature_request/cancel/%s", signatureRequestID)

	response, err := m.nakedPost(path)
	if err != nil {
		return nil, err
	}

	return response, err
}

// Private Methods

func (m *Client) marshalMultipartRequest(
	embRequest EmbeddedRequest) (*bytes.Buffer, *multipart.Writer, error) {

	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	structType := reflect.TypeOf(embRequest)
	val := reflect.ValueOf(embRequest)

	for i := 0; i < val.NumField(); i++ {

		valueField := val.Field(i)
		f := valueField.Interface()
		val := reflect.ValueOf(f)
		field := structType.Field(i)
		fieldTag := field.Tag.Get("form_field")

		switch val.Kind() {
		case reflect.Map:
			for k, v := range embRequest.Metadata {
				formField, err := w.CreateFormField(fmt.Sprintf("metadata[%v]", k))
				if err != nil {
					return nil, nil, err
				}
				formField.Write([]byte(v))
			}
		case reflect.Slice:
			switch fieldTag {
			case "signers":
				for i, signer := range embRequest.Signers {
					email, err := w.CreateFormField(fmt.Sprintf("signers[%v][email_address]", i))
					if err != nil {
						return nil, nil, err
					}
					email.Write([]byte(signer.Email))

					name, err := w.CreateFormField(fmt.Sprintf("signers[%v][name]", i))
					if err != nil {
						return nil, nil, err
					}
					name.Write([]byte(signer.Name))

					if signer.Order != 0 {
						order, err := w.CreateFormField(fmt.Sprintf("signers[%v][order]", i))
						if err != nil {
							return nil, nil, err
						}
						order.Write([]byte(strconv.Itoa(signer.Order)))
					}

					if signer.Pin != "" {
						pin, err := w.CreateFormField(fmt.Sprintf("signers[%v][pin]", i))
						if err != nil {
							return nil, nil, err
						}
						pin.Write([]byte(signer.Pin))
					}
				}
			case "cc_email_addresses":
				for k, v := range embRequest.CCEmailAddresses {
					formField, err := w.CreateFormField(fmt.Sprintf("cc_email_addresses[%v]", k))
					if err != nil {
						return nil, nil, err
					}
					formField.Write([]byte(v))
				}
			case "form_fields_per_document":
				if len(embRequest.FormFieldsPerDocument) > 0 {
					formField, err := w.CreateFormField(fieldTag)
					if err != nil {
						return nil, nil, err
					}
					ffpdJSON, err := json.Marshal(embRequest.FormFieldsPerDocument)
					if err != nil {
						return nil, nil, err
					}
					formField.Write([]byte(ffpdJSON))
				}
			case "file":
				for i, path := range embRequest.File {
					file, _ := os.Open(path)

					formField, err := w.CreateFormFile(fmt.Sprintf("file[%v]", i), file.Name())
					if err != nil {
						return nil, nil, err
					}
					_, err = io.Copy(formField, file)
				}
			case "file_url":
				for i, fileURL := range embRequest.FileURL {
					formField, err := w.CreateFormField(fmt.Sprintf("file_url[%v]", i))
					if err != nil {
						return nil, nil, err
					}
					formField.Write([]byte(fileURL))
				}
			}
		case reflect.Bool:
			formField, err := w.CreateFormField(fieldTag)
			if err != nil {
				return nil, nil, err
			}
			formField.Write([]byte(m.boolToIntString(val.Bool())))
		default:
			if val.String() != "" {
				formField, err := w.CreateFormField(fieldTag)
				if err != nil {
					return nil, nil, err
				}
				formField.Write([]byte(val.String()))
			}
		}
	}

	w.Close()
	return &b, w, nil
}

func (m *Client) get(path string) (*http.Response, error) {
	endpoint := fmt.Sprintf("%s%s", m.getEndpoint(), path)

	var b bytes.Buffer
	request, _ := http.NewRequest("GET", endpoint, &b)
	request.SetBasicAuth(m.APIKey, "")

	response, err := m.getHTTPClient().Do(request)
	if err != nil {
		return nil, err
	}

	return response, err
}

func (m *Client) post(path string, params *bytes.Buffer, w multipart.Writer) (*http.Response, error) {
	return m.request("POST", path, params, w)
}

func (m *Client) request(method string, path string, params *bytes.Buffer, w multipart.Writer) (*http.Response, error) {
	endpoint := fmt.Sprintf("%s%s", m.getEndpoint(), path)
	request, _ := http.NewRequest(method, endpoint, params)
	request.Header.Add("Content-Type", w.FormDataContentType())
	request.SetBasicAuth(m.APIKey, "")

	response, err := m.getHTTPClient().Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode >= 400 {
		e := &ErrorResponse{}
		json.NewDecoder(response.Body).Decode(e)
		msg := fmt.Sprintf("%s: %s", e.Error.Name, e.Error.Message)
		return response, errors.New(msg)
	}

	return response, err
}

func (m *Client) nakedPost(path string) (*http.Response, error) {
	endpoint := fmt.Sprintf("%s%s", m.getEndpoint(), path)
	var b bytes.Buffer
	request, _ := http.NewRequest("POST", endpoint, &b)
	request.SetBasicAuth(m.APIKey, "")

	response, err := m.getHTTPClient().Do(request)
	if err != nil {
		return nil, err
	}
	return response, err
}

func (m *Client) sendSignatureRequest(response *http.Response) (*SignatureRequest, error) {
	defer response.Body.Close()

	sigRequestResponse := &SignatureRequestResponse{}

	err := json.NewDecoder(response.Body).Decode(sigRequestResponse)

	sigRequest := sigRequestResponse.SignatureRequest

	return sigRequest, err
}

func (m *Client) getEndpoint() string {
	var url string
	if m.BaseURL != "" {
		url = m.BaseURL
	} else {
		url = baseURL
	}
	return url
}

func (m *Client) getHTTPClient() *http.Client {
	var httpClient *http.Client
	if m.HTTPClient != nil {
		httpClient = m.HTTPClient
	} else {
		httpClient = &http.Client{}
	}
	return httpClient
}

func (m *Client) boolToIntString(value bool) string {
	if value == true {
		return "1"
	}
	return "0"
}
