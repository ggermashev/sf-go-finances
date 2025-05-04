package config

type CryptoConfig struct {
	PGPKey  string
	HMACKey string
}

func GetCryptoConfig() CryptoConfig {
	cfg := CryptoConfig{
		PGPKey:  "pgpkey",
		HMACKey: "hmackey",
	}

	return cfg
}