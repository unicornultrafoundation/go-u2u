package ethapi

// PublicSfcAPI provides an API to access SFC related information.
type PublicSfcAPI struct {
	b Backend
}

func NewPublicSfcAPI(b Backend) *PublicSfcAPI {
	return &PublicSfcAPI{b}
}
