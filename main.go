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
    "path/filepath" // Dosya işlemleri için gerekli, kalsın.
    "strings"
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
    ID             int            `json:"id"`
    ScreenCode     string         `json:"screen_code"`
    RestaurantID   int            `json:"restaurant_id"`
    RestaurantName string         `json:"restaurant_name"`
    ContentType    string         `json:"content_type"`
    MediaURL       string         `json:"media_url"`
    VersionHash    string         `json:"version_hash"`
    LastSeenAt     sql.NullTime   `json:"last_seen_at"` // Null güvenli
    IsOnline       bool           `json:"is_online"`
    TickerText     sql.NullString `json:"ticker_text"`     // Null güvenli
    TickerStartAt  sql.NullTime   `json:"ticker_start_at"`  // Null güvenli
    TickerEndAt    sql.NullTime   `json:"ticker_end_at"`    // Null güvenli
    TickerActive   bool           `json:"ticker_active"`
}

type PlayerResponse struct {
    MediaURL    string `json:"media_url"`
    ContentType string `json:"content_type"`
    VersionHash string `json:"version_hash"`
    TickerText  string `json:"ticker_text"` // Android'in beklediği yeni alan
}

// --- Ana Fonksiyon ---

func main() {
    godotenv.Load()
    initDB()
    defer DB.Close()

    // Arka plan işlerini başlat
    startStatusChecker()
    startCleanupTask() 
    log.Println("Sinyal bekçisi ve temizlik görevlisi arka planda başlatıldı...")

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

func initDB() {
    // .env dosyasından bilgileri çekiyoruz
    user := os.Getenv("DB_USER")
    pass := os.Getenv("DB_PASSWORD")
    host := os.Getenv("DB_HOST")
    port := os.Getenv("DB_PORT")
    name := os.Getenv("DB_NAME")

    // Kritik nokta: Sona eklenen ?parseTime=true
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, pass, host, port, name)

    var err error
    DB, err = sql.Open("mysql", dsn)
    if err != nil {
        log.Fatalf("❌ Veritabanı sürücüsü başlatılamadı: %v", err)
    }

    // Bağlantıyı test et
    err = DB.Ping()
    if err != nil {
        log.Fatalf("❌ Veritabanına bağlanılamadı! Ayarlarını kontrol et: %v", err)
    }
    log.Printf("🔍 Şu an bağlandığım veritabanı: %s", name) // Bunu ekle
    log.Println("✅ Veritabanı bağlantısı başarıyla dış fonksiyonda kuruldu.")
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
		// Başlığı al
		authHeader := r.Header.Get("Authorization")
		
		// Log: Terminalde ne gördüğümüzü kesinleştirelim
		fmt.Printf("--- Yeni İstek Geldi ---\n")
		fmt.Printf("Gelen Header: %s\n", authHeader)

		// Bearer 1211 formatındaki string'den sadece 1211 kısmını al
		token := strings.TrimSpace(strings.Replace(authHeader, "Bearer", "", 1))
		
		// Şifre kontrolü
		// Not: Eğer .env kullanıyorsan os.Getenv("ADMIN_PASSWORD") yazabilirsin
		if token != "1211" {
			fmt.Printf("❌ Reddedildi! Beklenen: 1211, Gelen: [%s]\n", token)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, `{"error": "Yetkisiz erişim"}`)
			return
		}

		fmt.Printf("✅ Onaylandı!\n")
		next.ServeHTTP(w, r)
	})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		log.Println("JSON Hatası:", err)
		http.Error(w, "Geçersiz veri formatı", http.StatusBadRequest)
		return
	}

	var dbPassword string
	err = DB.QueryRow("SELECT password FROM konsol LIMIT 1").Scan(&dbPassword)
	if err != nil {
		log.Println("DB Hatası:", err)
		http.Error(w, "Sistem hatası", http.StatusInternalServerError)
		return
	}

	// Hem girilen şifreyi hem db şifresini temizleyip karşılaştırıyoruz
	if strings.TrimSpace(input.Password) == strings.TrimSpace(dbPassword) {
		log.Println("✅ Giriş Başarılı!")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"token": dbPassword})
	} else {
		log.Printf("❌ Hatalı Şifre Denemesi! Girilen: [%s], Beklenen: [%s]", input.Password, dbPassword)
		http.Error(w, "Şifre hatalı", http.StatusUnauthorized)
	}
}

// --- Diğer Handler Fonksiyonları (getRestaurants, getScreens vb. aynı kalıyor) ---
// Not: Mesajın boyutu sebebiyle alt kısımdaki değişmeyen handler'ları tekrar yazmıyorum, 
// ama yukarıdaki 'strings' içeren kritik kısımların tamamı düzeltildi.

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
    // "s." veya "r." gibi kısaltmaları kaldırıp tablo isimlerini net yazıyoruz
    query := `
        SELECT 
            screens.id, 
            screens.screen_code, 
            screens.restaurant_id, 
            IFNULL(restaurants.name, '') as restaurant_name, 
            screens.content_type, 
            screens.media_url, 
            screens.version_hash, 
            screens.last_seen_at, 
            screens.is_online, 
            screens.ticker_text, 
            screens.ticker_start_at, 
            screens.ticker_end_at, 
            screens.ticker_active
        FROM screens
        LEFT JOIN restaurants ON screens.restaurant_id = restaurants.id`

    rows, err := DB.Query(query)
    if err != nil {
        // Hata buraya düşerse terminalde çok net göreceğiz
        log.Printf("❌ SQL Sorgu Hatası: %v", err)
        http.Error(w, "Veritabanı sorgu hatası", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var screens []Screen
    for rows.Next() {
        var s Screen
        // Scan sırası yukarıdaki SELECT sırasıyla aynı (13 sütun)
        err := rows.Scan(
            &s.ID, &s.ScreenCode, &s.RestaurantID, &s.RestaurantName,
            &s.ContentType, &s.MediaURL, &s.VersionHash, &s.LastSeenAt,
            &s.IsOnline, &s.TickerText, &s.TickerStartAt, &s.TickerEndAt, &s.TickerActive,
        )
        
        if err != nil {
            log.Printf("❌ Satır Okuma (Scan) Hatası: %v", err)
            continue 
        }
        screens = append(screens, s)
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
        // Burada filepath.Join kullanabilmen için import kısmında "path/filepath" olmalı
        filePath := filepath.Join("./uploads", fileName)
        os.MkdirAll("./uploads", os.ModePerm)
        dst, err := os.Create(filePath)
        if err == nil {
            defer dst.Close()
            io.Copy(dst, file)
            mediaURL = "/uploads/" + fileName 
        }
    }

    query := `INSERT INTO screens (restaurant_id, screen_code, content_type, media_url, is_online, version_hash, created_at) 
              VALUES (?, ?, ?, ?, true, ?, NOW())`
    _, err = DB.Exec(query, resID, screenCode, contentType, mediaURL, versionHash)
    if err != nil {
        log.Printf("createScreen DB Hatası: %v", err)
        http.Error(w, "DB Hatası", http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{"version_hash": versionHash})
}

func updateScreen(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]

    // 1. Form verisini oku
    err := r.ParseMultipartForm(10 << 20)
    if err != nil {
        log.Println("Form ayrıştırma hatası:", err)
        http.Error(w, "Dosya çok büyük veya geçersiz", http.StatusBadRequest)
        return
    }

    // Temel bilgiler
    resID := r.FormValue("restaurant_id")
    screenCode := r.FormValue("screen_code")
    contentType := r.FormValue("content_type")
    
    // --- YENİ: Ticker Bilgilerini Formdan Al ---
    tickerText := r.FormValue("ticker_text")
    tickerActive := r.FormValue("ticker_active") == "true" // "true" gelirse true, aksi halde false
    tickerStart := r.FormValue("ticker_start_at") 
    tickerEnd := r.FormValue("ticker_end_at")

    now := time.Now().UnixNano()
	// Her güncellemede cihazın değişikliği anlaması için hash yeniliyoruz
    versionHash := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%d", now))))

    // 2. SQL Hazırlığı
    // Temel alanları ve yeni ticker alanlarını ekliyoruz
    // NULLIF(?,'') kullanarak eğer tarih boş gelirse DB'ye boş string değil NULL yazılmasını sağlıyoruz
    sqlQueryBase := `UPDATE screens SET 
        restaurant_id = ?, 
        screen_code = ?, 
        content_type = ?, 
        version_hash = ?,
        ticker_text = ?,
        ticker_active = ?,
        ticker_start_at = NULLIF(?, ''), 
        ticker_end_at = NULLIF(?, '')`
    
    var args []interface{}
    args = append(args, resID, screenCode, contentType, versionHash, tickerText, tickerActive, tickerStart, tickerEnd)

    // 3. Dosya kontrolü (Yeni dosya seçilmiş mi?)
    var mediaUpdateQuery string
    file, handler, err := r.FormFile("media_file")
    if err == nil {
        defer file.Close()
        fileName := fmt.Sprintf("%d-%s", now, handler.Filename)
        filePath := filepath.Join("./uploads", fileName)
        
        dst, err := os.Create(filePath)
        if err == nil {
            defer dst.Close()
            io.Copy(dst, file)
            mediaUpdateQuery = ", media_url = ?"
            args = append(args, "/uploads/"+fileName)
        }
    }

    // Sorguyu tamamla ve ID'yi ekle
    fullSQL := sqlQueryBase + mediaUpdateQuery + " WHERE id = ?"
    args = append(args, id)
    
    _, err = DB.Exec(fullSQL, args...)
    if err != nil {
        log.Printf("SQL Güncelleme Hatası (ID: %s): %v", id, err)
        http.Error(w, "Veritabanı güncelleme hatası: "+err.Error(), http.StatusInternalServerError)
        return
    }

    log.Printf("✅ Ekran ve Ticker başarıyla güncellendi: ID %s", id)
    w.Header().Set("Content-Type", "application/json")
    fmt.Fprint(w, `{"message": "Ekran ve Ticker güncellendi"}`)
}

func deleteScreen(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]
    DB.Exec("DELETE FROM screens WHERE id = ?", id)
    w.WriteHeader(http.StatusOK)
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
        logoURL = fileName
    }

    _, err = DB.Exec("INSERT INTO restaurants (name, phone, address, logo_url, created_at) VALUES (?, ?, ?, ?, NOW())", name, phone, address, logoURL)
    w.WriteHeader(http.StatusCreated)
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
        args = append(args, fileName)
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
	var s Screen

	// 1. Veritabanından tüm gerekli bilgileri alıyoruz
	query := `SELECT id, content_type, media_url, version_hash, 
                     ticker_text, ticker_start_at, ticker_end_at, ticker_active 
              FROM screens WHERE screen_code = ?`

	err := DB.QueryRow(query, vars["screen_code"]).Scan(
		&s.ID, &s.ContentType, &s.MediaURL, &s.VersionHash,
		&s.TickerText, &s.TickerStartAt, &s.TickerEndAt, &s.TickerActive,
	)

	if err != nil {
		log.Printf("Ekran bulunamadı: %s, Hata: %v", vars["screen_code"], err)
		http.Error(w, "Ekran bulunamadı", http.StatusNotFound)
		return
	}

	// 2. Cihazın online olduğunu güncelle
	DB.Exec("UPDATE screens SET last_seen_at = NOW(), is_online = true WHERE id = ?", s.ID)

	// 3. Zaman ve Aktiflik Kontrolü
	finalTicker := ""
	now := time.Now()

	// Eğer Ticker aktifse ve metin boş değilse
	if s.TickerActive && s.TickerText.Valid && s.TickerText.String != "" {
		// Eğer başlangıç ve bitiş tarihleri doluysa zaman kontrolü yap
		if s.TickerStartAt.Valid && s.TickerEndAt.Valid {
			if now.After(s.TickerStartAt.Time) && now.Before(s.TickerEndAt.Time) {
				finalTicker = s.TickerText.String
			}
		} else {
			// Tarihler boş ama aktifse direkt göster
			finalTicker = s.TickerText.String
		}
	}
        // TERMINALDE GÖRMEK İÇİN:
        log.Printf("📢 SCREEN: %s", vars["screen_code"])
        log.Printf("🕒 NOW: %v", now.Format("15:04:05"))
        log.Printf("✅ ACTIVE: %v | VALID: %v", s.TickerActive, s.TickerText.Valid)
        log.Printf("📝 FINAL TEXT: [%s]", finalTicker)

	// 4. JSON Yanıtını Oluştur ve Gönder
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(PlayerResponse{
		MediaURL:    s.MediaURL,
		ContentType: s.ContentType,
		VersionHash: s.VersionHash,
		TickerText:  finalTicker,
	})
}

func startCleanupTask() {
	go func() {
		for {
			// Her 24 saatte bir çalışsın
			time.Sleep(24 * time.Hour)
			log.Println("🧹 Gereksiz dosyalar için temizlik başlatılıyor...")

			// 1. Veritabanındaki tüm aktif medya dosyalarını al
			rows, err := DB.Query("SELECT media_url FROM screens UNION SELECT logo_url FROM restaurants")
			if err != nil {
				log.Println("Temizlik hatası (DB):", err)
				continue
			}

			activeFiles := make(map[string]bool)
			for rows.Next() {
				var fileName string
				if err := rows.Scan(&fileName); err == nil && fileName != "" {
					// Sadece dosya adını alıyoruz (eğer path ile kayıtlıysa temizliyoruz)
					activeFiles[filepath.Base(fileName)] = true
				}
			}
			rows.Close()

			// 2. Uploads klasöründeki dosyaları tara
			files, err := os.ReadDir("./uploads")
			if err != nil {
				log.Println("Temizlik hatası (Klasör):", err)
				continue
			}

			for _, file := range files {
				if file.IsDir() {
					continue
				}
				// Eğer dosya veritabanında yoksa sil
				if !activeFiles[file.Name()] {
					err := os.Remove(filepath.Join("./uploads", file.Name()))
					if err == nil {
						log.Printf("🗑️ Silindi: %s", file.Name())
					}
				}
			}
			log.Println("✅ Temizlik tamamlandı.")
		}
	}()
}