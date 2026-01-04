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
	"path/filepath"
	//"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

// --- Veri Yapıları (Structs) ---

var DB *sql.DB

type Restaurant struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Phone       string    `json:"phone"`
	Address     string    `json:"address"`
	LogoURL     string    `json:"logo_url"`
	ScreenCount int       `json:"screen_count"`
	CreatedAt   time.Time `json:"created_at"`
}

type Screen struct {
	ID             int          `json:"id"`
	ScreenCode     string       `json:"screen_code"`
	RestaurantID   int          `json:"restaurant_id"`
	RestaurantName string       `json:"restaurant_name"`
	ContentType    string       `json:"content_type"`
	MediaURL       string       `json:"media_url"`
	VersionHash    string       `json:"version_hash"`
	LastSeenAt     sql.NullTime `json:"last_seen_at"`
	IsOnline       bool         `json:"is_online"`
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

	// Arka plan işlerini başlat
	startStatusChecker()
	log.Println("Sinyal bekçisi arka planda başlatıldı...")

	router := mux.NewRouter()

	// --- Rotalar (Routes) ---

	// Auth gerektirmeyen rotalar
	router.HandleFunc("/api/login", loginHandler).Methods("POST")
	router.HandleFunc("/api/heartbeat", screenHeartbeat).Methods("POST")
	router.HandleFunc("/api/v1/player/{screen_code}", getPlayerConfig).Methods("GET")
	router.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	// Auth gerektiren API rotaları
	api := router.PathPrefix("/api").Subrouter()
	api.Use(authMiddleware)

	// Restoran Yönetimi
	api.HandleFunc("/restaurants", getRestaurants).Methods("GET")
	api.HandleFunc("/restaurants/status", getRestaurants).Methods("GET")
	api.HandleFunc("/restaurants/{id:[0-9]+}", getRestaurantDetail).Methods("GET")
	api.HandleFunc("/restaurants", createRestaurant).Methods("POST")
	api.HandleFunc("/restaurants/{id}", updateRestaurant).Methods("POST", "PUT")
	api.HandleFunc("/restaurants/{id}", deleteRestaurant).Methods("DELETE")

	// Ekran Yönetimi
	api.HandleFunc("/screens", getScreens).Methods("GET")
	api.HandleFunc("/screens", createScreen).Methods("POST")
	api.HandleFunc("/screens/{id}", updateScreen).Methods("POST")
	api.HandleFunc("/screens/{id}", deleteScreen).Methods("DELETE")

	// Dashboard
	api.HandleFunc("/dashboard", getDashboard).Methods("GET")

	// Sunucuyu CORS ve Auth ile sarmalayarak başlat
	log.Println("Sunucu 8081 portunda baslatildi...")
	http.ListenAndServe(":8081", corsHandler(router))
}

// --- Middleware & Güvenlik ---

func corsHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token != "1211" {
			http.Error(w, "Yetkisiz erişim!", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var creds struct{ Password string `json:"password"` }
	json.NewDecoder(r.Body).Decode(&creds)
	if creds.Password == "1211" {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"token": "1211"}`)
	} else {
		http.Error(w, "Hatalı Şifre", http.StatusUnauthorized)
	}
}

// --- API İşleyicileri (Handlers) ---

func getRestaurants(w http.ResponseWriter, r *http.Request) {
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
		err := rows.Scan(&res.ID, &res.Name, &res.Phone, &res.Address,
			&res.LogoURL, &res.CreatedAt, &res.ScreenCount)
		if err == nil {
			results = append(results, res)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func getRestaurantDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var res Restaurant
	err := DB.QueryRow("SELECT id, name, phone, address, logo_url, created_at FROM restaurants WHERE id = ?", id).
		Scan(&res.ID, &res.Name, &res.Phone, &res.Address, &res.LogoURL, &res.CreatedAt)

	if err != nil {
		http.Error(w, "Restoran bulunamadı", http.StatusNotFound)
		return
	}

	rows, _ := DB.Query("SELECT id, screen_code, content_type, media_url, is_online FROM screens WHERE restaurant_id = ?", id)
	defer rows.Close()

	var screens []Screen
	for rows.Next() {
		var s Screen
		rows.Scan(&s.ID, &s.ScreenCode, &s.ContentType, &s.MediaURL, &s.IsOnline)
		screens = append(screens, s)
	}

	response := map[string]interface{}{"restaurant": res, "screens": screens}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getScreens(w http.ResponseWriter, r *http.Request) {
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
		err := rows.Scan(&s.ID, &s.RestaurantID, &s.RestaurantName, &s.ScreenCode,
			&s.ContentType, &s.MediaURL, &s.IsOnline, &s.LastSeenAt)
		if err == nil {
			screens = append(screens, s)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(screens)
}

func createScreen(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)
	resID := r.FormValue("restaurant_id")
	screenCode := r.FormValue("screen_code")
	contentType := r.FormValue("content_type")

	now := time.Now().UnixNano()
	versionHash := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%d", now))))

	var mediaURL string
	file, handler, err := r.FormFile("media_file")
	if err == nil {
		defer file.Close()
		fileName := fmt.Sprintf("%d-%s", now, handler.Filename)
		filePath := filepath.Join("./uploads", fileName)
		os.MkdirAll("./uploads", os.ModePerm)
		dst, _ := os.Create(filePath)
		defer dst.Close()
		io.Copy(dst, file)
		mediaURL = "/uploads/" + fileName
	}

	query := `INSERT INTO screens (restaurant_id, screen_code, content_type, media_url, is_online, version_hash, created_at) 
              VALUES (?, ?, ?, ?, true, ?, NOW())`
	_, err = DB.Exec(query, resID, screenCode, contentType, mediaURL, versionHash)
	if err != nil {
		http.Error(w, "DB Hatası", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"version_hash": versionHash})
}

func updateScreen(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	r.ParseMultipartForm(10 << 20)

	resID := r.FormValue("restaurant_id")
	screenCode := r.FormValue("screen_code")
	contentType := r.FormValue("content_type")
	now := time.Now().UnixNano()
	versionHash := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%d", now))))

	var mediaUpdateQuery string
	var args []interface{}
	args = append(args, resID, screenCode, contentType, versionHash)

	file, handler, err := r.FormFile("media_file")
	if err == nil {
		defer file.Close()
		fileName := fmt.Sprintf("%d-%s", now, handler.Filename)
		filePath := filepath.Join("./uploads", fileName)
		dst, _ := os.Create(filePath)
		defer dst.Close()
		io.Copy(dst, file)
		mediaUpdateQuery = ", media_url = ?"
		args = append(args, "/uploads/"+fileName)
	}

	args = append(args, id)
	sqlQuery := fmt.Sprintf("UPDATE screens SET restaurant_id = ?, screen_code = ?, content_type = ?, version_hash = ? %s WHERE id = ?", mediaUpdateQuery)
	_, err = DB.Exec(sqlQuery, args...)
	if err != nil {
		http.Error(w, "Güncelleme hatası", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"message": "Ekran güncellendi"}`)
}

func deleteScreen(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	DB.Exec("DELETE FROM screens WHERE id = ?", id)
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"message": "Ekran silindi"}`)
}

func createRestaurant(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)
	name := r.FormValue("name")
	phone := r.FormValue("phone")
	address := r.FormValue("address")

	var logoURL string
	file, handler, err := r.FormFile("logo")
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

	_, err = DB.Exec("INSERT INTO restaurants (name, phone, address, logo_url, created_at) VALUES (?, ?, ?, ?, NOW())", name, phone, address, logoURL)
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, `{"message": "Başarıyla oluşturuldu"}`)
}

func updateRestaurant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	r.ParseMultipartForm(10 << 20)

	name := r.FormValue("name")
	phone := r.FormValue("phone")
	address := r.FormValue("address")

	var logoUpdateQuery string
	var args []interface{}
	args = append(args, name, phone, address)

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

	args = append(args, id)
	sqlQuery := fmt.Sprintf("UPDATE restaurants SET name = ?, phone = ?, address = ? %s WHERE id = ?", logoUpdateQuery)
	DB.Exec(sqlQuery, args...)
	w.WriteHeader(http.StatusOK)
}

func deleteRestaurant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	DB.Exec("DELETE FROM screens WHERE restaurant_id = ?", id)
	DB.Exec("DELETE FROM restaurants WHERE id = ?", id)
	w.WriteHeader(http.StatusOK)
}

func screenHeartbeat(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ScreenCode  string `json:"screen_code"`
		CurrentHash string `json:"current_hash"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	var dbHash, mediaURL, contentType string
	var screenID int
	err := DB.QueryRow("SELECT id, version_hash, media_url, content_type FROM screens WHERE screen_code = ?", req.ScreenCode).Scan(&screenID, &dbHash, &mediaURL, &contentType)

	if err == nil {
		DB.Exec("UPDATE screens SET is_online = true, last_seen_at = NOW() WHERE id = ?", screenID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":       "ok",
			"needs_update": req.CurrentHash != dbHash,
			"new_hash":     dbHash,
			"media_url":    mediaURL,
			"content_type": contentType,
		})
	} else {
		http.Error(w, "Ekran bulunamadı", http.StatusNotFound)
	}
}

func startStatusChecker() {
	go func() {
		for {
			DB.Exec(`UPDATE screens SET is_online = false WHERE last_seen_at < DATE_SUB(NOW(), INTERVAL 2 MINUTE) AND is_online = true`)
			time.Sleep(1 * time.Minute)
		}
	}()
}

func getDashboard(w http.ResponseWriter, r *http.Request) {
	var resCount, screenCount, activeCount int
	DB.QueryRow("SELECT COUNT(*) FROM restaurants").Scan(&resCount)
	DB.QueryRow("SELECT COUNT(*) FROM screens").Scan(&screenCount)
	DB.QueryRow("SELECT COUNT(*) FROM screens WHERE is_online = true").Scan(&activeCount)

	queryStatus := `SELECT r.id, r.name, COUNT(s.id), COALESCE(SUM(CASE WHEN s.is_online = true THEN 1 ELSE 0 END) > 0, false)
                    FROM restaurants r LEFT JOIN screens s ON r.id = s.restaurant_id GROUP BY r.id`
	rows, _ := DB.Query(queryStatus)
	var restaurantStatus []map[string]interface{}
	for rows.Next() {
		var id, count int
		var name string
		var isOnline bool
		rows.Scan(&id, &name, &count, &isOnline)
		restaurantStatus = append(restaurantStatus, map[string]interface{}{"id": id, "name": name, "screen_count": count, "is_any_screen_online": isOnline})
	}
	rows.Close()

	queryRecent := `SELECT s.screen_code, r.name, s.last_seen_at FROM screens s JOIN restaurants r ON s.restaurant_id = r.id 
                    WHERE s.last_seen_at IS NOT NULL ORDER BY s.last_seen_at DESC LIMIT 5`
	rowsR, _ := DB.Query(queryRecent)
	var recentScreens []map[string]interface{}
	for rowsR.Next() {
		var code, resName string
		var lastSeen sql.NullTime
		rowsR.Scan(&code, &resName, &lastSeen)
		recentScreens = append(recentScreens, map[string]interface{}{"screen_code": code, "restaurant_name": resName, "last_seen_at": lastSeen})
	}
	rowsR.Close()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"restaurant_count": resCount, "screen_count": screenCount, "active_screens": activeCount,
		"restaurant_status": restaurantStatus, "recent_screens": recentScreens,
	})
}

func getPlayerConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var screen Screen
	err := DB.QueryRow("SELECT id, content_type, media_url, version_hash FROM screens WHERE screen_code = ?", vars["screen_code"]).
		Scan(&screen.ID, &screen.ContentType, &screen.MediaURL, &screen.VersionHash)

	if err == nil {
		DB.Exec("UPDATE screens SET last_seen_at = NOW(), is_online = true WHERE id = ?", screen.ID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(PlayerResponse{MediaURL: screen.MediaURL, ContentType: screen.ContentType, VersionHash: screen.VersionHash})
	} else {
		http.Error(w, "Ekran bulunamadı", http.StatusNotFound)
	}
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