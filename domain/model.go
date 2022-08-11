package domain

type Model struct {
	Id string

	Owner    Account
	Name     ModelName
	Desc     ProjDesc
	RepoType RepoType
	Protocol ProtocolName

	Tags []string

	RepoId string

	Version int
}
