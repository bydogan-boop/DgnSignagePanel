
package main

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "time"

    _ "github.com/go-sql-driver/mysql"
    "github.com/gorilla/mux"
)

// DB değişkeni veritabanı bağlantısını tutar.
var DB *sql.DB

// Restaurant veritabanındaki `restaurants` tablosunu temsil eder.
type Restaurant struct {
    ID        int       `json:"id"`
    Name      string    `json:"name"`
    Phone     string    `json:"phone"`
    Address   string    `json:"address"`
    CreatedAt time.Time `json:"created_at"`
}

// Screen veritabanındaki `screens` tablosunu temsil eder.
type Screen struct {
    ID           int          `json:"id"`
    ScreenCode   string       `json:"screen_code"`
    RestaurantID int          `json:"restaurant_id"`
    ContentType  string       `json:"content_type"`
    MediaURL     string       `json:"media_url"`
    VersionHash  string       `json:"version_hash"`
    LastSeenAt   sql.NullTime `json:"last_seen_at"`
    IsOnline     bool         `json:"is_online"`
}

// PlayerResponse, Android cihaza gönderilecek JSON yanıtını temsil eder.
type PlayerResponse struct {
    MediaURL    string `json:"media_url"`
    ContentType string `json:"content_type"`
    VersionHash string `json:"version_hash"`
}

func main() {
    // Veritabanı bağlantısını başlat.
    initDB()
    defer DB.Close()

    // Yeni bir router oluştur.
    router := mux.NewRouter()

    // API endpoint'ini tanımla.
    router.HandleFunc("/api/v1/player/{screen_code}", getPlayerConfig).Methods("GET")

    // Sunucuyu başlat.
    log.Println("Sunucu :8080 portunda başlatıldı")
    log.Fatal(http.ListenAndServe(":8080", router))
}

func initDB() {
    // Veritabanı bağlantı bilgilerini ortam değişkenlerinden al.
    dbUser := os.Getenv("DB_USER")
    dbPassword := os.Getenv("DB_PASSWORD")
    dbHost := os.Getenv("DB_HOST")
    dbName := os.Getenv("DB_NAME")

    // Veritabanı bağlantı cümlesini oluştur.
    // Örnek: "user:password@tcp(127.0.0.1:3306)/database_name"
    dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", dbUser, dbPassword, dbHost, dbName)

    var err error
    DB, err = sql.Open("mysql", dsn)
    if err != nil {
        log.Fatalf("Veritabanı bağlantısı oluşturulamadı: %v", err)
    }

    // Bağlantıyı kontrol et.
    err = DB.Ping()
    if err != nil {
        log.Fatalf("Veritabanına bağlanılamadı: %v", err)
    }

    log.Println("Veritabanı bağlantısı başarılı.")
}

func getPlayerConfig(w http.ResponseWriter, r *http.Request) {
    // URL'den screen_code parametresini al.
    vars := mux.Vars(r)
    screenCode := vars["screen_code"]

    // Veritabanından ilgili ekranı sorgula.
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

    // Cihazın son görülme zamanını ve online durumunu güncelle.
    _, err = DB.Exec("UPDATE screens SET last_seen_at = ?, is_online = ? WHERE id = ?", time.Now(), true, screen.ID)
    if err != nil {
        // Bu hatayı loglamak önemli, ama client'a 500 hatası dönmek yerine devam edebiliriz.
        log.Printf("last_seen_at güncellenirken hata oluştu: %v", err)
    }

    // Android cihaz için yanıtı oluştur.
    response := PlayerResponse{
        MediaURL:    screen.MediaURL,
        ContentType: screen.ContentType,
        VersionHash: screen.VersionHash,
    }

    // Yanıtı JSON formatında gönder.
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
