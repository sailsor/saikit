package factory

type extendFieldInfo struct {
	name string

	typeName string

	typePath string
}

func newLoggerFieldInfo() extendFieldInfo {
	efi := extendFieldInfo{
		name:     "logger",
		typeName: "Logger",
		typePath: "code.jshyjdtech.com/godev/hykit/log",
	}

	return efi
}

func newConfigFieldInfo() extendFieldInfo {
	efi := extendFieldInfo{
		name:     "conf",
		typeName: "Config",
		typePath: "code.jshyjdtech.com/godev/hykit/config",
	}

	return efi
}
