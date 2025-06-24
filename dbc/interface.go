package dbc

type UpgradeTable interface {
	OldTableName() string
	Upgrade(*DbClient) error
}
