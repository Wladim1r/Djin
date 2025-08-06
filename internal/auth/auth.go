package auth

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Region struct {
	ID       uint   `json:"id"    gorm:"primarykey"`
	Name     string `json:"name"  gorm:"not null;unique"`
	Login    string `json:"login" gorm:"not null;unique"`
	Password string `json:"-"     gorm:"not null"`
}

type LoginRequest struct {
	Login    string `json:"login"    binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
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

func (s *AuthService) Authenticate(login, password string) (*Region, error) {
	var region Region

	err := s.db.Where("login = ?", login).First(&region).Error
	if err != nil {
		return nil, err
	}

	if !s.CheckPassword(password, region.Password) {
		return nil, gorm.ErrRecordNotFound
	}

	return &region, nil
}

func (s *AuthService) CreateRegion(name, login, password string) (*Region, error) {
	hashedPassword, err := s.HashPassword(password)
	if err != nil {
		return nil, err
	}

	region := Region{
		Name:     name,
		Login:    login,
		Password: hashedPassword,
	}

	err = s.db.Create(&region).Error
	return &region, err
}

func (s *AuthService) GetRegionByID(id uint) (*Region, error) {
	var region Region
	err := s.db.First(&region, id).Error
	return &region, err
}

// Контроллер авторизации
type AuthController struct {
	authService *AuthService
}

func NewAuthController(authService *AuthService) *AuthController {
	return &AuthController{authService: authService}
}

func (ac *AuthController) ShowLoginForm(c *gin.Context) {
	if regionID := sessions.Default(c).Get("region_id"); regionID != nil {
		c.Redirect(http.StatusSeeOther, "/dashboard")
		return
	}

	c.HTML(http.StatusOK, "login.html", gin.H{
		"title": "Вход в систему",
	})
}

func (ac *AuthController) Login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, LoginResponse{
			Success: false,
			Message: "Неверный формат данных",
		})
		return
	}

	region, err := ac.authService.Authenticate(req.Login, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, LoginResponse{
			Success: false,
			Message: "Неверный логин или пароль",
		})
		return
	}

	session := sessions.Default(c)
	session.Set("region_id", region.ID)
	session.Set("region_name", region.Name)
	session.Set("region_login", region.Login)
	session.Save()

	c.JSON(http.StatusOK, LoginResponse{
		Success:    true,
		Message:    "Успешный вход",
		RegionID:   region.ID,
		RegionName: region.Name,
	})
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
	regionID := session.Get("region_id")
	regionName := session.Get("region_name")
	regionLogin := session.Get("region_login")

	if regionID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Не авторизован",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"region_id":    regionID,
		"region_name":  regionName,
		"region_login": regionLogin,
	})
}

// Middleware для проверки авторизации
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		regionID := session.Get("region_id")

		if regionID == nil {
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

		// Добавляем информацию о регионе в контекст
		c.Set("region_id", regionID)
		c.Set("region_name", session.Get("region_name"))
		c.Set("region_login", session.Get("region_login"))

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

// Инициализация регионов
func InitializeRegions(db *gorm.DB) error {
	authService := NewAuthService(db)

	regions := []struct {
		Name     string
		Login    string
		Password string
	}{
		{"Московская область", "moscow", "moscow123"},
		{"Санкт-Петербург", "spb", "spb456"},
		{"Краснодарский край", "krasnodar", "krd789"},
		{"Республика Татарстан", "kazan", "kazan321"},
	}

	for _, r := range regions {
		var existing Region
		if err := db.Where("login = ?", r.Login).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				_, err := authService.CreateRegion(r.Name, r.Login, r.Password)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
