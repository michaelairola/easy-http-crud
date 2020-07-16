package easyHttpCrud

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"

	"github.com/dgrijalva/jwt-go"
)

// FOR INJECTING PROPER ENV VARIABLES INTO THE DESIRED CONTAINERS

var (
	Port = 8080
	// for all services to connect to db and other services within the Gate
	// call EstablishEnvironment() to populate these variables with the environment variables in the container
	HostDbMap map[string]string
	HiveHost  string
	SqlProxy  string
	// Conn          string
	// Gate          string
	// DbName        string
	Version       string
	HostName      string
	DevMode       bool
	Debug         bool
	MigrationDown bool
	// this is for certain changes within dev mode
	OS string

	// for licensor stuff
	// call EstablishInitEmail() to populate this variable with the environment variable. ONLY FOR TESTING currently
	Email     string
	FirstName string
	LastName  string
	// for any services that have access to creating or accessing JWTokens
	// currently, only signin has these priveledges
	// call EstablishKeys() to populate these variables with the environment variables in the container
	PrivateKey   *rsa.PrivateKey
	PublicKey    *rsa.PublicKey
	ClientId     string
	ClientSecret string
	JWTIssuer    string

	// TODO HiveCluster data in profile should be here?
	// for any services that need access to any api calls on defie's cluster
	// only profile has these priveledes (to get access to licensor apis)
	// call EstablishHiveClusterConn() to populate these variables with the environment variables in the container
	// HiveClusterGate    string
	// HiveClusterPrivKey string
)

func EstablishEnvironment() {
	devOrProd(func() { DevMode = true }, func() {})
	devOrProd(getDevEnv, getProdEnv)
	getEnvVar("HOSTNAME", &HostName, "localhost")
	getEnvVar("SQL_PROXY", &SqlProxy, "localhost")
	HostDbMap = ParseHostDbMap()
}

func EstablishKeys() {
	devOrProd(getDevKeys, getProdKeys)
}

func EstablishInitEmail() {
	devOrProd(getDevEmail, getLaurasEmail)
}

func getProdEnv() {
	fmt.Println("Running service in Production mode")
	getEnvVar("VERSION", &Version, "")
	warnNoEnvVar(Version)
}

func getProdKeys() {
	var err error
	var publicKeyPEM string
	getEnvVar("PUBLIC_KEY", &publicKeyPEM, "")
	keyBytes := []byte(publicKeyPEM)
	PublicKey, err = jwt.ParseRSAPublicKeyFromPEM(keyBytes)
	if warnNoEnvVar(publicKeyPEM) == nil && err != nil {
		fmt.Println("WARNING: PUBLIC_KEY environment variable not parseable")
	}

	var privateKeyPEM string
	getEnvVar("PRIVATE_KEY", &privateKeyPEM, "")
	keyBytes = []byte(privateKeyPEM)
	PrivateKey, err = jwt.ParseRSAPrivateKeyFromPEM(keyBytes)
	if warnNoEnvVar(privateKeyPEM) == nil && err != nil {
		fmt.Println("WARNING: PRIVATE_KEY environment variable not parseable")
	}
	getEnvVar("CLIENT_ID", &ClientId, "")
	warnNoEnvVar(ClientId)
	getEnvVar("CLIENT_SECRET", &ClientSecret, "")
	warnNoEnvVar(ClientSecret)
	getEnvVar("JWT_ISSUER", &JWTIssuer, "")
	warnNoEnvVar(JWTIssuer)
}

func getDevEnv() {
	fmt.Println("Running service in Development mode")
	getEnvVarInt("PORT", &Port, 8080)
	getEnvVar("OS", &OS, "linux")
	Version = "Development"
	getEnvVarBool("DEBUG", &Debug, false)
	getEnvVarBool("MIGRATION_DOWN", &MigrationDown, false)
	// getEnvVar("DB_NAME", &DbName, "defiedev")
	// getEnvVar("GATE", &Gate, "http://localhost")
	// getEnvVar("DB_CONN", &Conn, "root:password@/defie?charset=utf8&parseTime=True&loc=Local")
}

func getDevKeys() {
	JWTIssuer = "fqsystems.company"
	getEnvVar("CLIENT_ID", &ClientId, "801031897599-rhci67sjv62jp2agbkk0nfr0npvjekmp.apps.googleusercontent.com")
	getEnvVar("CLIENT_SECRET", &ClientSecret, "Zz-AzlNhuo-35DxC46Vjx9QV")

	PrivateKey, _ = jwt.ParseRSAPrivateKeyFromPEM(
		[]byte(`-----BEGIN PRIVATE KEY-----
MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQDGv9ZAchtwRDKG
qptC6dMF7kb4cc5gzoT+lE+MFDzd75AqexgdEaZoQ+tGFg6xzF2mXVD53ao4u/gK
w5NldYY5mFG+j74lvmIGy+hICKLp7WyIpubxXRf6qMPBdVgF+veDZE2DZLIbu0ao
/cX3nqL9DmiZM3QSBtv4KkYZn22Y3WMLqlEmGhEom96aDpx3URI/XZ0EaV1ZWm1B
GDrjyV5fkEobNDH9NsjBlhdifT60Yt9o45hnVVYJ1QV32UtYoSYo5PWkTvpkSTpR
j0clH3l0LDESwz6e91XcbWxD0U51CheSKr3TuspQ1xKYEkwujLuD7CbcB5119BmW
kO5HfhC9AgMBAAECggEAH/s5HPwvlaBnt/tGih0pn4vw4CQiBzpcNMIFTd9ozvaw
bmGJ5w2SxzbcqlW8zhf6Xt2nvNlSPZhjqMnBU2N2sphj2QP030p0KC9SJJs9KeLS
YufmhCLMi8Fx5JS/EhFJGFvAzFqc/XDkhSd23mpoxEs4AiUBMbBoX5Xf21on7t+y
Why0OVJsffDbTJFGRw+PDOb9ZUCjYwSn9KR8r+X2AfAmJmN5UqHPctjIrOVoB+/+
+beqpSEykTuM2JyR98fBGpiEUFTJ4cLbrRzj/pcFLpsSakovIWtaSaqxFEhxi7P2
ffjZ7P0IOKRcSIMfoNUO2oPNzIgFnleSH22mmLO2uQKBgQDrbLHU63jbxOTbsqZ2
o7hmo4qdt8r5HPGLa8LKldbsPfuNSOBt4ikIql2dhZ6E1kntg9pew7q3BIblpnP0
BiycO9V3g5Uteiouuo4Niujj8tcJAKPTbNx+qC5M2ao0dhPXdYD9osufPhBIjJ1+
Y3+sEfDb/y4GO+WXHT21ItuhdQKBgQDYHpV7oKkpxvGArO0WpwmEU4mt0eBi5aSh
TXqlAKtjejx6Ga/qYpPKNasN6ak7O9qF/H6a5X8pQefCj5gMzpZIuf0ZsWW2TQ4V
ommZgIFkvUe1IyGTORWZJnJWXkNvUdTDn9TuxeZhscCxIR45GLuvdykZIHkLcQLX
GOQwHDjBKQKBgQCCXa9gC4DimfZtXlFl0yVy2M8SpUslhYyQOv0j97OLIIui4h89
WgaFAOpUJ0DvqEZJ20DaRyKm5D/a/cCp+Z1Maqm7w8ag50xikfXFtA1g5QGEv7D4
jAjRrkntTn89C8HTBwOF8Fc3eEPGGO/BpFF3tHufJ4CR/SsW8h9QpdvfXQKBgQCN
2x1RRXZj5qKUTG1exgtYISEox67aG1QBUILlFVhp9k0F9CBk34mUW7/IZQvgxnu8
IKzVbTgXU5wVN/2cwmkgmXwJRki17UQGbVGVISVNg3Qy8DlUkk6b9QFtsXff9L2M
Hjg89gcvLuaFHXDiF1ryZGvEg20QVoB/wWXeMlgKaQKBgDKFgvXdvOV5EHbkVlxC
TUZT7UAI77tKxeAXC3bvSx6bk804gNiLjE7pgIbKxYnTzbWyMBAKJpkzRBkyJVcE
kouiF6c5lBYxxUk2tE6JIvoTIjh9iQB93bGeV4TqqvuWmwiPzYjLGKIoEkBl/PRu
3SyBMuobE+jIAEYHYTWdOOY5
-----END PRIVATE KEY-----
`))

	PublicKey, _ = jwt.ParseRSAPublicKeyFromPEM(
		[]byte(`-----BEGIN CERTIFICATE-----
MIIC+jCCAeKgAwIBAgIIBqgfhTziZDMwDQYJKoZIhvcNAQEFBQAwIDEeMBwGA1UE
AxMVMTExNDk1NTU0NDkwNDAwODc0MjYyMB4XDTE4MDcyODAwNDAwNloXDTI4MDcy
NTAwNDAwNlowIDEeMBwGA1UEAxMVMTExNDk1NTU0NDkwNDAwODc0MjYyMIIBIjAN
BgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAxr/WQHIbcEQyhqqbQunTBe5G+HHO
YM6E/pRPjBQ83e+QKnsYHRGmaEPrRhYOscxdpl1Q+d2qOLv4CsOTZXWGOZhRvo++
Jb5iBsvoSAii6e1siKbm8V0X+qjDwXVYBfr3g2RNg2SyG7tGqP3F956i/Q5omTN0
Egbb+CpGGZ9tmN1jC6pRJhoRKJvemg6cd1ESP12dBGldWVptQRg648leX5BKGzQx
/TbIwZYXYn0+tGLfaOOYZ1VWCdUFd9lLWKEmKOT1pE76ZEk6UY9HJR95dCwxEsM+
nvdV3G1sQ9FOdQoXkiq907rKUNcSmBJMLoy7g+wm3AeddfQZlpDuR34QvQIDAQAB
ozgwNjAMBgNVHRMBAf8EAjAAMA4GA1UdDwEB/wQEAwIHgDAWBgNVHSUBAf8EDDAK
BggrBgEFBQcDAjANBgkqhkiG9w0BAQUFAAOCAQEAkWnuS+Mtk+E1VmAR5P3DwGZ7
NxSgPsC+TzsYbOpDzQDRKTfruqMF2j7HqYec9YgjZKJgiioOkHxCYfmIQkmX04dG
vdWJybFH3hyqi3fsr/sab/2A8J6g3jTmV/4qi8+Vl0tJ/LUwu7Cr6WadT1038h0l
5QwEa4b/AwpgR0xNcGJZ/dh7WsMmSqc658Jm0RewPwLgXD6zZjFi2sPEdWqlrTeT
gX9T/UIso5T+cyKI5F3MjR4uT8r5kZReid964QAxZ5a6smkklV0uZKj0DuIyAiyw
9qI1nE0aPxqJ7KBjHAaeYfnx77WmNhvccTE5x1qH54FVgMHhYIj0W5FcaNkZhw==
-----END CERTIFICATE-----
`))
}

func getLaurasEmail() {
	Email = "lduong@defie.co"
	FirstName = "Laura"
	LastName = "Duong"
}
func getDevEmail() {
	getEnvVar("EMAIL", &Email, "developer@defie.co")
	getEnvVar("FIRST_NAME", &FirstName, "Dev")
	getEnvVar("LAST_NAME", &LastName, "Eloper")
}

// Private helper functions

func devOrProd(devFn func(), prodFn func()) {
	dev_mode, prod_mode := strconv.ParseBool(os.Getenv("DEV_MODE"))
	if prod_mode != nil {
		prodFn()
	} else if dev_mode {
		devFn()
	} else {
		fmt.Println("WARNING: environment variable DEV_MODE set to false. Default mode of all services is set to production mode, so it is unnecessary to set DEV_MODE if production is desired.")
		prodFn()
	}
}

func getEnvVar(env_variable string, holder *string, default_value string) error {
	*holder = os.Getenv(env_variable)
	if *holder == "" {
		*holder = default_value
	}
	return nil
}
func getEnvVarInt(env_variable string, holder *int, default_value int) error {
	var err error
	interim_holder := os.Getenv(env_variable)
	*holder, err = strconv.Atoi(interim_holder)
	if interim_holder == "" || err != nil {
		*holder = default_value
	}
	return nil
}
func getEnvVarBool(env_variable string, holder *bool, default_value bool) error {
	var err error
	interim_holder := os.Getenv(env_variable)
	*holder, err = strconv.ParseBool(interim_holder)
	if interim_holder == "" || err != nil {
		*holder = default_value
	}
	return nil
}

func warnNoEnvVar(env_var string) error {
	if env_var == "" {
		message := "WARNING: no " + env_var + " environment variable injected upon deployment of service"
		fmt.Println(message)
		return errors.New("NO ENV VAR")
	}
	return nil
}
