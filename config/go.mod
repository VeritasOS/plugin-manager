module github.com/VeritasOS/plugin-manager/v2/config

go 1.20

replace github.com/VeritasOS/plugin-manager/v2/utils/log => ../utils/log

replace gopkg.in/yaml.v3 => ../vendor/gopkg.in/yaml.v3

require (
	github.com/VeritasOS/plugin-manager/v2/utils/log v1.0.0
	gopkg.in/yaml.v3 v3.0.1
)
