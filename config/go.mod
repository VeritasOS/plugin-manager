module github.com/VeritasOS/plugin-manager/v2/config

go 1.19

replace github.com/VeritasOS/plugin-manager/v2/utils/log => ../utils/log

require (
	github.com/VeritasOS/plugin-manager/v2/utils/log v1.0.0
	gopkg.in/yaml.v3 v3.0.1
)
