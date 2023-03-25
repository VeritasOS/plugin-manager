module github.com/VeritasOS/plugin-manager/config

go 1.20

replace github.com/VeritasOS/plugin-manager/utils/log => ../utils/log

require (
	github.com/VeritasOS/plugin-manager/utils/log v1.0.5
	gopkg.in/yaml.v3 v3.0.1
)
