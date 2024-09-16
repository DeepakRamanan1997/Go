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
    keycloakURL  = "https://3.110.90.224:8443"
    clientID     = "Blue"
    clientSecret = "rMkVkPtb8JPpB9IY1BkoDP4zTcYSlU3J"
    realm        = "master"
)

var client *http.Client

func init() {
    tr := &http.Transport{
        TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // Note: Use this only for development
    }
    client = &http.Client{Transport: tr}
}

func checkAuthorization(token string) (bool, error) {
    authzURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", keycloakURL, realm)
    data := url.Values{}
    data.Set("grant_type", "urn:ietf:params:oauth:grant-type:uma-ticket")
    data.Set("audience", clientID)
    // We're not specifying any particular resource or scope here
    // This will check against all resources and scopes configured for this client
    req, err := http.NewRequest("POST", authzURL, strings.NewReader(data.Encode()))
    if err != nil {
        return false, err
    }
    req.Header.Add("Authorization", "Bearer "+token)
    req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
    resp, err := client.Do(req)
    if err != nil {
        return false, err
    }
    defer resp.Body.Close()
    if resp.StatusCode == http.StatusOK {
        return true, nil
    }
    body, _ := ioutil.ReadAll(resp.Body)
    log.Printf("Authorization denied. Response: %s", string(body))
    return false, nil
}

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        // Redirect to Keycloak authorization URL
        authURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/auth", keycloakURL, realm)
        params := url.Values{
            "client_id":     {clientID},
            "redirect_uri":  {"http://3.110.90.224:7000/callback"},
            "response_type": {"code"},
            "scope":         {"openid profile email"},
        }
        authURL += "?" + params.Encode()
        http.Redirect(w, r, authURL, http.StatusFound)
    })

    http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
        code := r.URL.Query().Get("code")
        if code == "" {
            http.Error(w, "Invalid authorization code", http.StatusBadRequest)
            return
        }

        // Exchange authorization code for access token
        tokenURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", keycloakURL, realm)
        data := url.Values{}
        data.Set("grant_type", "authorization_code")
        data.Set("code", code)
        data.Set("redirect_uri", "http://3.110.90.224:7000/callback")
        data.Set("client_id", clientID)
        data.Set("client_secret", clientSecret)
        resp, err := client.PostForm(tokenURL, data)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        defer resp.Body.Close()

        var tokenResp struct {
            AccessToken string `json:"access_token"`
        }
        if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        // Check authorization using Keycloak's Authorization Services
        authorized, err := checkAuthorization(tokenResp.AccessToken)
        if err != nil {
            http.Error(w, fmt.Sprintf("Authorization check failed: %v", err), http.StatusInternalServerError)
            return
        }
        if !authorized {
            http.Error(w, "Insufficient permissions", http.StatusForbidden)
            return
        }

        // If we've reached this point, the user is authenticated and authorized
        http.ServeFile(w, r, "index.html")
    })

    log.Fatal(http.ListenAndServe(":7000", nil))
}
