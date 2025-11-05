package config

// JWTConfig holds the configuration for JWT authentication.
//
// JWTIssuer is the issuer field for all tokens.
// JWTAudience is the audience field for tokens.
//
// JWTAccessSecretKey is the secret used for signing access tokens (must be >= 32 bytes).
// JWTAccessExpTimeMinute is the lifetime of access tokens in minutes.
//
// JWTRefreshSecretKey is the secret used for signing refresh tokens (must be >= 32 bytes).
// JWTRefreshExpTimeDays is the lifetime of refresh tokens in days.
type JWTConfig struct {
	JWTIssuer   string `envDefault:"passKeeper"`
	JWTAudience string `envDefault:"passKeeper-server"`

	JWTAccessSecretKey     string `envDefault:"12345678901234567890123456789012"`
	JWTAccessExpTimeMinute int    `envDefault:"15"`

	JWTRefreshSecretKey   string `envDefault:"12345678901234567890123456789012"`
	JWTRefreshExpTimeDays int    `envDefault:"30"`
}
