package auth

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

// Token will expire Thursday, February 14, 2030 7:18:54 PM GMT
const tokenString = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6IkN3Sk12ZXNLdUVqdkJPZFppaTM3dkZKbDJZMzBKWG41Y1ZPMzRielZhWTQifQ.eyJuYmYiOjAsImlhdCI6MTU4MTQyNzEzNCwiaXNzIjoiaHR0cHM6Ly9hY2NvdW50LW9pZGNtb2NrLW1mbS1nZW5lcmFsLmFlLmRldi5jbG91ZGhoLmRlL2F1dGgvcmVhbG1zL0ZpZWxtYW5uIiwiYXVkIjpbIm1mbSJdLCJzdWIiOiJmMTM1MTc3OS00ZDUzLTRjNzUtOWU1Ny1hYmYyM2FlMGQ3MzkiLCJ0eXAiOiJCZWFyZXIiLCJhenAiOiJtZm0tYWNjb3VudC1mZSIsImF1dGhfdGltZSI6MCwic2Vzc2lvbl9zdGF0ZSI6ImYxMzUxNzc5LTRkNTMtNGM3NS05ZTU3LWFiZjIzYWUwZDczOSIsImFjciI6IjEiLCJhbGxvd2VkLW9yaWdpbnMiOlsiKiJdLCJyb2xlcyI6W10sInNjb3BlIjoibWZtLXB1YmxpYyIsIm1mbSI6eyJhY2NvdW50LWlkIjoiMTIzNDU2NzgiLCJtb2NrY29uZmlnIjoiY29tcGxldGUifSwiZXhwIjoxODk3MzI3MTM0fQ.kmDlkya82PNyAaeSdN4J6Bn9cWBQUl6qmojuEkUnwLsqSkym0wsDunnRfGOWQeXjOIX44l9sTX31KV-Ee7pTdWARjGWQ9FgiHrew9zQUott60p0nhqAWWvnZgeiLWX8WAOFr7mE_mG0-zeRyeZSds442_WgAWj3PQIo7lU9G24mfW0Kyd_VIcIGixuEn0eUa6tb_cASeYtP_Z2AtSBiF3GcyBz28LTEhDffZpffmzvcKo8tlOCx5S7tP6Vogj4I58ZB3fPO__FJPeS8gZ4kYi8ZWnD77P0ErLeiBdnkxmXrmZNXaOyoM1L4lOU6xtd2G0u8HIVraJzDt69dst2MO4Q"
const kid = "CwJMvesKuEjvBOdZii37vFJl2Y30JXn5cVO34bzVaY4"

type MockOIDClient struct {
	mock.Mock
}

func (mock *MockOIDClient) GetJWK(kid string) (*JWK, error) {
	args := mock.Called(kid)
	return args.Get(0).(*JWK), args.Error(1)
}

func TestJWTValidation(t *testing.T) {
	// arrange
	mc := &MockOIDClient{}
	authTokenValidationIssuer := "https://account-oidcmock-mfm-general.ae.dev.cloudhh.de/auth/realms/Fielmann"
	authTokenValidationAudience := "mfm"
	jwtHandler := NewJwtHandler(mc, authTokenValidationIssuer, authTokenValidationAudience)
	mc.On("GetJWK", kid).Return(&JWK{
		Kty: "RSA",
		Use: "sig",
		Kid: kid,
		Alg: "RS256",
		N:   "pVym2SDO1yMeXzjowy7i2wvTJ6CBVvwsUEq5VsKjCI59tV87xCJ3s4z5p1fkdql4eB4lRO56BgY7fmaV6Vhhb9h57sy3UF7cx8EGAVdcHBjwJEHZQvjcquo4iH8S6GpJ_VZXtt_wAROudQWQoP0v9hBz4xjAOHSCMFinjNlgx5BiI75S9R0QdJuMKBhjpZuct-5oM40zYXFfNZs9l0MoJwdfojvS95xjm1kPyNSwSguKsGfcru7D5mFY15vaqBlXrGPxTTAys0Xd5MQYdVxC-fA5-n4VRs2CriiGcdrKdZj0d5XqqtclmnA7Cb71ViN1n3SjFIxH5PAOHucjdiuPvQ",
		E:   "AQAB",
	}, nil)
	// act
	err := jwtHandler.ValidateJWTToken(tokenString)

	// assert
	mc.AssertExpectations(t)
	assert.Nil(t, err)
}

func TestJWTValidation_whenAudienceMismatch_thenFail(t *testing.T) {
	// arrange
	a := assert.New(t)
	mc := &MockOIDClient{}
	authTokenValidationIssuer := "https://account-oidcmock-mfm-general.ae.dev.cloudhh.de/auth/realms/Fielmann"
	authTokenValidationAudience := "false-audience"
	jwtHandler := NewJwtHandler(mc, authTokenValidationIssuer, authTokenValidationAudience)
	mc.On("GetJWK", kid).Return(&JWK{
		Kty: "RSA",
		Use: "sig",
		Kid: kid,
		Alg: "RS256",
		N:   "pVym2SDO1yMeXzjowy7i2wvTJ6CBVvwsUEq5VsKjCI59tV87xCJ3s4z5p1fkdql4eB4lRO56BgY7fmaV6Vhhb9h57sy3UF7cx8EGAVdcHBjwJEHZQvjcquo4iH8S6GpJ_VZXtt_wAROudQWQoP0v9hBz4xjAOHSCMFinjNlgx5BiI75S9R0QdJuMKBhjpZuct-5oM40zYXFfNZs9l0MoJwdfojvS95xjm1kPyNSwSguKsGfcru7D5mFY15vaqBlXrGPxTTAys0Xd5MQYdVxC-fA5-n4VRs2CriiGcdrKdZj0d5XqqtclmnA7Cb71ViN1n3SjFIxH5PAOHucjdiuPvQ",
		E:   "AQAB",
	}, nil)

	// act
	err := jwtHandler.ValidateJWTToken(tokenString)

	// assert
	mc.AssertExpectations(t)
	a.Error(err)
	a.Contains(err.Error(), "unauthorized audience")
}

func TestJWTValidation_whenIssuerMismatch_thenFail(t *testing.T) {
	// arrange
	mc := &MockOIDClient{}
	authTokenValidationIssuer := "https://test-issuer.com/auth/realms/Fielmann"
	authTokenValidationAudience := "oidcmock"
	jwtHandler := NewJwtHandler(mc, authTokenValidationIssuer, authTokenValidationAudience)
	mc.On("GetJWK", kid).Return(&JWK{
		Kty: "RSA",
		Use: "sig",
		Kid: kid,
		Alg: "RS256",
		N:   "pVym2SDO1yMeXzjowy7i2wvTJ6CBVvwsUEq5VsKjCI59tV87xCJ3s4z5p1fkdql4eB4lRO56BgY7fmaV6Vhhb9h57sy3UF7cx8EGAVdcHBjwJEHZQvjcquo4iH8S6GpJ_VZXtt_wAROudQWQoP0v9hBz4xjAOHSCMFinjNlgx5BiI75S9R0QdJuMKBhjpZuct-5oM40zYXFfNZs9l0MoJwdfojvS95xjm1kPyNSwSguKsGfcru7D5mFY15vaqBlXrGPxTTAys0Xd5MQYdVxC-fA5-n4VRs2CriiGcdrKdZj0d5XqqtclmnA7Cb71ViN1n3SjFIxH5PAOHucjdiuPvQ",
		E:   "AQAB",
	}, nil)

	// act
	err := jwtHandler.ValidateJWTToken(tokenString)

	// assert
	mc.AssertExpectations(t)
	assert.EqualError(t, err, "unauthorized issuer")
}

func TestJWTValidation_whenTokenExpired_thenFail(t *testing.T) {
	// arrange
	a := assert.New(t)
	mc := &MockOIDClient{}
	authTokenValidationIssuer := "https://account-oidcmock-mfm-general.ae.dev.cloudhh.de/auth/realms/Fielmann"
	authTokenValidationAudience := "oidcmock"
	jwtHandler := NewJwtHandler(mc, authTokenValidationIssuer, authTokenValidationAudience)
	expiredTokenString := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6IkN3Sk12ZXNLdUVqdkJPZFppaTM3dkZKbDJZMzBKWG41Y1ZPMzRielZhWTQifQ.eyJuYmYiOjAsImlhdCI6MTU3ODQ2NjI5MywiaXNzIjoiaHR0cHM6Ly9hY2NvdW50LW9pZGNtb2NrLW1mbS1nZW5lcmFsLmFlLmRldi5jbG91ZGhoLmRlL2F1dGgvcmVhbG1zL0ZpZWxtYW5uIiwiYXVkIjpbIm9pZGNtb2NrIl0sInN1YiI6IjEyOGI2ZGZhLTgwNDMtNDZjZC1iMmUxLTg4MWVlNGJhZmE0YyIsInR5cCI6IkJlYXJlciIsImF6cCI6InNzby1vaWRjbW9jayIsImF1dGhfdGltZSI6MCwic2Vzc2lvbl9zdGF0ZSI6IjEyOGI2ZGZhLTgwNDMtNDZjZC1iMmUxLTg4MWVlNGJhZmE0YyIsImFjciI6IjEiLCJhbGxvd2VkLW9yaWdpbnMiOlsiaHR0cHM6Ly9hY2NvdW50LW9pZGNtb2NrLW1mbS1nZW5lcmFsLmFlLmRldi5jbG91ZGhoLmRlIl0sInJvbGVzIjpbXSwic2NvcGUiOiJvaWRjbW9jayIsIm1mbSI6eyJhY2NvdW50LWlkIjoiMTI4YjZkZmEtODA0My00NmNkLWIyZTEtODgxZWU0YmFmYTRjIn0sImV4cCI6MTU3ODQ2OTg5M30.FbhDA6_s76e6h06nrPQYsFcza4dUHlfUUm9aLMShWpjAidIBjifta-yNAaTIxqqYuacQma4eYiIKuiExViYfl9rZnN5D-6uumFuSC0twsHxLK6KbSHgj2s4Ru20oBb18w4LHSOelYCXPMLjweMkSNgl2PVnCWqjYSnY3WWjj1rSb5EcxvGuMxBYk6Txt7zkTbMLvn2u-8IBls4uqBqwcHH7UmLj3UG_GtGhvCwyF6YRUDTZaNifCSTlCEVmFVDGVZ0BSz9vmnj5R5qsr7ysMqluX8qX80QKmuT1RIiEByctE54LQsEHUx5l-cdAnvlh7gzYTX0S1glSI9WNLl8OVtQ"
	mc.On("GetJWK", kid).Return(&JWK{
		Kty: "RSA",
		Use: "sig",
		Kid: kid,
		Alg: "RS256",
		N:   "pVym2SDO1yMeXzjowy7i2wvTJ6CBVvwsUEq5VsKjCI59tV87xCJ3s4z5p1fkdql4eB4lRO56BgY7fmaV6Vhhb9h57sy3UF7cx8EGAVdcHBjwJEHZQvjcquo4iH8S6GpJ_VZXtt_wAROudQWQoP0v9hBz4xjAOHSCMFinjNlgx5BiI75S9R0QdJuMKBhjpZuct-5oM40zYXFfNZs9l0MoJwdfojvS95xjm1kPyNSwSguKsGfcru7D5mFY15vaqBlXrGPxTTAys0Xd5MQYdVxC-fA5-n4VRs2CriiGcdrKdZj0d5XqqtclmnA7Cb71ViN1n3SjFIxH5PAOHucjdiuPvQ",
		E:   "AQAB",
	}, nil)

	// act
	err := jwtHandler.ValidateJWTToken(expiredTokenString)

	// assert
	mc.AssertExpectations(t)
	a.Error(err)
	a.Contains(err.Error(), "Token is expired")
}

func TestJWTValidation_whenTokenIsNotValid_thenFail(t *testing.T) {
	// arrange
	a := assert.New(t)
	mc := &MockOIDClient{}
	authTokenValidationIssuer := "https://account-oidcmock-mfm-general.ae.dev.cloudhh.de/auth/realms/Fielmann"
	authTokenValidationAudience := "oidcmock"
	jwtHandler := NewJwtHandler(mc, authTokenValidationIssuer, authTokenValidationAudience)
	invalidTokenString := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6IkN3Sk12ZXNLdUVqdkJPZFppaTM3dkZKbDJZMzBKWG41Y1ZPMzRielZhWTQifQ.ewogICJqdGkiOiAiY2I1ZWIwMmYtZmU3Mi00NTJkLWE2NzItNzcwMTI2YTFkMDlmIiwKICAiZXhwIjogNDEwMjM1ODQwMCwKICAibmJmIjogMCwKICAiaWF0IjogNDEwMjM1ODQwMCwKICAiaXNzIjogImh0dHBzOi8vc3NvLWdlbmVyYWwtYXV0aC5hZS5xYS5jbG91ZGhoLmRlL2F1dGgvcmVhbG1zL0ZpZWxtYW5uIiwKICAiYXVkIjogWwogICAgInR2LWZyb250ZW5kIiwKICAgICJtZm0iLAogICAgInR2IgogIF0sCiAgInN1YiI6ICIwODBjMDkzYS00MzdhLTQ2MTItYjNmYS00ODIxMmUyNDdmNjUiLAogICJ0eXAiOiAiQmVhcmVyIiwKICAiYXpwIjogIm1mbS1hY2NvdW50LWZlIiwKICAibm9uY2UiOiAiMmRkNWM1Zjc1MDFhNDZjMDlkZDA4YTYyNmMwODdhYjYiLAogICJhdXRoX3RpbWUiOiAxNTc4NDY1OTY4LAogICJzZXNzaW9uX3N0YXRlIjogIjM1YTk4ZWM0LTlkZTQtNDA4OC05YmFjLTQ5OWI5ZmVhNmRjMCIsCiAgImFjciI6ICIxIiwKICAiYWxsb3dlZC1vcmlnaW5zIjogWwogICAgImh0dHBzOi8vZnJvbnRlbmQtbWZtLmFlLnFhLmNsb3VkaGguZGUiLAogICAgImh0dHBzOi8vcHdhLmFlLnFhLmNsb3VkaGguZGUiCiAgXSwKICAic2NvcGUiOiAib3BlbmlkIG1mbS1hY2NvdW50LXJlYWQgbWZtLXB1YmxpYyBtZm0tYWNjb3VudC13cml0ZSIsCiAgIm1mbSI6IHsKICAgICJhY2NvdW50LWlkIjogIjU0YjI3NDUxLTdiMTctNDRkZi04NTZjLTM4ZDY1MGUxOWMxYyIKICB9Cn0.FbhDA6_s76e6h06nrPQYsFcza4dUHlfUUm9aLMShWpjAidIBjifta-yNAaTIxqqYuacQma4eYiIKuiExViYfl9rZnN5D-6uumFuSC0twsHxLK6KbSHgj2s4Ru20oBb18w4LHSOelYCXPMLjweMkSNgl2PVnCWqjYSnY3WWjj1rSb5EcxvGuMxBYk6Txt7zkTbMLvn2u-8IBls4uqBqwcHH7UmLj3UG_GtGhvCwyF6YRUDTZaNifCSTlCEVmFVDGVZ0BSz9vmnj5R5qsr7ysMqluX8qX80QKmuT1RIiEByctE54LQsEHUx5l-cdAnvlh7gzYTX0S1glSI9WNLl8OVtQ"
	mc.On("GetJWK", kid).Return(&JWK{
		Kty: "RSA",
		Use: "sig",
		Kid: kid,
		Alg: "RS256",
		N:   "pVym2SDO1yMeXzjowy7i2wvTJ6CBVvwsUEq5VsKjCI59tV87xCJ3s4z5p1fkdql4eB4lRO56BgY7fmaV6Vhhb9h57sy3UF7cx8EGAVdcHBjwJEHZQvjcquo4iH8S6GpJ_VZXtt_wAROudQWQoP0v9hBz4xjAOHSCMFinjNlgx5BiI75S9R0QdJuMKBhjpZuct-5oM40zYXFfNZs9l0MoJwdfojvS95xjm1kPyNSwSguKsGfcru7D5mFY15vaqBlXrGPxTTAys0Xd5MQYdVxC-fA5-n4VRs2CriiGcdrKdZj0d5XqqtclmnA7Cb71ViN1n3SjFIxH5PAOHucjdiuPvQ",
		E:   "AQAB",
	}, nil)

	// act
	err := jwtHandler.ValidateJWTToken(invalidTokenString)

	// assert
	mc.AssertExpectations(t)
	a.Error(err)
	a.Contains(err.Error(), "crypto/rsa: verification error")
}

func TestJWTValidation_whenTokenIsManipulated_thenFail(t *testing.T) {
	// arrange
	a := assert.New(t)
	mc := &MockOIDClient{}
	authTokenValidationIssuer := "https://account-oidcmock-mfm-general.ae.dev.cloudhh.de/auth/realms/Fielmann"
	authTokenValidationAudience := "oidcmock"
	jwtHandler := NewJwtHandler(mc, authTokenValidationIssuer, authTokenValidationAudience)
	invalidTokenString := "manipulatedSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6IkN3Sk12ZXNLdUVqdkJPZFppaTM3dkZKbDJZMzBKWG41Y1ZPMzRielZhWTQifQ.eyJuYmYiOjAsImlhdCI6MTU3OTU5OTk1OCwiaXNzIjoiaHR0cHM6Ly9hY2NvdW50LW9pZGNtb2NrLW1mbS1nZW5lcmFsLmFlLmRldi5jbG91ZGhoLmRlL2F1dGgvcmVhbG1zL0ZpZWxtYW5uIiwiYXVkIjpbIm1mbSJdLCJzdWIiOiI5YzMxODJlNi02ZDE1LTRmZjItYTIzNC1lMmIwOTA2NzVmYWMiLCJ0eXAiOiJCZWFyZXIiLCJhenAiOiJtZm0tYWNjb3VudC1mZSIsImF1dGhfdGltZSI6MCwic2Vzc2lvbl9zdGF0ZSI6IjljMzE4MmU2LTZkMTUtNGZmMi1hMjM0LWUyYjA5MDY3NWZhYyIsImFjciI6IjEiLCJhbGxvd2VkLW9yaWdpbnMiOlsiKiJdLCJyb2xlcyI6W10sInNjb3BlIjoibWZtLXB1YmxpYyIsIm1mbSI6eyJhY2NvdW50LWlkIjoiMTIzNDU2NzgifSwiZXhwIjoxNTc5NjAzNTU4fQ.WmVJj5yb15-sg1Czdh3Vvggsb7BhZrn9ezkTl0N2ywemPTJzOA1BNwgcDcbUY5lIbABHA6IwTZHYcUJ4hm8P4SdPLgOWEbBGT-XJl99R-5hcOYlVNq528JGs_gr8US6t4p6Entueh3rlxtF6Au6a1ZFhEvUH6d5u6cP45Z1nh4G6V72OslCcioBKDBAZwFT495afIvmcRI3WV2oSZsL_AmKCEW-m_h4lQkuBbcGuU4XVHAMCWKhTHew1rdKUiN-sFjNpY-LR0TBG8DHSbQtIqBsMJ_-Lbc85aMp5lrDmf5-tUNtkQcj3TQ-sxhPUNZZ0W6OO4NfDJowe8xEzyvaSTg"

	// act
	err := jwtHandler.ValidateJWTToken(invalidTokenString)

	// assert
	mc.AssertNotCalled(t, "GetJWK", mock.Anything)
	a.Error(err)
	a.Contains(err.Error(), "invalid character")
}
