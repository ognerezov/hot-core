package aws

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/ognerezov/hot-core/tools"
	"github.com/rs/zerolog/log"
)

// GraphQLRequest represents a standard GraphQL request structure.
type GraphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

// withSysInfo adds system information (like timestamp) to GraphQL variables.
func withSysInfo(variables map[string]any) map[string]any {
	variables["timestamp"] = time.Now().UTC().String()
	return variables
}

// SendAppSyncMutation sends a GraphQL mutation to AppSync using IAM authorization (SigV4)
func SendAppSyncMutation(ctx context.Context, query string, _variables map[string]any) error {
	appSyncUrl := os.Getenv("APPSYNC_URL")
	variables := withSysInfo(_variables)

	reqBody := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}
	jsonBody, err := json.Marshal(reqBody)
	log.Info().Msgf("Sending AppSync request: %s to %s", string(jsonBody), appSyncUrl)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", appSyncUrl, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	// IMPORTANT: Explicitly set Host header for SigV4 signing
	req.Header.Set("Host", req.URL.Host)

	// Load AWS Config
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}

	creds, err := cfg.Credentials.Retrieve(ctx)
	if err != nil {
		return err
	}

	// Compute SHA256 hash of the body
	hash := sha256.Sum256(jsonBody)
	payloadHash := hex.EncodeToString(hash[:])

	// Sign the request
	signer := v4.NewSigner()
	// Log for debugging
	log.Info().Msgf("Signing AppSync request. Region: %s, URL: %s", cfg.Region, appSyncUrl)

	err = signer.SignHTTP(ctx, creds, req, payloadHash, "appsync", cfg.Region, time.Now())
	if err != nil {
		return fmt.Errorf("failed to sign request: %w", err)
	}

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	tools.CloseAny(resp.Body)

	if resp.StatusCode != http.StatusOK {
		// Read body for error details
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		log.Error().Msgf("AppSync Error Body: %s", buf.String())
		return fmt.Errorf("appsync returned status %d", resp.StatusCode)
	}

	return nil
}
