module github.com/VeritasOS/plugin-manager/v2

go 1.20

replace github.com/VeritasOS/plugin-manager/config/v2 => ./config

replace github.com/VeritasOS/plugin-manager/pluginmanager/v2 => ./pluginmanager

replace github.com/VeritasOS/plugin-manager/utils/log/v2 => ./utils/log

require (
	github.com/VeritasOS/plugin-manager/config/v2 v2.0.2
	github.com/VeritasOS/plugin-manager/pluginmanager/v2 v2.0.2
	github.com/VeritasOS/plugin-manager/utils/log/v2 v2.0.2
	github.com/abhijithda/pm-graph/v3 v3.0.3
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/VeritasOS/plugin-manager v1.0.3 // indirect
	github.com/fogleman/gg v1.3.0 // indirect
	github.com/goccy/go-graphviz v0.1.1 // indirect
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/image v0.6.0 // indirect
)
