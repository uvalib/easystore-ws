package main

import (
	"github.com/gin-gonic/gin"
	"github.com/uvalib/easystore/uvaeasystore"
	"log"
	"net/http"
)

// this is our service implementation
type serviceImpl struct {
	EasyStore uvaeasystore.EasyStore
	cfg       *ServiceConfig
}

func NewService(cfg *ServiceConfig) *serviceImpl {

	es, err := uvaeasystore.NewEasyStore(cfg.esCfg)
	if err != nil {
		log.Fatalf("create easystore failed: %s", err.Error())
	}

	return &serviceImpl{EasyStore: es, cfg: cfg}
}

// IgnoreFavicon is a dummy to handle browser favicon requests without warnings
func (s *serviceImpl) IgnoreFavicon(c *gin.Context) {
}

// GetVersion reports the version of the service
func (s *serviceImpl) GetVersion(c *gin.Context) {

	vMap := make(map[string]string)
	vMap["build"] = Version()
	c.JSON(http.StatusOK, vMap)
}

// HealthCheck reports the health of the service
func (s *serviceImpl) HealthCheck(c *gin.Context) {

	type hcResp struct {
		Healthy bool   `json:"healthy"`
		Message string `json:"message,omitempty"`
	}
	hcMap := make(map[string]hcResp)
	hcMap["mint-token"] = hcResp{Healthy: true}

	c.JSON(http.StatusOK, hcMap)
}

// MintToken creates a new token
//func (s *serviceImpl) MintToken(c *gin.Context) {
//
//	tMap := make(map[string]string)
//	token, expires := s.makeToken()
//	tMap["token"] = token
//	tMap["expires"] = expires.Format("2006-01-02T15:04:05-0700")
//	c.JSON(http.StatusOK, tMap)
//}
//
//// RenewToken validates the supplied tonen and if valid yields a new token
//func (s *serviceImpl) RenewToken(c *gin.Context) {
//
//	authorization := c.Request.Header.Get("Authorization")
//	components := strings.Split(strings.Join(strings.Fields(authorization), " "), " ")
//
//	// must have two components, the first of which is "Token", and the second a non-empty token
//	if len(components) != 2 || components[0] != "Token" || components[1] == "" {
//		log.Printf("ERROR: invalid Authorization header: [%s]", authorization)
//		c.AbortWithStatus(http.StatusUnauthorized)
//		return
//	}
//
//	// validate the token
//	if s.validateToken(components[1]) == false {
//		log.Printf("ERROR: invalid token in header: [%s]", components[1])
//		c.AbortWithStatus(http.StatusUnauthorized)
//		return
//	}
//
//	tMap := make(map[string]string)
//	token, expires := s.makeToken()
//	tMap["token"] = token
//	tMap["expires"] = expires.Format("2006-01-02T15:04:05-0700")
//	c.JSON(http.StatusOK, tMap)
//}
//
//// creates a new token
//func (s *serviceImpl) makeToken() (string, time.Time) {
//
//	// Declare the expiration time of the token
//	expirationTime := time.Now().Add(time.Duration(s.cfg.ExpireDays*24) * time.Hour)
//
//	// Create the JWT claims, which includes expiry time
//	claims := &jwt.StandardClaims{
//		// In JWT, the expiry time is expressed as unix milliseconds
//		ExpiresAt: expirationTime.Unix(),
//	}
//
//	// Declare the token with the algorithm used for signing, and the claims
//	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
//
//	// Create the JWT string
//	tokenString, err := token.SignedString([]byte(s.cfg.SharedSecret))
//	if err != nil {
//		log.Fatal(err)
//	}
//	return tokenString, expirationTime
//}
//
//// validates the supplied token
//func (s *serviceImpl) validateToken(token string) bool {
//
//	// Initialize a new instance of the standard claims
//	claims := &jwt.StandardClaims{}
//
//	tkn, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
//		return []byte(s.cfg.SharedSecret), nil
//	})
//
//	if err != nil {
//		log.Printf("ERROR: JWT parse returns: %s", err.Error())
//		return false
//	}
//
//	if !tkn.Valid {
//		log.Printf("ERROR: JWT is INVALID")
//		return false
//	} else {
//		log.Printf("INFO: token is valid, Expires %s", time.Unix(claims.ExpiresAt, 0))
//	}
//	return true
//}

//
// end of file
//
