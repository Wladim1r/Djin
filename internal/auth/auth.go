package auth

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Роли пользователей
const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

type Region struct {
	ID    uint   `json:"id"              gorm:"primarykey"`
	Name  string `json:"name"            gorm:"not null;unique"`
	Users []User `json:"users,omitempty" gorm:"foreignKey:RegionID"`
}

type User struct {
	ID       uint   `json:"id"               gorm:"primarykey"`
	Username string `json:"username"         gorm:"not null;unique"`
	Password string `json:"-"                gorm:"not null"`
	Role     string `json:"role"             gorm:"not null;default:'user'"`
	RegionID uint   `json:"region_id"        gorm:"not null"`
	Region   Region `json:"region,omitempty" gorm:"foreignKey:RegionID"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	UserID     uint   `json:"user_id,omitempty"`
	Username   string `json:"username,omitempty"`
	Role       string `json:"role,omitempty"`
	RegionID   uint   `json:"region_id,omitempty"`
	RegionName string `json:"region_name,omitempty"`
}

// Структуры для управления пользователями
type CreateUserRequest struct {
	Username string `json:"username"  binding:"required"`
	Password string `json:"password"  binding:"required,min=3"`
	Role     string `json:"role"      binding:"required,oneof=user admin"`
	RegionID uint   `json:"region_id" binding:"required"`
}

type UpdateUserRequest struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Role     string `json:"role,omitempty"      binding:"omitempty,oneof=user admin"`
	RegionID uint   `json:"region_id,omitempty"`
}

type UserResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	RegionID uint   `json:"region_id"`
	Region   Region `json:"region"`
}

// Сервис авторизации
type AuthService struct {
	db *gorm.DB
}

func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{db: db}
}

func (s *AuthService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (s *AuthService) CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (s *AuthService) Authenticate(username, password string) (*User, error) {
	var user User
	err := s.db.Preload("Region").Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}

	if !s.CheckPassword(password, user.Password) {
		return nil, gorm.ErrRecordNotFound
	}

	return &user, nil
}

func (s *AuthService) CreateUser(username, password, role string, regionID uint) (*User, error) {
	// Проверяем существование региона
	var region Region
	if err := s.db.First(&region, regionID).Error; err != nil {
		return nil, errors.New("region not found")
	}

	hashedPassword, err := s.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := User{
		Username: username,
		Password: hashedPassword,
		Role:     role,
		RegionID: regionID,
	}

	err = s.db.Create(&user).Error
	return &user, err
}

func (s *AuthService) UpdateUser(userID uint, req UpdateUserRequest) (*User, error) {
	var user User
	if err := s.db.Preload("Region").First(&user, userID).Error; err != nil {
		return nil, err
	}

	// Обновляем только переданные поля
	updates := make(map[string]interface{})

	if req.Username != "" {
		updates["username"] = req.Username
	}

	if req.Password != "" {
		hashedPassword, err := s.HashPassword(req.Password)
		if err != nil {
			return nil, err
		}
		updates["password"] = hashedPassword
	}

	if req.Role != "" {
		updates["role"] = req.Role
	}

	if req.RegionID != 0 {
		// Проверяем существование региона
		var region Region
		if err := s.db.First(&region, req.RegionID).Error; err != nil {
			return nil, errors.New("region not found")
		}
		updates["region_id"] = req.RegionID
	}

	if err := s.db.Model(&user).Updates(updates).Error; err != nil {
		return nil, err
	}

	// Перезагружаем пользователя с актуальными данными
	if err := s.db.Preload("Region").First(&user, userID).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *AuthService) DeleteUser(userID uint) error {
	return s.db.Delete(&User{}, userID).Error
}

func (s *AuthService) GetAllUsers() ([]User, error) {
	var users []User
	err := s.db.Preload("Region").Find(&users).Error
	return users, err
}

func (s *AuthService) GetUserByID(id uint) (*User, error) {
	var user User
	err := s.db.Preload("Region").First(&user, id).Error
	return &user, err
}

func (s *AuthService) GetRegions() ([]Region, error) {
	var regions []Region
	err := s.db.Find(&regions).Error
	return regions, err
}

// Контроллер авторизации
type AuthController struct {
	authService *AuthService
}

func NewAuthController(authService *AuthService) *AuthController {
	return &AuthController{authService: authService}
}

func (ac *AuthController) ShowLoginForm(c *gin.Context) {
	if userID := sessions.Default(c).Get("user_id"); userID != nil {
		c.Redirect(http.StatusSeeOther, "/dashboard")
		return
	}

	c.HTML(http.StatusOK, "login.html", gin.H{
		"title": "Вход в систему",
	})
}

func (ac *AuthController) Login(c *gin.Context) {
	var req LoginRequest

	log.Printf("[LOGIN] Начало процедуры входа")
	log.Printf("[LOGIN] Content-Type: %s", c.GetHeader("Content-Type"))
	log.Printf("[LOGIN] User-Agent: %s", c.GetHeader("User-Agent"))
	log.Printf("[LOGIN] Remote Address: %s", c.ClientIP())

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, LoginResponse{
			Success: false,
			Message: "Неверный формат данных",
		})
		return
	}

	log.Printf("[LOGIN] Попытка авторизации пользователя: %s", req.Username)

	user, err := ac.authService.Authenticate(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, LoginResponse{
			Success: false,
			Message: "Неверное имя пользователя или пароль",
		})
		return
	}

	log.Printf("[LOGIN] Пользователь %s успешно авторизован с ролью %s", user.Username, user.Role)

	session := sessions.Default(c)
	session.Set("user_id", user.ID)
	session.Set("username", user.Username)
	session.Set("role", user.Role)
	session.Set("region_id", user.RegionID)
	session.Set("region_name", user.Region.Name)

	if err := session.Save(); err != nil {
		log.Printf("[LOGIN] Ошибка сохранения сессии: %v", err)
		c.JSON(http.StatusInternalServerError, LoginResponse{
			Success: false,
			Message: "Ошибка создания сессии",
		})
		return
	}

	log.Printf(
		"[LOGIN] Сессия сохранена для пользователя %s (ID: %d, Role: %s)",
		user.Username,
		user.ID,
		user.Role,
	)

	response := LoginResponse{
		Success:    true,
		Message:    "Успешный вход",
		UserID:     user.ID,
		Username:   user.Username,
		Role:       user.Role,
		RegionID:   user.RegionID,
		RegionName: user.Region.Name,
	}

	log.Printf("[LOGIN] Отправляем успешный ответ: %+v", response)
	c.JSON(http.StatusOK, response)
}

func (ac *AuthController) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Вы вышли из системы",
	})
}

func (ac *AuthController) GetCurrentUser(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	username := session.Get("username")
	role := session.Get("role")
	regionID := session.Get("region_id")
	regionName := session.Get("region_name")

	if userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Не авторизован",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"user_id":     userID,
		"username":    username,
		"role":        role,
		"region_id":   regionID,
		"region_name": regionName,
	})
}

// Административные методы
func (ac *AuthController) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Неверный формат данных",
			"error":   err.Error(),
		})
		return
	}

	user, err := ac.authService.CreateUser(req.Username, req.Password, req.Role, req.RegionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Не удалось создать пользователя",
			"error":   err.Error(),
		})
		return
	}

	// Загружаем пользователя с регионом для ответа
	fullUser, _ := ac.authService.GetUserByID(user.ID)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Пользователь успешно создан",
		"user": UserResponse{
			ID:       fullUser.ID,
			Username: fullUser.Username,
			Role:     fullUser.Role,
			RegionID: fullUser.RegionID,
			Region:   fullUser.Region,
		},
	})
}

func (ac *AuthController) UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	var req UpdateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Неверный формат данных",
			"error":   err.Error(),
		})
		return
	}

	// Конвертируем строку в uint
	var id uint
	if _, err := fmt.Sscanf(userID, "%d", &id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Неверный ID пользователя",
		})
		return
	}

	user, err := ac.authService.UpdateUser(id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Не удалось обновить пользователя",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Пользователь успешно обновлен",
		"user": UserResponse{
			ID:       user.ID,
			Username: user.Username,
			Role:     user.Role,
			RegionID: user.RegionID,
			Region:   user.Region,
		},
	})
}

func (ac *AuthController) DeleteUser(c *gin.Context) {
	userID := c.Param("id")

	var id uint
	if _, err := fmt.Sscanf(userID, "%d", &id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Неверный ID пользователя",
		})
		return
	}

	if err := ac.authService.DeleteUser(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Не удалось удалить пользователя",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Пользователь успешно удален",
	})
}

func (ac *AuthController) GetAllUsers(c *gin.Context) {
	users, err := ac.authService.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Не удалось получить список пользователей",
			"error":   err.Error(),
		})
		return
	}

	var response []UserResponse
	for _, user := range users {
		response = append(response, UserResponse{
			ID:       user.ID,
			Username: user.Username,
			Role:     user.Role,
			RegionID: user.RegionID,
			Region:   user.Region,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"users":   response,
	})
}

func (ac *AuthController) GetRegions(c *gin.Context) {
	regions, err := ac.authService.GetRegions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Не удалось получить список регионов",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"regions": regions,
	})
}

func (ac *AuthController) GetMe(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	username := session.Get("username")
	role := session.Get("role")
	regionID := session.Get("region_id")
	regionName := session.Get("region_name")

	if userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Не авторизован",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"user_id":     userID,
		"username":    username,
		"role":        role,
		"region_id":   regionID,
		"region_name": regionName,
	})
}

// Middleware для проверки авторизации
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("[AUTH_MIDDLEWARE] Проверка авторизации для URL: %s", c.Request.URL.Path)

		session := sessions.Default(c)
		userID := session.Get("user_id")

		log.Printf("[AUTH_MIDDLEWARE] UserID из сессии: %v", userID)

		if userID == nil {
			log.Printf("[AUTH_MIDDLEWARE] Пользователь не авторизован, перенаправление на /login")

			// Для AJAX запросов возвращаем JSON
			if c.GetHeader("Content-Type") == "application/json" ||
				c.GetHeader("Accept") == "application/json" {
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"message": "Требуется авторизация",
				})
				c.Abort()
				return
			}

			// Для обычных запросов перенаправляем на страницу входа
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}

		// Добавляем информацию о пользователе в контекст
		c.Set("user_id", userID)
		c.Set("username", session.Get("username"))
		c.Set("role", session.Get("role"))
		c.Set("region_id", session.Get("region_id"))
		c.Set("region_name", session.Get("region_name"))

		c.Next()
	}
}

// Middleware для проверки роли администратора
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("role")
		if role != RoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Доступ запрещен. Требуется роль администратора",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// Middleware для извлечения regionID
func RegionContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if regionID, exists := c.Get("region_id"); exists {
			if id, ok := regionID.(uint); ok {
				c.Set("region_id_uint", id)
			}
		}
		c.Next()
	}
}

// Middleware для инъекции regionID в контекст для ваших существующих handlers
func InjectRegionID(handler gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		if regionID, exists := c.Get("region_id"); exists {
			c.Set("current_region_id", regionID.(uint))
		}
		handler(c)
	}
}

// Хелпер для получения regionID из контекста
func GetRegionIDFromContext(c *gin.Context) uint {
	if regionID, exists := c.Get("region_id_uint"); exists {
		return regionID.(uint)
	}
	if regionID, exists := c.Get("current_region_id"); exists {
		return regionID.(uint)
	}
	return 0
}

// Хелпер для получения userID из контекста
func GetUserIDFromContext(c *gin.Context) uint {
	if userID, exists := c.Get("user_id"); exists {
		return userID.(uint)
	}
	return 0
}

// Хелпер для получения роли пользователя из контекста
func GetUserRoleFromContext(c *gin.Context) string {
	if role, exists := c.Get("role"); exists {
		return role.(string)
	}
	return RoleUser
}

// Проверка, является ли пользователь администратором
func IsAdmin(c *gin.Context) bool {
	return GetUserRoleFromContext(c) == RoleAdmin
}

// Инициализация регионов и пользователей
func InitializeRegions(db *gorm.DB) error {
	authService := NewAuthService(db)

	// Создаем регионы
	regions := []struct {
		Name string
	}{
		{"Санкт-Петербург"},
		{"Тихорецк"},
	}

	regionMap := make(map[string]uint)

	for _, r := range regions {
		var existing Region
		if err := db.Where("name = ?", r.Name).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				region := Region{Name: r.Name}
				if err := db.Create(&region).Error; err != nil {
					return err
				}
				regionMap[r.Name] = region.ID
			}
		} else {
			regionMap[r.Name] = existing.ID
		}
	}

	// Создаем тестовых пользователей
	users := []struct {
		Username   string
		Password   string
		Role       string
		RegionName string
	}{
		{"ivan", "123", RoleAdmin, "Санкт-Петербург"},
		{"anna", "123", RoleUser, "Санкт-Петербург"},
		{"man", "123", RoleUser, "Санкт-Петербург"},

		{"Дима С", "123", RoleUser, "Тихорецк"},
		{"Дима Т", "123", RoleUser, "Тихорецк"},
		{"Боря", "123", RoleUser, "Тихорецк"},
		{"Даня", "123", RoleUser, "Тихорецк"},
		{"Андрей", "123", RoleAdmin, "Тихорецк"},
	}

	for _, u := range users {
		var existing User
		if err := db.Where("username = ?", u.Username).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				regionID := regionMap[u.RegionName]
				_, err := authService.CreateUser(u.Username, u.Password, u.Role, regionID)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
