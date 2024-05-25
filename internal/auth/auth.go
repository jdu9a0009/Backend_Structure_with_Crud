package auth

import (
	"crypto/rsa"
	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"strings"
	"sync"
)

// These are the expected values for Claims.Roles.
const (
	RoleAdmin    = "ADMIN"
	RoleEmployee = "EMPLOYEE"
)

// ctxKey represents the type of value for the context key.
type ctxKey int

// Key is used to store/retrieve a Claims value from a context.Context.
const Key ctxKey = 1

// Claims represents the authorization claims transmitted via a JWT.
type Claims struct {
	jwt.StandardClaims
	UserId int    `json:"user_id"`
	Role   string `json:"roles"`
}

type ClaimsParse struct {
	jwt.StandardClaims
	UserId string `json:"user_id"`
	Role   string `json:"roles"`
}

// Authorized returns true if the claims has at least one of the provided roles.
func (c Claims) Authorized(role ...string) bool {
	//if strings.Compare(role, c.Role) == 0 {
	//	return true
	//}

	for _, r := range role {
		if strings.Compare(r, c.Role) == 0 {
			return true
		}
	}

	return false
}

// Keys represents an in memory storage of keys.
type Keys map[string]*rsa.PrivateKey

// PublicKeyLookup defines the signature of a function to lookup public keys.
//
// In a production system, a key id (KID) is used to retrieve the correct
// public key to parse a JWT for auth and claims. A key lookup function is
// provided to perform the task of retrieving a KID for a given public key.
//
// A key lookup function is required for creating an Authenticator.
//
// * Private keys should be rotated. During the transition period, tokens
// signed with the old and new keys can coexist by looking up the correct
// public key by KID.
//
// * KID to public key resolution is usually accomplished via a public JWKS
// endpoint. See https://auth0.com/docs/jwks for more details.
type PublicKeyLookup func(kid string) (*rsa.PublicKey, error)

// Auth is used to authenticate clients. It can generate a token for a
// set of area claims and recreate the claims by parsing the token.
type Auth struct {
	mu        sync.RWMutex
	algorithm string
	method    jwt.SigningMethod
	keyFunc   func(t *jwt.Token) (interface{}, error)
	parser    *jwt.Parser
	keys      Keys
}

// New creates an *Authenticator for use.
func New(algorithm string, lookup PublicKeyLookup, keys Keys) (*Auth, error) {
	method := jwt.GetSigningMethod(algorithm)
	if method == nil {
		return nil, errors.Errorf("unknown algorithm %v", algorithm)
	}

	keyFunc := func(t *jwt.Token) (interface{}, error) {
		kid, ok := t.Header["kid"]
		if !ok {
			return nil, errors.New("missing key id (kid) in token header")
		}
		kidID, ok := kid.(string)
		if !ok {
			return nil, errors.New("area token key id (kid) must be string")
		}

		return lookup(kidID)
	}

	// Create the token parser to use. The algorithm used to sign the JWT must be
	// validated to avoid a critical vulnerability:
	// https://auth0.com/blog/critical-vulnerabilities-in-json-web-token-libraries/
	parser := jwt.Parser{
		ValidMethods: []string{algorithm},
	}

	a := Auth{
		algorithm: algorithm,
		method:    method,
		keyFunc:   keyFunc,
		parser:    &parser,
		keys:      keys,
	}

	return &a, nil
}

// AddKey adds a private key and combination kid id to our local store.
func (a *Auth) AddKey(privateKey *rsa.PrivateKey, kid string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.keys[kid] = privateKey
}

// RemoveKey removes a private key and combination kid id to our local store.
func (a *Auth) RemoveKey(kid string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.keys, kid)
}

// GenerateToken generates a signed JWT token string representing the area Claims.
func (a *Auth) GenerateToken(kid string, claims Claims) (string, error) {
	token := jwt.NewWithClaims(a.method, claims)
	token.Header["kid"] = kid

	var privateKey *rsa.PrivateKey
	a.mu.RLock()
	{
		var ok bool
		privateKey, ok = a.keys[kid]
		if !ok {
			return "", errors.New("kid lookup failed")
		}
	}
	a.mu.RUnlock()

	str, err := token.SignedString(privateKey)
	if err != nil {
		return "", errors.Wrap(err, "signing token")
	}

	return str, nil
}

// ValidateToken recreates the Claims that were used to generate a token. It
// verifies that the token was signed using our key.
func (a *Auth) ValidateToken(tokenStr string) (Claims, error) {
	var claims Claims
	token, err := a.parser.ParseWithClaims(tokenStr, &claims, a.keyFunc)
	if err != nil {
		return Claims{}, errors.Wrap(err, "parsing token")
	}

	if !token.Valid {
		return Claims{}, errors.New("invalid token")
	}

	//userId, err := strconv.Atoi(claimsParse.UserId)
	//if err != nil {
	//	return Claims{}, errors.Wrap(err, "parsing user_id")
	//}

	//return Claims{
	//	claimsParse.StandardClaims,
	//	userId,
	//	claimsParse.Role,
	//}, nil
	return claims, nil
}

// Authorized returns true if the claims has at least one of the provided roles.
//func (c Claims) Authorized(role string) bool {
//	if strings.Compare(role, c.Role) == 0 {
//		return true
//	}
//	return false
//
//	//for _, has := range c.Roles {
//	//	for _, want := range roles {
//	//		if want == has {
//	//			return true
//	//		}
//	//	}
//	//}
//}

//func (a *Auth) CheckClaimsDataFromDatabase(ctx context.Context, claims Claims) error {
//	a.mu.Lock()
//	defer a.mu.Unlock()
//
//	userDetail := entity.User{}
//
//	if err := a.postgresDB.
//		NewSelect().
//		Model(userDetail).
//		Where(
//			"id = ? AND organization_id = ? AND deleted_at IS NULL",
//			claims.UserId,
//			claims.OrganizationID,
//		).
//		Scan(ctx); err != nil {
//		if err == sql.ErrNoRows {
//			return web.NewRequestError(errors.New("area not found"), http.StatusUnauthorized)
//		}
//
//		return web.NewRequestError(errors.Wrap(err, "area not found"), http.StatusUnauthorized)
//	}
//
//	return nil
//}
