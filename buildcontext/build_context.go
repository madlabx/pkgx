package buildcontext

import (
	"strings"

	"github.com/madlabx/pkgx/utils"
)

type BuildInfo struct {
	Module    string
	Version   string
	Branch    string
	Commit    string
	BuildDate string
	Arch      string
}

const ConstUnknownValue = "(unknown)"

var (
	Version   = ConstUnknownValue
	Commit    = ConstUnknownValue
	BuildDate = ConstUnknownValue
	Module    = ConstUnknownValue
	Branch    = ConstUnknownValue
	Arch      = ConstUnknownValue
)

func Get() *BuildInfo {
	return &BuildInfo{
		Arch:      Arch,
		Module:    Module,
		Version:   Version,
		Commit:    Commit,
		Branch:    Branch,
		BuildDate: BuildDate,
	}
}

func GetArch() string {
	return Arch
}

func GetVersionBase() string {
	vb := strings.Split(Version, "-")
	return vb[0]
}

func (bi *BuildInfo) String() string {
	return utils.ToString(bi)
}
