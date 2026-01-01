-- Restoranlar Tablosu
CREATE TABLE restaurants (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    address TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Ekranlar Tablosu
CREATE TABLE screens (
    id INT AUTO_INCREMENT PRIMARY KEY,
    screen_code VARCHAR(50) UNIQUE NOT NULL, -- Cihazın benzersiz kodu (Örn: REST1_EKRAN1)
    restaurant_id INT,
    content_type ENUM('video', 'image') DEFAULT 'video', -- İçerik türü
    media_url VARCHAR(500), -- VPS üzerindeki dosyanın yolu
    version_hash VARCHAR(64), -- Dosya değişti mi kontrolü için (Örn: md5 hash)
    last_seen_at TIMESTAMP NULL, -- Cihaz en son ne zaman sinyal verdi?
    is_online BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (restaurant_id) REFERENCES restaurants(id) ON DELETE CASCADE
);
