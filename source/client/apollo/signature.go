package apollo

import (
    "crypto/hmac"
    "crypto/sha1"
    "encoding/base64"
    "fmt"
    "net/http"
    "time"
)

func hmacSHA1(data string, secret string) string {
    h := hmac.New(sha1.New, []byte(secret))
    h.Write([]byte(data))
    return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func sign(req *http.Request, appId, token string) {
    if len(token) > 0 {
        timestamp := fmt.Sprintf("%d", time.Now().UnixNano() / 1000000)
        strToSign := timestamp + "\n" + req.URL.String()
        signature := hmacSHA1(strToSign, token)
        req.Header.Set("Authorization", fmt.Sprintf("Apollo %s:%s", appId, signature))
        req.Header.Set("Timestamp", timestamp)
    }
}