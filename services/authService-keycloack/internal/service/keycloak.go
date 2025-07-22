package service

import (
	"auth_service/internal/config"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"net/url"
)

type Credentials struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func LoginUser(c *gin.Context, cfg *config.Config) (map[string]any, error) {
	var creds Credentials
	if err := c.ShouldBindJSON(&creds); err != nil {
		return nil, errors.New("неправильный формат запроса")
	}

	endpoint := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", cfg.KeycloakURL, cfg.KeycloakRealm)
	data := url.Values{
		"grant_type":    {"password"},
		"client_id":     {cfg.ClientID},
		"client_secret": {cfg.ClientSecret},
		//"email":         {creds.Email},
		"username": {creds.Username},
		"password": {creds.Password},
	}

	resp, err := http.PostForm(endpoint, data)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса к Keycloak: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ошибка авторизации: %s", string(bodyBytes))
	}

	var result map[string]any
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, errors.New("не удалось декодировать ответ от keycloak")
	}

	return result, nil
}

func RefreshToken(c *gin.Context, cfg *config.Config) (map[string]any, error) {
	var input struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		return nil, errors.New("неправильный формат запроса")
	}

	endpoint := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", cfg.KeycloakURL, cfg.KeycloakRealm)
	data := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {cfg.ClientID},
		"client_secret": {cfg.ClientSecret},
		"refresh_token": {input.RefreshToken},
	}

	resp, err := http.PostForm(endpoint, data)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса к Keycloak: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ошибка обновления токена: %s", string(bodyBytes))
	}

	var result map[string]any
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, errors.New("не удалось декодировать ответ от keycloak")
	}

	return result, nil
}

func RegisterUser(c *gin.Context, cfg *config.Config) error {
	var input Credentials
	if err := c.ShouldBindJSON(&input); err != nil {
		return errors.New("неправильный формат запроса")
	}

	tokenResp, err := getAdminToken(cfg)
	if err != nil {
		return err
	}

	client := &http.Client{}
	authHeader := "Bearer " + tokenResp

	userData := map[string]any{
		"username": input.Username,
		"email":    input.Email,
		"enabled":  true,
		"credentials": []map[string]any{
			{"type": "password", "value": input.Password, "temporary": false},
		},
	}
	userJSON, _ := json.Marshal(userData)

	req, err := http.NewRequest("POST",
		fmt.Sprintf("%s/admin/realms/%s/users", cfg.KeycloakURL, cfg.KeycloakRealm),
		bytes.NewBuffer(userJSON),
	)
	if err != nil {
		return fmt.Errorf("ошибка формирования запроса на создание пользователя: %v", err)
	}
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка запроса на создание пользователя: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("не удалось создать пользователя: %s", string(body))
	}

	// Получение user ID
	getReq, _ := http.NewRequest("GET",
		fmt.Sprintf("%s/admin/realms/%s/users?username=%s",
			cfg.KeycloakURL, cfg.KeycloakRealm, url.QueryEscape(input.Username)),
		nil)
	getReq.Header.Set("Authorization", authHeader)
	resp, err = client.Do(getReq)
	if err != nil {
		return errors.New("ошибка получения ID пользователя")
	}
	defer resp.Body.Close()

	var users []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil || len(users) == 0 {
		return errors.New("не удалось найти созданного пользователя")
	}
	userID := users[0]["id"].(string)

	// Получение роли user
	roleReq, _ := http.NewRequest("GET",
		fmt.Sprintf("%s/admin/realms/%s/roles/user", cfg.KeycloakURL, cfg.KeycloakRealm), nil)
	roleReq.Header.Set("Authorization", authHeader)
	resp, err = client.Do(roleReq)
	if err != nil {
		return errors.New("ошибка получения роли user")
	}
	defer resp.Body.Close()

	var role map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&role); err != nil {
		return errors.New("ошибка разбора роли user")
	}

	// Назначение роли
	roleData := []map[string]any{{
		"id":          role["id"],
		"name":        role["name"],
		"composite":   role["composite"],
		"clientRole":  role["clientRole"],
		"containerId": role["containerId"],
	}}
	mapJSON, _ := json.Marshal(roleData)

	assignReq, _ := http.NewRequest("POST",
		fmt.Sprintf("%s/admin/realms/%s/users/%s/role-mappings/realm", cfg.KeycloakURL, cfg.KeycloakRealm, userID),
		bytes.NewBuffer(mapJSON),
	)
	assignReq.Header.Set("Authorization", authHeader)
	assignReq.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(assignReq)
	if err != nil {
		return fmt.Errorf("ошибка назначения роли: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("не удалось назначить роль: %s", string(body))
	}

	return nil
}

func getAdminToken(cfg *config.Config) (string, error) {
	data := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {cfg.ClientID},
		"client_secret": {cfg.ClientSecret},
	}
	endpoint := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", cfg.KeycloakURL, cfg.KeycloakRealm)

	resp, err := http.PostForm(endpoint, data)
	if err != nil {
		return "", fmt.Errorf("ошибка получения токена администратора: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("не удалось получить админ токен: %s", string(body))
	}

	var tokenData map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&tokenData); err != nil {
		return "", errors.New("ошибка разбора токена администратора")
	}

	accessToken, ok := tokenData["access_token"].(string)
	if !ok {
		return "", errors.New("не найден access_token в ответе")
	}
	return accessToken, nil
}

func LogoutUser(c *gin.Context, cfg *config.Config) error {
	var input struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		return errors.New("неправильный формат запроса")
	}

	endpoint := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/logout", cfg.KeycloakURL, cfg.KeycloakRealm)
	data := url.Values{
		"client_id":     {cfg.ClientID},
		"client_secret": {cfg.ClientSecret},
		"refresh_token": {input.RefreshToken},
	}

	resp, err := http.PostForm(endpoint, data)
	if err != nil {
		return fmt.Errorf("ошибка запроса к Keycloak (logout): %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("не удалось выполнить logout: %s", string(body))
	}

	return nil
}
