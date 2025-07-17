package main

import (
	"bytes"
	"encoding/json"
	"github.com/MicahParks/keyfunc/v2"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var jwks *keyfunc.JWKS

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Ошибка чтения .env файла")
	}

	realm := os.Getenv("KEYCLOAK_REALM")
	jwksURL := os.Getenv("KEYCLOAK_URL") + "/realms/" + realm + "/protocol/openid-connect/certs"

	jwks, err = keyfunc.Get(jwksURL, keyfunc.Options{
		RefreshInterval: time.Hour,
		RefreshErrorHandler: func(err error) {
			log.Printf("Ошибка обновления JWKS: %v", err)
		},
	})
	if err != nil {
		log.Fatalf("Ошибка получения JWKS: %v", err)
	}

	router := gin.Default()

	router.POST("/login", loginHandler)
	router.POST("/refresh", refreshHandler)
	router.POST("/register", registerHandler)

	auth := router.Group("/auth")
	auth.Use(JWTMiddleware())
	auth.GET("/profile", profileHandler)

	log.Println("Сервер на :3000")
	router.Run(":3000")
}

func loginHandler(c *gin.Context) {
	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.BindJSON(&credentials); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неправильный формат"})
		return
	}

	tokenEndpoint := os.Getenv("KEYCLOAK_URL") + "/realms/" + os.Getenv("KEYCLOAK_REALM") + "/protocol/openid-connect/token"
	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("client_id", os.Getenv("KEYCLOAK_CLIENT_ID"))
	data.Set("client_secret", os.Getenv("KEYCLOAK_CLIENT_SECRET"))
	data.Set("username", credentials.Username)
	data.Set("password", credentials.Password)

	resp, err := http.PostForm(tokenEndpoint, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка запроса к Keycloak"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.JSON(http.StatusUnauthorized, gin.H{"error": string(body)})
		return
	}

	var body map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка декодирования ответа"})
		return
	}

	c.JSON(http.StatusOK, body)
}

func refreshHandler(c *gin.Context) {
	var input struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неправильный формат"})
		return
	}
	tokenEndpoint := os.Getenv("KEYCLOAK_URL") + "/realms/" + os.Getenv("KEYCLOAK_REALM") + "/protocol/openid-connect/token"
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", os.Getenv("KEYCLOAK_CLIENT_ID"))
	data.Set("client_secret", os.Getenv("KEYCLOAK_CLIENT_SECRET"))
	data.Set("refresh_token", input.RefreshToken)

	resp, err := http.PostForm(tokenEndpoint, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка запроса к Keycloak"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.JSON(http.StatusUnauthorized, gin.H{"error": string(body)})
		return
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка декодирования ответа"})
		return
	}

	c.JSON(http.StatusOK, body)
}

func registerHandler(c *gin.Context) {
	var input struct {
		Username  string `json:"username"`
		Password  string `json:"password"`
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неправильный формат"})
		return
	}

	client := &http.Client{}
	tokenEndpoint := os.Getenv("KEYCLOAK_URL") +
		"/realms/" + os.Getenv("KEYCLOAK_REALM") +
		"/protocol/openid-connect/token"

	data := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {os.Getenv("KEYCLOAK_CLIENT_ID")},
		"client_secret": {os.Getenv("KEYCLOAK_CLIENT_SECRET")},
	}

	resp, err := http.PostForm(tokenEndpoint, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка получения токена администратора"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.JSON(http.StatusInternalServerError, gin.H{"error": string(body)})
		return
	}

	var tokenData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&tokenData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка парсинга токена администратора"})
		return
	}
	adminToken := tokenData["access_token"].(string)
	authHeader := "Bearer " + adminToken

	// 1️⃣ Создаём пользователя с именем, фамилией, email и паролем
	userData := map[string]interface{}{
		"username":  input.Username,
		"email":     input.Email,
		"firstName": input.FirstName,
		"lastName":  input.LastName,
		"enabled":   true,
		"credentials": []map[string]interface{}{
			{"type": "password", "value": input.Password, "temporary": false},
		},
	}
	userJSON, _ := json.Marshal(userData)

	createResp, err := http.NewRequest("POST",
		os.Getenv("KEYCLOAK_URL")+"/admin/realms/"+os.Getenv("KEYCLOAK_REALM")+"/users",
		bytes.NewBuffer(userJSON),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка создания запроса"})
		return
	}
	createResp.Header.Set("Authorization", authHeader)
	createResp.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(createResp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка создания пользователя"})
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		c.JSON(http.StatusBadRequest, gin.H{"error": string(body)})
		return
	}

	// 2️⃣ Получаем ID пользователя по username
	req, _ := http.NewRequest("GET",
		os.Getenv("KEYCLOAK_URL")+"/admin/realms/"+os.Getenv("KEYCLOAK_REALM")+
			"/users?username="+url.QueryEscape(input.Username),
		nil,
	)
	req.Header.Set("Authorization", authHeader)
	resp, err = client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка поиска пользователя"})
		return
	}
	defer resp.Body.Close()

	var users []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil || len(users) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "пользователь не найден"})
		return
	}
	userID := users[0]["id"].(string)

	// 3️⃣ Получаем данные роли user
	roleReq, _ := http.NewRequest("GET",
		os.Getenv("KEYCLOAK_URL")+"/admin/realms/"+os.Getenv("KEYCLOAK_REALM")+"/roles/user", nil)
	roleReq.Header.Set("Authorization", authHeader)
	resp, err = client.Do(roleReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка получения роли"})
		return
	}
	defer resp.Body.Close()

	var roleData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&roleData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось распарсить роль"})
		return
	}

	// 4️⃣ Назначаем роль пользователю
	roleMapping := []map[string]interface{}{{
		"id":          roleData["id"],
		"name":        roleData["name"],
		"composite":   roleData["composite"],
		"clientRole":  roleData["clientRole"],
		"containerId": roleData["containerId"],
	}}
	mapJSON, _ := json.Marshal(roleMapping)
	assignReq, _ := http.NewRequest("POST",
		os.Getenv("KEYCLOAK_URL")+"/admin/realms/"+os.Getenv("KEYCLOAK_REALM")+
			"/users/"+userID+"/role-mappings/realm",
		bytes.NewBuffer(mapJSON),
	)
	assignReq.Header.Set("Authorization", authHeader)
	assignReq.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(assignReq)
	if err != nil || resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		c.JSON(http.StatusBadRequest, gin.H{"error": "ошибка назначения роли: " + string(body)})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "user создан с ролью user"})
}

func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "нет токена"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenString, jwks.Keyfunc)
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "невалидный токен"})
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		c.Set("claims", claims)
		c.Next()
	}
}

func profileHandler(c *gin.Context) {
	claims, _ := c.Get("claims")
	c.JSON(http.StatusOK, gin.H{"claims": claims})
}
