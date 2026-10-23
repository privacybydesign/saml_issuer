// Package irma starts sessions at an IRMA (Yivi) server.
package irma

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	issuanceContext = "https://irma.app/ld/request/issuance/v2"
	sessionPath     = "/session"
	requestTimeout  = 15 * time.Second
	maxResponse     = 1 << 20

	headerContentType   = "Content-Type"
	headerAuthorization = "Authorization"
	contentTypeJSON     = "application/json"
)

// Client talks to an IRMA server with the requestor API token.
type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

// NewClient returns a client for the IRMA server at baseURL. The token is
// sent as the Authorization header; leave it empty for a server without one.
func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		http:    &http.Client{Timeout: requestTimeout},
	}
}

// Issuance asks for one credential to be issued.
type Issuance struct {
	Credential string
	Attributes map[string]string
	Validity   time.Time
	// SDJWTBatchSize is the number of SD-JWT instances to issue.
	SDJWTBatchSize int
}

type issuanceRequest struct {
	Context     string              `json:"@context"`
	Credentials []credentialRequest `json:"credentials"`
}

type credentialRequest struct {
	Credential     string            `json:"credential"`
	Validity       int64             `json:"validity"`
	Attributes     map[string]string `json:"attributes"`
	SDJWTBatchSize int               `json:"sdJwtBatchSize"`
}

// StartIssuance starts an issuance session and returns the IRMA server's
// response, which holds the session pointer for the Yivi frontend.
func (c *Client) StartIssuance(ctx context.Context, issuance Issuance) ([]byte, error) {
	body, err := json.Marshal(issuanceRequest{
		Context: issuanceContext,
		Credentials: []credentialRequest{{
			Credential:     issuance.Credential,
			Validity:       issuance.Validity.Unix(),
			Attributes:     issuance.Attributes,
			SDJWTBatchSize: issuance.SDJWTBatchSize,
		}},
	})
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+sessionPath, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.Header.Set(headerContentType, contentTypeJSON)
	if c.token != "" {
		request.Header.Set(headerAuthorization, c.token)
	}

	response, err := c.http.Do(request)
	if err != nil {
		return nil, fmt.Errorf("irma server: %w", err)
	}
	defer func() { _ = response.Body.Close() }()

	payload, err := io.ReadAll(io.LimitReader(response.Body, maxResponse))
	if err != nil {
		return nil, fmt.Errorf("irma server: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("irma server: status %d", response.StatusCode)
	}

	return payload, nil
}
