package launchpad

// InformationType describes the visibility/confidentiality of a Launchpad
// resource. It is shared across many entry types (branches, bugs, recipes, etc.).
type InformationType string

const (
	InformationPublic          InformationType = "Public"
	InformationPublicSecurity  InformationType = "Public Security"
	InformationPrivateSecurity InformationType = "Private Security"
	InformationPrivate         InformationType = "Private"
	InformationProprietary     InformationType = "Proprietary"
	InformationEmbargoed       InformationType = "Embargoed"
)
