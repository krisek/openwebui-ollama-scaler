package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "strconv"
    "sync"
    "time"
)

// Response represents the response from the API call
type Response struct {
    ModelIDs []string `json:"model_ids"`
    UserIDs  []string `json:"user_ids"`
}

var (
    cacheLock      sync.Mutex // To handle concurrency
    cachedActiveUsers int
    lastUpdateTime time.Time
    cacheTimeout   time.Duration
)

// GetActiveUsers fetches the active users from the Open Web-UI API
func GetActiveUsers(apiURL, token string) (int, error) {
    req, err := http.NewRequest("GET", apiURL+"/api/usage", nil)
    if err != nil {
        return 0, err
    }
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return 0, err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return 0, fmt.Errorf("error fetching active users: %s", resp.Status)
    }
    var response Response
    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        return 0, err
    }
    return len(response.UserIDs), nil
}

// API handler to expose the active users count with caching
func ActiveUsersHandler(w http.ResponseWriter, r *http.Request) {
    apiURL := os.Getenv("API_URL")
    token := os.Getenv("TOKEN")
    if apiURL == "" {
        http.Error(w, "API_URL environment variable is not set", http.StatusInternalServerError)
        return
    }
    if token == "" {
        http.Error(w, "TOKEN environment variable is not set", http.StatusInternalServerError)
        return
    }
    cacheLock.Lock()
    defer cacheLock.Unlock()
    // Check if the cache is still valid
    if time.Since(lastUpdateTime) < cacheTimeout {
        log.Println("Returning cached result")
        response := map[string]int{
            "active_users": cachedActiveUsers,
        }
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(response)
        return
    }
    // Otherwise, compute the active users and update the cache
    activeUsers, err := GetActiveUsers(apiURL, token)
    if err != nil {
        http.Error(w, fmt.Sprintf("error counting active users: %v", err), http.StatusInternalServerError)
        return
    }
    // Update cache
    cachedActiveUsers = activeUsers
    lastUpdateTime = time.Now()
    // Respond with the updated count in JSON format
    response := map[string]int{
        "active_users": activeUsers,
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func main() {
    // Set cache timeout from the environment variable
    timeoutStr := os.Getenv("CACHE_TIMEOUT")
    if timeoutStr == "" {
        timeoutStr = "60" // Default cache timeout is 60 seconds
    }
    timeout, err := strconv.Atoi(timeoutStr)
    if err != nil {
        log.Printf("Invalid CACHE_TIMEOUT: %v - using default", err)
        timeout = 60
    }
    cacheTimeout = time.Duration(timeout) * time.Second
    // Set up the HTTP server and route
    http.HandleFunc("/active_users", ActiveUsersHandler)
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    log.Printf("Starting server on port %s with cache timeout of %d seconds...", port, timeout)
    if err := http.ListenAndServe(":"+port, nil); err != nil {
        log.Fatalf("Server failed: %v", err)
    }
}