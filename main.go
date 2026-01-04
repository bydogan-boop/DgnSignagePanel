package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

var DB *sql.DB

// --- Veri Yapıları (Structs) ---

type Restaurant struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	Address   string    `json:"address"`
	LogoURL   string    `json:"logo_url"`
	ScreenCount int       `json:"screen_count"`
	CreatedAt time.Time `json:"created_at"`
}

type Screen struct {
	ID           int          `json:"id"`
	ScreenCode   string       `json:"screen_code"`
	RestaurantID int          `json:"restaurant_id"`
	RestaurantName string       `json:"restaurant_name"`
	ContentType  string       `json:"content_type"`
	MediaURL     string       `json:"media_url"`
	VersionHash  string       `json:"version_hash"`
	LastSeenAt   sql.NullTime `json:"last_seen_at"`
	IsOnline     bool         `json:"is_online"`
}

type PlayerResponse struct {
	MediaURL    string `json:"media_url"`
	ContentType string `json:"content_type"`
	VersionHash string `json:"version_hash"`
}

// --- Ana Fonksiyon ---

func main() {
	godotenv.Load()
	initDB()
	defer DB.Close()

	router := mux.NewRouter()

	// CORS Ayarları
	corsHandler := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	// API Rotaları (Frontend'in beklediği tüm rotalar eklendi)
	router.HandleFunc("/api/v1/player/{screen_code}", getPlayerConfig).Methods("GET")
	router.HandleFunc("/api/restaurants", getRestaurants).Methods("GET")
	router.HandleFunc("/api/restaurants/status", getRestaurants).Methods("GET")
	router.HandleFunc("/api/restaurants/{id:[0-9]+}", getRestaurantDetail).Methods("GET")
	router.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))
	router.HandleFunc("/api/screens", createScreen).Methods("POST")
	router.HandleFunc("/api/screens/{id}", deleteScreen).Methods("DELETE")
	router.HandleFunc("/api/screens/{id}", updateScreen).Methods("POST")
	router.HandleFunc("/api/restaurants", createRestaurant).Methods("POST")
	router.HandleFunc("/api/restaurants/{id}", updateRestaurant).Methods("PUT") // Veya POST
	router.HandleFunc("/api/restaurants/{id}", deleteRestaurant).Methods("DELETE")
	router.HandleFunc("/api/screens", getScreens).Methods("GET")
	router.HandleFunc("/api/dashboard", getDashboard).Methods("GET")

	log.Println("✅ Backend :8081 portunda çalışıyor...")
	log.Fatal(http.ListenAndServe(":8081", corsHandler(router)))
}

// --- Fonksiyonlar (Handlers) ---

func getRestaurants(w http.ResponseWriter, r *http.Request) {
    // SQL sorgusunda LEFT JOIN ve GROUP BY kullanarak ekran sayılarını hesaplıyoruz
    query := `
        SELECT r.id, r.name, r.phone, r.address, r.logo_url, r.created_at, 
               COUNT(s.id) as screen_count
        FROM restaurants r
        LEFT JOIN screens s ON r.id = s.restaurant_id
        GROUP BY r.id`

    rows, err := DB.Query(query)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var results []Restaurant
    for rows.Next() {
        var res Restaurant
        // Scan sırasına dikkat: logo_url ve screen_count eklendi
        err := rows.Scan(&res.ID, &res.Name, &res.Phone, &res.Address, 
                         &res.LogoURL, &res.CreatedAt, &res.ScreenCount)
        if err != nil {
            continue
        }
        results = append(results, res)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(results)
}

func getRestaurantDetail(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]

    // 1. Restoran bilgisini çek
    var res Restaurant
    err := DB.QueryRow("SELECT id, name, phone, address, logo_url, created_at FROM restaurants WHERE id = ?", id).
        Scan(&res.ID, &res.Name, &res.Phone, &res.Address, &res.LogoURL, &res.CreatedAt)

    if err != nil {
        http.Error(w, "Restoran bulunamadı", http.StatusNotFound)
        return
    }

    // 2. Bu restorana ait ekranları çek
    rows, _ := DB.Query("SELECT id, screen_code, content_type, media_url, is_online FROM screens WHERE restaurant_id = ?", id)
    defer rows.Close()

    var screens []Screen
    for rows.Next() {
        var s Screen
        rows.Scan(&s.ID, &s.ScreenCode, &s.ContentType, &s.MediaURL, &s.IsOnline)
        screens = append(screens, s)
    }

    // 3. Frontend'in beklediği İÇ İÇE (nested) yapıyı oluştur
    response := map[string]interface{}{
        "restaurant": res,
        "screens":    screens,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func getScreens(w http.ResponseWriter, r *http.Request) {
    // Sorguda 'LEFT JOIN' kullanarak restoran adını (r.name) çekiyoruz
    query := `
        SELECT s.id, s.restaurant_id, r.name as restaurant_name, s.screen_code, 
               s.content_type, s.media_url, s.is_online, s.last_seen_at 
        FROM screens s
        LEFT JOIN restaurants r ON s.restaurant_id = r.id`
    
    rows, err := DB.Query(query)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var screens []Screen
    for rows.Next() {
        var s Screen
        // Scan fonksiyonuna s.RestaurantName'i de eklemeyi unutma
        err := rows.Scan(&s.ID, &s.RestaurantID, &s.RestaurantName, &s.ScreenCode, 
                         &s.ContentType, &s.MediaURL, &s.IsOnline, &s.LastSeenAt)
        if err != nil {
            log.Printf("Scan hatası: %v", err)
            continue
        }
        screens = append(screens, s)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(screens)
}

func createScreen(w http.ResponseWriter, r *http.Request) {
    err := r.ParseMultipartForm(10 << 20)
    if err != nil {
        http.Error(w, "Dosya çok büyük", http.StatusBadRequest)
        return
    }

    resID := r.FormValue("restaurant_id")
    screenCode := r.FormValue("screen_code")
    contentType := r.FormValue("content_type")

    // 1. 32 Karakterlik MD5 Hash Üret (Orijinal yapıya uygun)
    now := time.Now().UnixNano()
    hash := md5.Sum([]byte(fmt.Sprintf("%d", now)))
    versionHash := fmt.Sprintf("%x", hash) // Tam 32 karakter üretir

    file, handler, err := r.FormFile("media_file")
    var mediaURL string
    if err == nil {
        defer file.Close()
        
        // 2. 19 Karakterlik Hassas Zaman Damgası (UnixNano)
        // Bu, dosya isimlerinin çakışmasını önler ve orijinal yapıya uyar
        fileName := fmt.Sprintf("%d-%s", now, handler.Filename)
        
        os.MkdirAll("./uploads", os.ModePerm)
        filePath := filepath.Join("./uploads", fileName)
        dst, _ := os.Create(filePath)
        defer dst.Close()
        io.Copy(dst, file)
        
        mediaURL = "/uploads/" + fileName
    }

    query := `INSERT INTO screens (restaurant_id, screen_code, content_type, media_url, is_online, version_hash, created_at) 
              VALUES (?, ?, ?, ?, true, ?, NOW())`
    
    _, err = DB.Exec(query, resID, screenCode, contentType, mediaURL, versionHash)
    if err != nil {
        http.Error(w, "Veritabanı hatası", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{"version_hash": versionHash})
}

func deleteScreen(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]

    _, err := DB.Exec("DELETE FROM screens WHERE id = ?", id)
    if err != nil {
        http.Error(w, "Silme hatası: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    fmt.Fprint(w, `{"message": "Ekran silindi"}`)
}

func updateScreen(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]

    err := r.ParseMultipartForm(10 << 20)
    if err != nil {
        http.Error(w, "Dosya çok büyük", http.StatusBadRequest)
        return
    }

    resID := r.FormValue("restaurant_id")
    screenCode := r.FormValue("screen_code")
    contentType := r.FormValue("content_type")
    
    // 1. Yeni Version Hash Üret (32 Karakter MD5)
    now := time.Now().UnixNano()
    hash := md5.Sum([]byte(fmt.Sprintf("%d", now)))
    versionHash := fmt.Sprintf("%x", hash)

    // 2. Eğer yeni dosya seçilmişse onu işle
    var mediaUpdateQuery string
    var mediaUpdateParam interface{}
    
    file, handler, err := r.FormFile("media_file")
    if err == nil {
        defer file.Close()
        // 19 haneli zaman damgasıyla isimlendirme
        fileName := fmt.Sprintf("%d-%s", now, handler.Filename)
        filePath := filepath.Join("./uploads", fileName)
        
        dst, _ := os.Create(filePath)
        defer dst.Close()
        io.Copy(dst, file)
        
        mediaUpdateQuery = ", media_url = ?"
        mediaUpdateParam = "/uploads/" + fileName
    }

    // 3. Veritabanını güncelle
    sqlQuery := fmt.Sprintf("UPDATE screens SET restaurant_id = ?, screen_code = ?, content_type = ?, version_hash = ? %s WHERE id = ?", mediaUpdateQuery)
    
    var args []interface{}
    args = append(args, resID, screenCode, contentType, versionHash)
    if mediaUpdateParam != nil {
        args = append(args, mediaUpdateParam)
    }
    args = append(args, id)

    _, err = DB.Exec(sqlQuery, args...)
    if err != nil {
        http.Error(w, "Güncelleme hatası: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    fmt.Fprint(w, `{"message": "Ekran güncellendi", "version_hash": "`+versionHash+`"}`)
}

// Yeni Restoran Ekle
func createRestaurant(w http.ResponseWriter, r *http.Request) {
    r.ParseMultipartForm(10 << 20)

    name := r.FormValue("name")
    phone := r.FormValue("phone")
    address := r.FormValue("address")

    if name == "" {
        http.Error(w, "Restoran adı zorunludur", http.StatusBadRequest)
        return
    }

    var logoURL string
    file, handler, err := r.FormFile("logo")
    // Eğer dosya gönderilmişse (hata yoksa) kaydet
    if err == nil {
        defer file.Close()
        fileName := fmt.Sprintf("%d-%s", time.Now().UnixNano(), handler.Filename)
        filePath := filepath.Join("./uploads", fileName)
        
        os.MkdirAll("./uploads", os.ModePerm)
        dst, _ := os.Create(filePath)
        defer dst.Close()
        io.Copy(dst, file)
        logoURL = "/uploads/" + fileName
    }

    // Logo yoksa logoURL boş string kalacak ve DB'ye öyle yazılacak
    query := "INSERT INTO restaurants (name, phone, address, logo_url, created_at) VALUES (?, ?, ?, ?, NOW())"
    _, err = DB.Exec(query, name, phone, address, logoURL)
    if err != nil {
        http.Error(w, "DB Hatası: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    fmt.Fprint(w, `{"message": "Başarıyla oluşturuldu"}`)
}

// Restoran Düzenle
func updateRestaurant(w http.ResponseWriter, r *http.Request) {
    // 1. URL'den ID'yi al
    vars := mux.Vars(r)
    id := vars["id"]

    // 2. Formu ayrıştır
    r.ParseMultipartForm(10 << 20)

    name := r.FormValue("name")
    phone := r.FormValue("phone")
    address := r.FormValue("address")

    var logoUpdateQuery string
    var args []interface{}
    args = append(args, name, phone, address)

    // 3. Logo kontrolü (Eğer yeni logo seçilmişse)
    file, handler, err := r.FormFile("logo")
    if err == nil {
        defer file.Close()
        fileName := fmt.Sprintf("%d-%s", time.Now().UnixNano(), handler.Filename)
        filePath := filepath.Join("./uploads", fileName)
        
        dst, _ := os.Create(filePath)
        defer dst.Close()
        io.Copy(dst, file)
        
        logoUpdateQuery = ", logo_url = ?"
        args = append(args, "/uploads/"+fileName)
    }

    // 4. SQL Sorgusunu oluştur ve çalıştır
    args = append(args, id)
    sqlQuery := fmt.Sprintf("UPDATE restaurants SET name = ?, phone = ?, address = ? %s WHERE id = ?", logoUpdateQuery)
    
    _, err = DB.Exec(sqlQuery, args...)
    if err != nil {
        log.Printf("Düzenleme Hatası: %v", err)
        http.Error(w, "Veritabanı hatası", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    fmt.Fprint(w, `{"message": "Restoran başarıyla güncellendi"}`)
}

// Restoran Sil
func deleteRestaurant(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]

    // ÖNEMLİ: Restoran silinince ona bağlı ekranlar hata verebilir. 
    // Önce ekranları silmek veya bağlantıyı koparmak gerekebilir.
    _, err := DB.Exec("DELETE FROM restaurants WHERE id = ?", id)
    if err != nil {
        http.Error(w, "Silme hatası: "+err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(map[string]string{"message": "Restoran silindi"})
}

func getDashboard(w http.ResponseWriter, r *http.Request) {
    // 1. Genel Sayıları Al
    var resCount, screenCount, activeCount int
    DB.QueryRow("SELECT COUNT(*) FROM restaurants").Scan(&resCount)
    DB.QueryRow("SELECT COUNT(*) FROM screens").Scan(&screenCount)
    DB.QueryRow("SELECT COUNT(*) FROM screens WHERE is_online = true").Scan(&activeCount)

    // 2. Restoran Durumlarını Al
    queryStatus := `
        SELECT r.id, r.name, COUNT(s.id) as screen_count, 
               COALESCE(SUM(CASE WHEN s.is_online = true THEN 1 ELSE 0 END) > 0, false) as is_any_online
        FROM restaurants r
        LEFT JOIN screens s ON r.id = s.restaurant_id
        GROUP BY r.id`
    
    rowsStatus, _ := DB.Query(queryStatus)
    var restaurantStatus []map[string]interface{}
    for rowsStatus.Next() {
        var id, count int
        var name string
        var isOnline bool
        rowsStatus.Scan(&id, &name, &count, &isOnline)
        restaurantStatus = append(restaurantStatus, map[string]interface{}{
            "id":                   id,
            "name":                 name,
            "screen_count":         count,
            "is_any_screen_online": isOnline,
        })
    }
    rowsStatus.Close()

    // 3. SON AKTİF EKRANLARI AL (Eksik olan kısım burasıydı)
    queryRecent := `
        SELECT s.screen_code, r.name as restaurant_name, s.last_seen_at
        FROM screens s
        JOIN restaurants r ON s.restaurant_id = r.id
        WHERE s.last_seen_at IS NOT NULL
        ORDER BY s.last_seen_at DESC
        LIMIT 5`
    
    rowsRecent, _ := DB.Query(queryRecent)
    var recentScreens []map[string]interface{}
    for rowsRecent.Next() {
        var code, resName string
        var lastSeen sql.NullTime
        rowsRecent.Scan(&code, &resName, &lastSeen)
        recentScreens = append(recentScreens, map[string]interface{}{
            "screen_code":     code,
            "restaurant_name": resName,
            "last_seen_at":    lastSeen,
        })
    }
    rowsRecent.Close()

    // 4. Veriyi Paketle
    data := map[string]interface{}{
        "restaurant_count":  resCount,
        "screen_count":      screenCount,
        "active_screens":    activeCount,
        "restaurant_status": restaurantStatus,
        "recent_screens":    recentScreens, // Frontend burayı bekliyor
        "status":            "Sistem Aktif",
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(data)
}

func getPlayerConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	screenCode := vars["screen_code"]

	var screen Screen
	err := DB.QueryRow(
		"SELECT id, screen_code, restaurant_id, content_type, media_url, version_hash FROM screens WHERE screen_code = ?",
		screenCode,
	).Scan(&screen.ID, &screen.ScreenCode, &screen.RestaurantID, &screen.ContentType, &screen.MediaURL, &screen.VersionHash)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Ekran bulunamadı", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}


	DB.Exec("UPDATE screens SET last_seen_at = ?, is_online = ? WHERE id = ?", time.Now(), true, screen.ID)

	response := PlayerResponse{
		MediaURL:    screen.MediaURL,
		ContentType: screen.ContentType,
		VersionHash: screen.VersionHash,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func initDB() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", 
		os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"))

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil || DB.Ping() != nil {
		log.Fatalf("❌ Veritabanı hatası: %v", err)
	}
	log.Println("✅ Veritabanı bağlantısı başarılı.")
}