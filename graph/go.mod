module github.com/VeritasOS/plugin-manager/graph

go 1.19

replace config => ../config

replace os => ../utils/os

replace proto => ../proto


require github.com/VeritasOS/plugin-manager v1.0.0

require gopkg.in/yaml.v2 v2.4.0 // indirect
