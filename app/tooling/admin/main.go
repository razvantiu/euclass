package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ardanlabs/service/business/data/dbschema"
	"github.com/ardanlabs/service/business/sys/auth"
	"github.com/ardanlabs/service/business/sys/database"
	"github.com/ardanlabs/service/foundation/keystore"
	"github.com/golang-jwt/jwt/v4"
)

func main() {
	err := migrate()

	if err != nil {
		log.Fatalln(err)
	}
}

func migrate() error {
	dbConfig := database.Config{
		User:       "postgres",
		Password:   "postgres",
		Host:       "localhost",
		Name:       "postgres",
		DisableTLS: true,
	}

	db, err := database.Open(dbConfig)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := dbschema.Migrate(ctx, db); err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}

	fmt.Println("migrations complete")

	if err := dbschema.Seed(ctx, db); err != nil {
		return fmt.Errorf("seed database: %w", err)
	}

	fmt.Println("seed data complete")

	return nil
}

func gentoken2() error {

	// Construct a key store based on the key files stored in
	// the specified directory.
	keysFolder := "zarf/keys/"
	ks, err := keystore.NewFS(os.DirFS(keysFolder))
	if err != nil {
		return fmt.Errorf("reading keys: %w", err)
	}

	// Init the auth package.
	activeKID := "54bb2165-71e1-41a6-af3e-7da4a0e1e2c1"
	a, err := auth.New(activeKID, ks)
	if err != nil {
		return fmt.Errorf("constructing auth: %w", err)
	}

	claims := auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "123456789",
			Issuer:    "service project",
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(8760 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
		Roles: []string{auth.RoleAdmin},
	}

	token, err := a.GenerateToken(claims)
	if err != nil {
		return fmt.Errorf("generating token: %w", err)
	}

	fmt.Printf("-----BEGIN TOKEN-----\n%s\n-----END TOKEN-----\n", token)
	return nil
}

func gentoken() error {

	// Claims represents the authorization claims transmitted via a JWT.
	type Claims struct {
		jwt.RegisteredClaims
		Roles []string `json:"roles"`
	}

	// iss (issuer): Issuer of the JWT
	// sub (subject): Subject of the JWT (the user)
	// aud (audience): Recipient for which the JWT is intended
	// exp (expiration time): Time after which the JWT expires
	// nbf (not before time): Time before which the JWT must not be accepted for processing
	// iat (issued at time): Time at which the JWT was issued; can be used to determine age of the JWT
	// jti (JWT ID): Unique identifier; can be used to prevent the JWT from being replayed (allows a token to be used only once)
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "1234567",
			Issuer:    "service project",
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(8760 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
		Roles: []string{"ADMIN"},
	}

	method := jwt.GetSigningMethod("RS256")
	if method == nil {
		return errors.New("configuring algorithm RS256")
	}

	token := jwt.NewWithClaims(method, claims)
	token.Header["kid"] = "123123-235345-456-456"

	// Generate a new private key.
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	tokenStr, err := token.SignedString(privateKey)
	if err != nil {
		return fmt.Errorf("signing token: %w", err)
	}
	fmt.Printf("-----BEGIN TOKEN-----\n%s\n-----END TOKEN-----\n", tokenStr)

	// =========================================================================

	// Marshal the public key from the private key to PKIX.
	asn1Bytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("marshaling public key: %w", err)
	}

	// Construct a PEM block for the public key.
	publicBlock := pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: asn1Bytes,
	}

	// Write the public key to the private key file.
	if err := pem.Encode(os.Stdout, &publicBlock); err != nil {
		return fmt.Errorf("encoding to public file: %w", err)
	}

	// =========================================================================

	parser := jwt.NewParser(jwt.WithValidMethods([]string{"RS256"}))

	keyFunc := func(t *jwt.Token) (interface{}, error) {
		kid := t.Header["kid"]
		kidID := kid.(string)
		fmt.Println("KID", kidID)

		return &privateKey.PublicKey, nil
	}

	var xclaims Claims
	xtoken, err := parser.ParseWithClaims(tokenStr, &xclaims, keyFunc)
	if err != nil {
		return fmt.Errorf("parsing token: %w", err)
	}

	if !xtoken.Valid {
		return errors.New("invalid token")
	}

	fmt.Println("TOKEN VALIDATED")
	fmt.Printf("%#v\n", xclaims)

	// Maybe now hit the database for the user id and validate more!!!

	return nil
}

func genkey() error {

	// Generate a new private key.
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	// Create a file for the private key information in PEM form.
	privateFile, err := os.Create("private.pem")
	if err != nil {
		return fmt.Errorf("creating private file: %w", err)
	}
	defer privateFile.Close()

	// Construct a PEM block for the private key.
	privateBlock := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	// Write the private key to the private key file.
	if err := pem.Encode(privateFile, &privateBlock); err != nil {
		return fmt.Errorf("encoding to private file: %w", err)
	}

	// Marshal the public key from the private key to PKIX.
	asn1Bytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("marshaling public key: %w", err)
	}

	// Create a file for the public key information in PEM form.
	publicFile, err := os.Create("public.pem")
	if err != nil {
		return fmt.Errorf("creating public file: %w", err)
	}
	defer publicFile.Close()

	// Construct a PEM block for the public key.
	publicBlock := pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: asn1Bytes,
	}

	// Write the public key to the private key file.
	if err := pem.Encode(publicFile, &publicBlock); err != nil {
		return fmt.Errorf("encoding to public file: %w", err)
	}

	fmt.Println("private and public key files generated")
	return nil

}
