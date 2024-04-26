package service

import (
	"github.com/golang-jwt/jwt/v5"
	"time"
	"apollo/model"
	"log"
	"strings"
	"os"
)

var secretKey = []byte(os.Getenv("SECRET_KEY"))

func CreateToken(username string, permissions string) (string, error) {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, 
        jwt.MapClaims{ 
        "username": username, 	// subject scope id
		"permissions": permissions,
        "exp": time.Now().Add(time.Hour * 24).Unix(), 
        })

    tokenString, err := token.SignedString(secretKey)
    if err != nil {
    	return "", err
    }

 	return tokenString, nil
}

func GetClaimsFromJwt(req model.Token) []string {
	token, err := jwt.Parse(req.Token, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	var emptyPermissions []string
	if err != nil {
		log.Printf("Error parsing token: %v", err)
		return emptyPermissions
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if permissions, ok := claims["permissions"].(string); ok {
			return strings.Split(permissions, ",")
		} else {
			log.Println("Custom Claim permissions is not a string or does not exist.")
			return emptyPermissions
		}
	} else {
		log.Println("Invalid claims type.")
		return emptyPermissions
	}

}