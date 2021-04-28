package auth

type JWK struct {
	Keys []*JWK `json:"keys,omitempty"`

	Kty string `json:"kty"`
	Use string `json:"use,omitempty"`
	Kid string `json:"kid,omitempty"`
	Alg string `json:"alg,omitempty"`

	Crv string `json:"crv,omitempty"`
	X   string `json:"x,omitempty"`
	Y   string `json:"y,omitempty"`
	D   string `json:"d,omitempty"`
	N   string `json:"n,omitempty"`
	E   string `json:"e,omitempty"`
	K   string `json:"k,omitempty"`
}

type KeyList struct{ Keys []JWK }

func (ks *KeyList) GetKey(kid string) (*JWK, bool) {
	for _, key := range ks.Keys {
		if key.Kid == kid {
			return &key, true
		}
	}
	return &JWK{}, false
}
