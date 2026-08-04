package enums

type CertificationUnknownPolicy string

const (
	CertificationUnknownAllow  CertificationUnknownPolicy = "allow"
	CertificationUnknownReject CertificationUnknownPolicy = "reject"
)

func (policy CertificationUnknownPolicy) Valid() bool {
	return policy == CertificationUnknownAllow || policy == CertificationUnknownReject
}
