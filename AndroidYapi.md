📱 SignaPanel Android Player Geliştirme Rehberi
Bu döküman, Android cihazın bir "Digital Signage Player" olarak çalışması için gereken temel fonksiyonları ve Go backend ile iletişimini kapsar.

1. Temel Gereksinimler
İnternet İzni: AndroidManifest.xml dosyasına internet erişim izni eklenmelidir.

Kütüphane: HTTP istekleri için Retrofit veya OkHttp, görsel yükleme için Glide veya Coil önerilir.

Donanım: Cihazın sürekli açık kalması için "WakeLock" veya "Keep Screen On" özelliği aktif edilmelidir.

2. API Uç Noktaları (Endpoints)
Backend ile iletişim kuracağın iki ana adres:

Config Alımı: GET /api/v1/player/{screen_code}

Ekran açıldığında hangi görseli/videoyu oynatacağını buradan öğrenir.

Heartbeat (Sinyal): POST /api/heartbeat

Cihazın "Online" görünmesi için düzenli olarak buraya istek atılır.

3. Uygulama Mantığı (Logic)
A. Başlangıç (Initialization)
Uygulama açıldığında cihazın benzersiz screen_code bilgisiyle backend'e sor: "Benim içeriğim ne?"

İstek: GET /api/v1/player/ABC123 Yanıt:

JSON

{
  "media_url": "/uploads/12345-banner.jpg",
  "content_type": "image",
  "version_hash": "a1b2c3d4..."
}
Gelen media_url'i tam adrese çevir (örn: http://sunucu-ip:8081/uploads/...) ve ekranda göster.

version_hash bilgisini yerel hafızaya (SharedPreferences) kaydet.

B. Heartbeat (Sinyal Gönderimi) - Online Görünme İçin Kritik
Cihazın hayatta olduğunu kanıtlamak için bir Handler veya WorkManager kullanarak her 30 saniyede bir şu isteği gönder:

İstek: POST /api/heartbeat Body:

JSON

{
  "screen_code": "ABC123",
  "current_hash": "a1b2c3d4..."
}
Backend Yanıtına Göre Aksiyon:

Eğer yanıt içindeki needs_update: true dönerse; bu, panelden yeni bir içerik yüklendiği anlamına gelir.

Bu durumda uygulamayı durdurmadan yeni media_url'i çek ve ekranı güncelle.

4. Oynatıcı Kontrolleri (UI)
Görseller: ImageView kullanarak centerCrop ölçeklendirmesi yapın.

Videolar: VideoView veya ExoPlayer kullanın.

isLooping = true (Sürekli dönmeli)

volume = 0 (Genelde reklam ekranları sessiz olur)

Hata Yönetimi: Eğer internet kesilirse, ekranda çirkin bir hata mesajı yerine "Lütfen internet bağlantısını kontrol edin" yazan şık bir görsel gösterin.