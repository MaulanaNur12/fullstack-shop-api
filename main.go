package main

import (
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB

// Struktur untuk model pengguna
type User struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Struktur Produk
type Produk struct {
	ID        uint    `json:"id" gorm:"primaryKey"`
	Nama      string  `json:"nama"`
	Deskripsi string  `json:"deskripsi"`
	Harga     float64 `json:"harga"`
	URLGambar string  `json:"url_gambar"`
}

// Inisialisasi database
func inisialisasiDB() {
	var err error
	dsn := "root:@tcp(127.0.0.1:3306)/toko?charset=utf8mb4&parseTime=True&loc=Local"
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Gagal menghubungkan ke database: %v", err)
	}
	log.Println("Berhasil terhubung ke database")

	// Migrasi untuk produk dan pengguna
	if err := db.AutoMigrate(&Produk{}, &User{}); err != nil {
		log.Fatalf("Gagal melakukan migrasi database: %v", err)
	}
	log.Println("Migrasi database berhasil")
}

// Fungsi untuk pendaftaran pengguna
func registerUser(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Input tidak valid"})
		return
	}

	// Hash password
	hashedPassword, err := hashPassword(user.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengenkripsi kata sandi"})
		return
	}
	user.Password = hashedPassword

	if err := db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan pengguna"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Pendaftaran berhasil"})
}

// Fungsi untuk login pengguna
func loginUser(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Input tidak valid"})
		return
	}

	var user User
	if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Pengguna tidak ditemukan"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Kata sandi salah"})
		return
	}

	// Buat token (ganti dengan logika token yang sesuai)
	token := "some-jwt-token"
	c.JSON(http.StatusOK, gin.H{"token": token, "message": "Login berhasil"})
}

// Fungsi untuk mengenkripsi password
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// Mendapatkan daftar produk
func dapatkanProduk(c *gin.Context) {
	var produk []Produk
	if hasil := db.Find(&produk); hasil.Error != nil {
		c.JSON(500, gin.H{"error": "Gagal mendapatkan produk"})
		return
	}
	c.JSON(200, produk)
}

// Mendapatkan produk berdasarkan ID
func dapatkanProdukBerdasarkanID(c *gin.Context) {
	id := c.Param("id")
	var produk Produk
	if hasil := db.First(&produk, id); hasil.Error != nil {
		c.JSON(404, gin.H{"error": "Produk tidak ditemukan"})
		return
	}
	c.JSON(200, produk)
}

// Membuat produk baru
func buatProduk(c *gin.Context) {
	var produk Produk
	if err := c.ShouldBindJSON(&produk); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if hasil := db.Create(&produk); hasil.Error != nil {
		c.JSON(500, gin.H{"error": "Gagal membuat produk"})
		return
	}
	c.JSON(201, produk)
}

// Memperbarui produk
func perbaruiProduk(c *gin.Context) {
	id := c.Param("id")
	var produk Produk
	if hasil := db.First(&produk, id); hasil.Error != nil {
		c.JSON(404, gin.H{"error": "Produk tidak ditemukan"})
		return
	}

	if err := c.ShouldBindJSON(&produk); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if hasil := db.Save(&produk); hasil.Error != nil {
		c.JSON(500, gin.H{"error": "Gagal memperbarui produk"})
		return
	}
	c.JSON(200, produk)
}

// Menghapus produk
func hapusProduk(c *gin.Context) {
	id := c.Param("id")
	if hasil := db.Delete(&Produk{}, id); hasil.Error != nil {
		c.JSON(404, gin.H{"error": "Produk tidak ditemukan"})
		return
	}
	c.Status(204)
}

func main() {
	inisialisasiDB()

	r := gin.Default()
	r.Use(cors.Default())

	// Endpoint untuk pengguna
	r.POST("/register", registerUser)
	r.POST("/login", loginUser)

	// Endpoint untuk produk
	r.GET("/produk", dapatkanProduk)
	r.GET("/produk/:id", dapatkanProdukBerdasarkanID)
	r.POST("/produk", buatProduk)
	r.PUT("/produk/:id", perbaruiProduk)
	r.DELETE("/produk/:id", hapusProduk)

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}
