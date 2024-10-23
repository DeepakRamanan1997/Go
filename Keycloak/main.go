package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

const (
	keycloakURL  = "https://15.207.114.27:8443" // Keycloak server URL
	clientID     = "Blue"                       // Client ID for Keycloak
	clientSecret = "rMkVkPtb8JPpB9IY1BkoDP4zTcYSlU3J" // Client secret for Keycloak
	realm        = "master"                     // Realm in Keycloak
)

var client *http.Client

// init function sets up the HTTP client with insecure TLS verification for development
func init() {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // Note: Use this only for development
	}
	client = &http.Client{Transport: tr} // Create a new HTTP client
}

// checkAuthorization checks if the provided token has permission for the specified resource
func checkAuthorization(token, resource string) (bool, error) {
	authzURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", keycloakURL, realm)
	data := url.Values{}
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:uma-ticket")
	data.Set("audience", clientID) // Set the audience to the client ID
	data.Set("permission", resource) // Set the resource for permission check

	// Create a new POST request for authorization
	req, err := http.NewRequest("POST", authzURL, strings.NewReader(data.Encode()))
	if err != nil {
		return false, err // Return error if request creation fails
	}
	req.Header.Add("Authorization", "Bearer "+token) // Add the Bearer token to the header
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded") // Set content type

	resp, err := client.Do(req) // Send the request
	if err != nil {
		return false, err // Return error if request fails
	}
	defer resp.Body.Close() // Ensure the response body is closed

	// Check if the response status is OK
	if resp.StatusCode == http.StatusOK {
		return true, nil // Authorization successful
	}
	body, _ := ioutil.ReadAll(resp.Body) // Read the response body
	log.Printf("Authorization denied for resource %s. Response: %s", resource, string(body))
	return false, nil // Authorization denied
}

// serveProtectedResource serves the requested resource if the token is valid
func serveProtectedResource(w http.ResponseWriter, r *http.Request, resource string) {
	token := r.URL.Query().Get("token") // Get the token from the query parameters
	if token == "" {
		// Redirect to Keycloak login if no token is provided
		authURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/auth", keycloakURL, realm)
		params := url.Values{
			"client_id":     {clientID},
			"redirect_uri":  {"http://15.207.114.27:7000/callback"},
			"response_type": {"code"},
			"scope":         {"openid profile email"},
		}
		http.Redirect(w, r, authURL+"?"+params.Encode(), http.StatusFound) // Redirect to login
		return
	}

	// Check if the token is authorized for the requested resource
	authorized, err := checkAuthorization(token, resource)
	if err != nil {
		http.Error(w, fmt.Sprintf("Authorization check failed: %v", err), http.StatusInternalServerError)
		return // Return error if authorization check fails
	}
	if !authorized {
		http.Error(w, "Insufficient permissions", http.StatusForbidden) // Return forbidden if not authorized
		return
	}

	// Serve the requested resource (e.g., an image file)
	http.ServeFile(w, r, resource+".jpg")
}

func main() {
	// Handle the root URL
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token") // Get the token from the query parameters
		if token == "" {
			// Redirect to Keycloak login if no token is provided
			authURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/auth", keycloakURL, realm)
			params := url.Values{
				"client_id":     {clientID},
				"redirect_uri":  {"http://15.207.114.27:7000/callback"},
				"response_type": {"code "},
				"scope":         {"openid profile email"},
			}
			http.Redirect(w, r, authURL+"?"+params.Encode(), http.StatusFound) // Redirect to login
			return
		}

		// Check if the token is authorized for the root resource
		authorized, err := checkAuthorization(token, "root")
		if err != nil {
			http.Error(w, fmt.Sprintf("Authorization check failed: %v", err), http.StatusInternalServerError)
			return // Return error if authorization check fails
		}
		if !authorized {
			http.Error(w, "Insufficient permissions", http.StatusForbidden) // Return forbidden if not authorized
			return
		}

		// Display links to protected resources
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `
            <html>
                <body>
                    <h1>Protected Resources</h1>
                    <ul>
                        <li><a href="/sea?token=%s">Sea</a></li>
                        <li><a href="/mountain?token=%s">Mountain</a></li>
                    </ul>
                </body>
            </html>
        `, token, token)
	})

	// Handle the callback URL for authorization code exchange
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code") // Get the authorization code
		if code == "" {
			http.Error(w, "Invalid authorization code", http.StatusBadRequest) // Return bad request if code is empty
			return
		}

		// Exchange the authorization code for an access token
		tokenURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", keycloakURL, realm)
		data := url.Values{}
		data.Set("grant_type", "authorization_code")
		data.Set("code", code)
		data.Set("redirect_uri", "http://15.207.114.27:7000/callback")
		data.Set("client_id", clientID)
		data.Set("client_secret", clientSecret)
		resp, err := client.PostForm(tokenURL, data) // Send the request
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError) // Return error if request fails
			return
		}
		defer resp.Body.Close() // Ensure the response body is closed

		var tokenResp struct {
			AccessToken string `json:"access_token"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError) // Return error if token decoding fails
			return
		}

		// Redirect back to the root URL with the obtained access token
		http.Redirect(w, r, fmt.Sprintf("/?token=%s", tokenResp.AccessToken), http.StatusFound)
	})

	// Handle protected resources
	http.HandleFunc("/sea", func(w http.ResponseWriter, r *http.Request) {
		serveProtectedResource(w, r, "sea")
	})

	http.HandleFunc("/mountain", func(w http.ResponseWriter, r *http.Request) {
		serveProtectedResource(w, r, "mountain")
	})

	log.Fatal(http.ListenAndServe(":7000", nil)) // Start the HTTP server
}
