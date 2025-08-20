package auth

import (
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
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
	RegionID   uint   `json:"region_id,omitempty"`
	RegionName string `json:"region_name,omitempty"`
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

func (s *AuthService) CreateUser(username, password string, regionID uint) (*User, error) {
	hashedPassword, err := s.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := User{
		Username: username,
		Password: hashedPassword,
		RegionID: regionID,
	}

	err = s.db.Create(&user).Error
	return &user, err
}

func (s *AuthService) GetUserByID(id uint) (*User, error) {
	var user User
	err := s.db.Preload("Region").First(&user, id).Error
	return &user, err
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

	log.Printf("[LOGIN] Пользователь %s успешно авторизован", user.Username)

	session := sessions.Default(c)
	session.Set("user_id", user.ID)
	session.Set("username", user.Username)
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

	log.Printf("[LOGIN] Сессия сохранена для пользователя %s (ID: %d)", user.Username, user.ID)

	response := LoginResponse{
		Success:    true,
		Message:    "Успешный вход",
		UserID:     user.ID,
		Username:   user.Username,
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
		c.Set("region_id", session.Get("region_id"))
		c.Set("region_name", session.Get("region_name"))

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

// Инициализация регионов и пользователей
func InitializeRegions(db *gorm.DB) error {
	authService := NewAuthService(db)

	// Создаем регионы
	regions := []struct {
		Name string
	}{
		{"Московская область"},
		{"Санкт-Петербург"},
		{"Краснодарский край"},
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
		RegionName string
	}{
		{"ivan", "123", "Московская область"},
		{"ivan", "1234", "Санкт-Петербург"},
		{"anna", "123", "Санкт-Петербург"},
		{"elena", "123", "Санкт-Петербург"},
		// Дополнительные пользователи для демонстрации
		{"admin_msk", "123", "Московская область"},
		{"man", "123", "Санкт-Петербург"},
	}

	for _, u := range users {
		var existing User
		if err := db.Where("username = ?", u.Username).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				regionID := regionMap[u.RegionName]
				_, err := authService.CreateUser(u.Username, u.Password, regionID)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (ac *AuthController) GetMe(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	username := session.Get("username")
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
		"region_id":   regionID,
		"region_name": regionName,
	})
}
