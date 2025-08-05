# Traefik Net Gateway Middleware

Bu Traefik middleware plugin’i, gelen HTTP isteklerindeki `X-Api-Key` header’ını kullanarak bir OAuth token endpoint’ine doğrulama isteği gönderir. Doğrulama başarılı olursa istek devam eder, aksi halde `401 Unauthorized` döner.

Ayrıca, her işlem UDP üzerinden merkezi log sunucusuna JSON formatında log olarak gönderilir.

---

## 🔧 Plugin Config Parametreleri

Aşağıdaki parametreler `dynamic configuration` içinde plugin'e atanmalıdır.

### Authentication Ayarları

| Parametre      | Tip    | Açıklama                                                                  |
| -------------- | ------ | ------------------------------------------------------------------------- |
| `authUrl`      | string | OAuth token doğrulama endpoint’i (örnek: `http://auth.local/oauth/token`) |
| `clientId`     | string | OAuth Client ID                                                           |
| `clientSecret` | string | OAuth Client Secret                                                       |
| `redirectUri`  | string | OAuth Redirect URI                                                        |
| `scope`        | string | OAuth scope değeri (örnek: `sms`)                                         |

### Log Ayarları

| Parametre      | Tip    | Açıklama                                                   |
| -------------- | ------ | ---------------------------------------------------------- |
| `logHost`      | string | Log sunucusunun host adresi (örnek: ``) |
| `logPort`      | int    | UDP log portu (default: `514`)                             |
| `logProgram`   | string | Uygulama ismi (örnek: ``)                |
| `logService`   | string | Hizmet ismi (örnek: ``)                                |
| `logRetention` | string | Log retention süresi (örnek: `yearly`, `monthly`)          |
