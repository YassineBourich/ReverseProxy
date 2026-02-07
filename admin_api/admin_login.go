package adminapi

import (
	"encoding/json"
	"net/http"
	"os"
	errors "reverse_proxy/CustomErrors"
	"github.com/golang-jwt/jwt/v5"		// JSON Web Token
	"golang.org/x/crypto/bcrypt"		// Bcrypt for hashing
	"time"
	"fmt"
	"path/filepath"
)

// Defining the secret key used for authentication JWT tokens
var secret_key = []byte("Alifoun")

type Admin struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func verify_password(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// Function to generate JWT token
func GenerateToken(username string) (string, error) {
	// Create the Claims (the data inside the token)
	claims := jwt.MapClaims{
		"sub": username,
		"exp": time.Now().Add(time.Hour * 2).Unix(), // Expiration (2 hours)
		"iat": time.Now().Unix(),                    // Issued at
	}

	// Create the token object with a signing method
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the defined secret key
	return token.SignedString(secret_key)
}

func isTokenValid(tokenString string) bool {
	// Parsing the token to populate 'Valid' attribute in it
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
        // Validate the algorithm
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return secret_key, nil
    })

    // If err is nil and token.Valid is true, the signature is correct 
    // and the expiration has not passed.
    return err == nil && token.Valid
}

// Login handler function
func HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Getting credentials from configuration file
		cred_file, err := os.ReadFile(filepath.Join("config", "admin.json"))
		if err != nil {
			http.Error(w, errors.HttpError(http.StatusInternalServerError).Error(), http.StatusInternalServerError)
			return
		}

		// Unmarshal credentials into a variable of type Admin
		var admin Admin
		if err := json.Unmarshal(cred_file, &admin); err != nil {
			http.Error(w, errors.HttpError(http.StatusInternalServerError).Error(), http.StatusInternalServerError)
			return
		}
		// Reading credentials sent by the user
		defer r.Body.Close()
		var user Admin
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			http.Error(w, errors.HttpError(http.StatusInternalServerError).Error(), http.StatusInternalServerError)
			return
		}

		// Validating credentials
		if user.Username != admin.Username {
			http.Error(w, errors.HttpError(http.StatusUnauthorized).Error(), http.StatusUnauthorized)
			return
		}
		if !verify_password(admin.Password, user.Password) {
			http.Error(w, errors.HttpError(http.StatusUnauthorized).Error(), http.StatusUnauthorized)
			return
		}
		// Responding with a token
		token, err := GenerateToken(user.Username)
		if err != nil {
			http.Error(w, errors.HttpError(http.StatusInternalServerError).Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte(token))
	} else {
		http.Error(w, errors.HttpError(http.StatusMethodNotAllowed).Error(), http.StatusMethodNotAllowed)
	}
}